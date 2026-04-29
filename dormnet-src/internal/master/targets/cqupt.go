package targets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/openwrt-dormnet/dormnet/internal/master/targets/enums/cqupt"
	"github.com/openwrt-dormnet/dormnet/internal/master/targets/registry"
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	"github.com/openwrt-dormnet/dormnet/shared/utils"
	"go.uber.org/zap"
)

type Cqupt struct {
	registry.AbsDormnetClient
}

type CquptExtraArgs struct {
	Operator   cqupt.CquptClientOperator `json:"operator" type:"ListValue" title:"Operator" desc:"Operator export" required:"true"`
	ClientType cqupt.CquptClientType     `json:"client_type" type:"ListValue" title:"Client type" desc:"Client type used during login" required:"true"`
}

func (c *Cqupt) DefaultExtraArgs() any {
	return &CquptExtraArgs{
		Operator:   "",
		ClientType: "",
	}
}

func (c *Cqupt) Constants() registry.DormnetClientConstants {
	return registry.DormnetClientConstants{
		DormId:   "cqupt",
		DormName: "Chongqing University of Posts and Telecommunications",
	}
}

type cquptEportalResponse struct {
	Result  string `json:"result"`
	Message string `json:"msg"`
}

func (c *Cqupt) Process(ifaces []*registry.DormnetClientBindIface) errx.Exception {
	accountId := c.AccountId()

	contexts := map[*utils.RequestContext]bool{}
	for _, iface := range ifaces {
		ctx, err := c.NewRequestContext(iface)
		if err != nil {
			return errx.NewExceptionWithCause(err, "failed to create request context")
		}
		status, err := ctx.IfaceController.Status()
		if err != nil {
			return errx.NewExceptionWithCause(err, "failed to query status of interface")
		}
		if !status.Up {
			log.Info("interface stopped, starting...")
			if err = ctx.IfaceController.WaitForUp(time.Second * 5); err != nil {
				return errx.NewExceptionWithCause(err, "failed to start iface",
					zap.String("iface", iface.Iface))
			}
		}
		if _, err = ctx.IfaceController.WaitForIp(time.Second * 5); err != nil {
			return errx.NewExceptionWithCause(err, "error occur during waiting for ip",
				zap.String("iface", iface.Iface))
		}
		if !status.Up {
			utils.DelayAndLog(time.Second * 15)
		}

		log.Info("checking interface whether available", zap.String("iface", iface.Iface))
		dormStatus, err := c.checkStatus(ctx)
		if dormStatus == cqupt.CquptLoginStatusLogin {
			contexts[ctx] = false
		} else if dormStatus == cqupt.CquptLoginStatusNotLogin {
			contexts[ctx] = true
		} else {
			log.Warn("error occur on the campus network, try restarting the interface...",
				zap.String("iface", iface.Iface), zap.String("status", string(dormStatus)),
				zap.Error(err))
			_ = ctx.IfaceController.WaitForDown(time.Second * 10)
			_ = ctx.IfaceController.WaitForUp(time.Second * 10)
			// 重启后 DHCP 续约通常要 10~20s，给宽裕一点
			if _, err := ctx.IfaceController.WaitForIp(time.Second * 30); err != nil {
				return errx.NewExceptionWithCause(err, "error occur during waiting for ip after restart",
					zap.String("iface", iface.Iface))
			}
			utils.DelayAndLog(time.Second * 15)
			nextDormStatus, err := c.checkStatus(ctx)
			if nextDormStatus == cqupt.CquptLoginStatusLogin {
				log.Info("network available after restart iface", zap.String("iface", iface.Iface))
				contexts[ctx] = false
			} else if nextDormStatus == cqupt.CquptLoginStatusNotLogin {
				log.Info("network available after restart iface, but need process", zap.String("iface", iface.Iface))
				contexts[ctx] = true
			} else if dormStatus == cqupt.CquptLoginStatusLoginButNotAvailable {
				log.Warn("error occur on the campus network")
				contexts[ctx] = false
			} else {
				return errx.NewException("could not connected to the campus network",
					zap.String("status", string(dormStatus)), zap.Error(err))
			}
		}
		utils.DelayAndLog(time.Second * 5)
	}
	ifacesNeedProcess := checkNeedProcess(contexts)
	if len(ifacesNeedProcess) <= 0 {
		log.Info("internet available, skip process", zap.String("account", accountId))
		return nil
	}
	log.Info("process needed", zap.Strings("ifaces", ifacesNeedProcess))

	log.Info("down all interface before batch login", zap.String("account", accountId))
	for ctx := range contexts {
		err := ctx.IfaceController.WaitForDown(time.Second * 5)
		if err != nil {
			log.Warn("failed to down interface", zap.String("iface", ctx.IfaceName), zap.Error(err))
		}
	}
	var err errx.Exception
	for ctx, needProcess := range contexts {
		if !needProcess {
			continue
		}
		log.Info("process login request", zap.String("account", accountId), zap.String("iface", ctx.IfaceName))
		log.Info("up interface before login", zap.String("account", accountId), zap.String("iface", ctx.IfaceName))
		err = ctx.IfaceController.WaitForUp(time.Second * 5)
		if err != nil {
			log.Error("failed to up interface", zap.String("iface", ctx.IfaceName), zap.Error(err))
			continue
		}
		utils.DelayAndLog(time.Second * 5)
		err = c.login(ctx)
		if err != nil {
			log.Error("error during login process", zap.String("iface", ctx.IfaceName), zap.Error(err))
		}
		log.Info("down interface after login", zap.String("account", accountId), zap.String("iface", ctx.IfaceName))
		err = ctx.IfaceController.WaitForDown(time.Second * 5)
		if err != nil {
			log.Error("failed to down interface", zap.String("iface", ctx.IfaceName), zap.Error(err))
		}
		utils.DelayAndLog(time.Second * 10)
	}
	log.Info("up all interface after batch login", zap.String("account", accountId))
	for ctx := range contexts {
		err := ctx.IfaceController.WaitForUp(time.Second * 10)
		if err != nil {
			log.Warn("failed to up interface", zap.String("iface", ctx.IfaceName), zap.Error(err))
		}
		<-time.After(time.Second * 15)
	}
	return err
}

func (c *Cqupt) checkStatus(ctx *utils.RequestContext) (cqupt.CquptLoginStatus, errx.Exception) {
	pinger := utils.NewPinger(ctx.IfaceName)
	if err := pinger.PingUntilSuccess("192.168.200.2", 3); err != nil {
		return cqupt.CquptLoginStatusPortalUnreachable, nil
	}
	if err := pinger.PingUntilSuccess("10.16.0.1", 3); err != nil {
		return cqupt.CquptLoginStatusGatewayUnreachable, nil
	}

	request := ctx.NewRequest()
	resp, err2 := request.Get("http://192.168.200.2/")
	if err2 != nil {
		return "", errx.NewExceptionWithError(err2, "failed to request self-service session")
	}

	bodyHtml, err := utils.Utf8StringRespBody(resp, "gbk")
	if err != nil {
		return "", errx.NewExceptionWithCause(err, "failed to convert body to utf8")
	}
	if strings.Contains(bodyHtml, "<title>上网登陆页</title>") {
		return cqupt.CquptLoginStatusNotLogin, nil
	}
	if resp.StatusCode() != 200 || !strings.Contains(bodyHtml, "<title>注销页</title>") {
		return "", errx.NewException("unknown login status")
	}

	if ok, err := c.CheckInternet(ctx); !ok {
		return cqupt.CquptLoginStatusLoginButNotAvailable, err
	}

	return cqupt.CquptLoginStatusLogin, nil
}

func (c *Cqupt) login(ctx *utils.RequestContext) errx.Exception {
	loginIface := c.LoginIface
	if loginIface == "" {
		loginIface = ctx.IfaceName
	}
	loginIfaceController := utils.NewIfaceController(loginIface)

	log.Info("wait for ip before login", zap.String("account", c.AccountId()),
		zap.String("iface", loginIface))
	ip, err := loginIfaceController.WaitForIp(time.Second * 5)
	if err != nil {
		return errx.NewExceptionWithCause(err, "failed to wait for ip", zap.String("iface", loginIface))
	}
	mac, err := loginIfaceController.QueryMac()
	if err != nil {
		return errx.NewExceptionWithCause(err, "failed to query mac", zap.String("iface", loginIface))
	}
	utils.DelayAndLog(time.Second * 10)

	extraArgs := ctx.ExtraArgs.(*CquptExtraArgs)

	request := ctx.NewRequest()
	request.SetQueryParam("c", "Portal")
	request.SetQueryParam("a", "page_type_data")
	request.SetHeader("User-Agent", extraArgs.ClientType.UserAgent())
	request.SetHeader("Referer", "http://192.168.200.2/")
	request.SetHeader("DNT", "1")
	resp, err2 := request.Get("http://192.168.200.2:801/eportal/")
	if err2 != nil {
		return errx.NewExceptionWithError(err2, "failed to login in step 1: failed to request")
	}
	var PHPSESSID *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "PHPSESSID" {
			PHPSESSID = c
			break
		}
	}
	if PHPSESSID == nil {
		return errx.NewException("failed to find cookie PHPSESSID")
	}

	<-time.After(time.Second * 10)

	request = ctx.NewRequest()
	request.SetQueryParam("c", "Portal")
	request.SetQueryParam("a", "login")
	request.SetQueryParam("callback", extraArgs.ClientType.Callback())
	request.SetQueryParam("login_method", "1")
	request.SetQueryParam("user_account",
		fmt.Sprintf(",%s,%s@%s",
			extraArgs.ClientType.AccountPrefix(),
			c.Username,
			extraArgs.Operator.String()))
	request.SetQueryParam("user_password", c.Password)
	request.SetQueryParam("wlan_user_ip", ip)
	request.SetQueryParam("wlan_user_ipv6", "")
	request.SetQueryParam("wlan_user_mac", strings.ReplaceAll(mac, ":", ""))
	request.SetQueryParam("wlan_ac_ip", "")
	request.SetQueryParam("wlan_ac_name", "")
	request.SetQueryParam("jsVersion", "3.3.3")
	request.SetCookie(PHPSESSID)
	request.SetHeader("User-Agent", extraArgs.ClientType.UserAgent())
	request.SetHeader("Referer", "http://192.168.200.2/")
	request.SetHeader("DNT", "1")
	resp, err2 = request.Get("http://192.168.200.2:801/eportal/")
	if err2 != nil {
		return errx.NewExceptionWithError(err2, "failed to request login")
	}

	respBody := resp.String()
	if len(respBody) <= 2 {
		return errx.NewException("empty response of login request")
	}
	respObj := &cquptEportalResponse{}
	err2 = json.Unmarshal([]byte(respBody[7:len(respBody)-1]), respObj)
	if err2 != nil {
		return errx.NewExceptionWithError(err2, "failed to parse response", zap.String("origin_resp", respBody))
	}
	if respObj.Result != "1" {
		return errx.NewException("server response an error", zap.String("message", respObj.Message))
	}

	return nil
}

func checkNeedProcess(ctxs map[*utils.RequestContext]bool) (ifaces []string) {
	ifaces = make([]string, 0)
	for ctx, val := range ctxs {
		if val {
			ifaces = append(ifaces, ctx.IfaceName)
		}
	}
	return
}

func init() {
	registry.Register(func(args *registry.AbsDormnetClient) registry.DormnetClient {
		return &Cqupt{*args}
	})
}

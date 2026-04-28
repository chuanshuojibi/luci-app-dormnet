package master

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openwrt-dormnet/dormnet/airkv"
	"github.com/openwrt-dormnet/dormnet/internal/master/daemon"
	"github.com/openwrt-dormnet/dormnet/internal/master/daemon/easytier"
	"github.com/openwrt-dormnet/dormnet/internal/master/daemon/standalone"
	"github.com/openwrt-dormnet/dormnet/internal/master/daemon/zerotier"
	_ "github.com/openwrt-dormnet/dormnet/internal/master/targets"
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	"github.com/openwrt-dormnet/dormnet/shared/utils"
	"github.com/openwrt-dormnet/dormnet/shared/utils/uci"
	"go.uber.org/zap"
)

func DormMasterDaemon() errx.Exception {
	err := utils.CheckExecutableExist("mwan3")
	if err != nil {
		return errx.NewExceptionWithCause(err, "environment check failed, refuse to start")
	}

	gin.DefaultWriter = log.NewSimpleLineWriter(log.RootLogger())

	g := utils.NewErrGroup()
	g.Go(func() errx.Exception {
		utils.DelayAndLog(time.Minute * 1)

		shell, err := utils.NewShellCli()
		if err != nil {
			return errx.NewExceptionWithCause(err, "failed to create shell command")
		}
		defer shell.Close()
		var workWith string
		if uci.GetBool("basic", "use_sd_network", false) {
			workWith = "standalone"
		} else {
			workWith = uci.GetString("basic", "work_with", "standalone")
		}

		return daemon.StartInternetCheck(func() ([]string, errx.Exception) {
			switch workWith {
			case "easytier":
				return easytier.CheckWhetherAllMembersOnline(shell)
			case "zerotier":
				return zerotier.CheckWhetherAllMembersOnline(shell)
			case "standalone":
				return standalone.CheckWhetherAllMembersOnline(shell)
			default:
				return nil, errx.NewException("unknown target to work with", zap.String("work-with", workWith))
			}
		})
	})
	workMode := uci.GetString("basic", "work_with", "master")
	workWith := uci.GetBool("basic", "use_sd_network", false)
	if workMode == "master" && workWith != true {
		listen := uci.GetString("basic", "listen", "127.0.0.1:10721")

		r := utils.RunWithGin(listen, g)
		g.Go(func() errx.Exception {
			return airkv.CreateAirKVDaemon(r)
		})
	}
	return g.Wait()
}

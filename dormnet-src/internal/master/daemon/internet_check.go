package daemon

import (
	"fmt"
	"runtime/debug"
	"slices"
	"time"

	"github.com/openwrt-dormnet/dormnet/internal/master/targets/registry"
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	"github.com/openwrt-dormnet/dormnet/shared/utils"
	"github.com/openwrt-dormnet/dormnet/shared/utils/uci"
	"go.uber.org/zap"
)

const internetPingTarget = "8.8.8.8"

func StartInternetCheck(allMemberOnline func() ([]string, errx.Exception)) errx.Exception {
	ResetAllStatesOnStartup()

	tickInterval := time.Duration(uci.GetInt("dormnet", "internet_check_interval", LightCheckEverySec)) * time.Second
	if tickInterval < 5*time.Second {
		tickInterval = LightCheckEverySec * time.Second
	}

	for {
		runOneCycle(allMemberOnline)
		<-time.After(tickInterval)
	}
}

// runOneCycle 每个 tick 跑一次：先列出当前所有 (account, iface) 对，对每对做轻量 ping；
// 通了就 Online、不动；不通才走 Process（受退避节流）。
func runOneCycle(allMemberOnline func() ([]string, errx.Exception)) {
	dormnetAccounts := uci.GetSections("account")
	if len(dormnetAccounts) == 0 {
		return
	}

	selected := uci.GetList("basic", "login_account")
	if len(selected) == 0 {
		if v := uci.GetString("basic", "login_account", ""); v != "" {
			selected = []string{v}
		}
	}
	if len(selected) > 0 {
		allowed := map[string]struct{}{}
		for _, a := range selected {
			if a != "" {
				allowed[a] = struct{}{}
			}
		}
		dormnetAccounts = slices.DeleteFunc(dormnetAccounts, func(account string) bool {
			_, ok := allowed[account]
			return !ok
		})
	}

	active := map[string]struct{}{}

	for accIdx, account := range dormnetAccounts {
		ifaces := bindIfacesOf(account)
		if len(ifaces) == 0 {
			active[keyOf(account, "")] = struct{}{}
			updateState(account, "", func(s *AccountState) {
				s.Status = StatusFailed
				s.LastError = "no bound interface"
			})
			continue
		}
		for ifaceIdx, ifaceName := range ifaces {
			active[keyOf(account, ifaceName)] = struct{}{}
			handleAccountIface(account, accIdx, ifaceIdx, ifaceName, allMemberOnline)
		}
	}

	resetMissingAccounts(active)
}

func bindIfacesOf(account string) []string {
	all := uci.GetSections("bind_iface")
	out := make([]string, 0, len(all))
	for _, sid := range all {
		pa, _ := uci.LookupString(sid, "parent_account")
		if pa != account {
			continue
		}
		ifn := uci.GetString(sid, "iface", "")
		if ifn != "" {
			out = append(out, ifn)
		}
	}
	return out
}

// handleAccountIface 实现状态机的核心：先 ping，通就停，不通且到点才 Process。
func handleAccountIface(account string, accIdx, ifaceIdx int, ifaceName string,
	allMemberOnline func() ([]string, errx.Exception)) {

	now := time.Now().Unix()

	stateMu.RLock()
	prev := states[keyOf(account, ifaceName)]
	var (
		prevStatus AccountStatus
		prevRetry  int
		prevNext   int64
	)
	if prev != nil {
		prevStatus = prev.Status
		prevRetry = prev.Retry
		prevNext = prev.NextAttemptAt
	}
	stateMu.RUnlock()

	if prevStatus == StatusGivenUp {
		return
	}

	updateState(account, ifaceName, func(s *AccountState) {
		s.Status = StatusChecking
	})
	pinger := utils.NewPinger(ifaceName)
	if err := pinger.PingUntilSuccess(internetPingTarget, 2); err == nil {
		updateState(account, ifaceName, func(s *AccountState) {
			s.Status = StatusOnline
			s.Retry = 0
			s.NextAttemptAt = 0
			s.LastError = ""
		})
		return
	}

	if prevNext > now {
		updateState(account, ifaceName, func(s *AccountState) {
			s.Status = StatusBackoff
		})
		return
	}

	updateState(account, ifaceName, func(s *AccountState) {
		s.Status = StatusLoggingIn
	})

	err := safeProcess(account, accIdx, ifaceIdx, ifaceName, allMemberOnline)

	if err == nil {
		updateState(account, ifaceName, func(s *AccountState) {
			s.Status = StatusOnline
			s.Retry = 0
			s.NextAttemptAt = 0
			s.LastError = ""
		})
		return
	}

	newRetry := prevRetry + 1
	backoffMin := 1
	for i := 1; i < newRetry; i++ {
		backoffMin *= 2
		if backoffMin >= MaxBackoffMin {
			backoffMin = MaxBackoffMin
			break
		}
	}
	nextAt := time.Now().Add(time.Duration(backoffMin) * time.Minute).Unix()

	if newRetry >= MaxRetry {
		updateState(account, ifaceName, func(s *AccountState) {
			s.Status = StatusGivenUp
			s.Retry = newRetry
			s.NextAttemptAt = 0
			s.LastError = err.Error()
		})
		log.Error("account given up after consecutive failures",
			zap.String("account", account), zap.String("iface", ifaceName),
			zap.Int("retry", newRetry), zap.Error(err))
		return
	}

	updateState(account, ifaceName, func(s *AccountState) {
		s.Status = StatusBackoff
		s.Retry = newRetry
		s.NextAttemptAt = nextAt
		s.LastError = err.Error()
	})
	log.Warn("login failed, backing off",
		zap.String("account", account), zap.String("iface", ifaceName),
		zap.Int("retry", newRetry), zap.Int("backoff_min", backoffMin),
		zap.Error(err))
}

func safeProcess(account string, accIdx, ifaceIdx int, ifaceName string,
	allMemberOnline func() ([]string, errx.Exception)) (errResult errx.Exception) {

	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Error("panic during Process; recovered",
				zap.String("account", account), zap.String("iface", ifaceName),
				zap.Any("panic", r), zap.String("stack", stack))
			errResult = errx.NewException(fmt.Sprintf("panic: %v", r))
		}
	}()

	return doActionOnAccount(allMemberOnline, accIdx, ifaceIdx, account, ifaceName)
}

// doActionOnAccount 仅处理单个 (account, iface)。
func doActionOnAccount(allMemberOnline func() ([]string, errx.Exception),
	accIdx, ifaceIdx int, account, only string) errx.Exception {

	rawBindIfaces := uci.GetSections("bind_iface")
	rawBindIfaces = slices.DeleteFunc(rawBindIfaces, func(s string) bool {
		pa, ok := uci.LookupString(s, "parent_account")
		if !ok || pa != account {
			return true
		}
		ifn := uci.GetString(s, "iface", "")
		return ifn != only
	})
	if len(rawBindIfaces) == 0 {
		return errx.NewException("no bound interface for account",
			zap.String("account", account), zap.String("iface", only))
	}

	username := uci.GetString(account, "username", "")
	password := uci.GetString(account, "password", "")
	loginIface := uci.GetString(account, "login_iface", "")
	if loginIface == "" {
		loginIface = only
	}
	if username == "" || password == "" {
		return errx.NewException("empty username or password")
	}
	typ := uci.GetString(account, "type", "")
	if typ == "" || !registry.HasClient(typ) {
		return errx.NewException("empty or unknown type")
	}
	client, err := registry.CreateClient(typ, account, username, password, loginIface)
	if err != nil {
		return errx.NewExceptionWithCause(err, "failed to create dormnet client")
	}

	offline, mErr := allMemberOnline()
	if mErr != nil {
		log.Warn("failed to check whether all members are online", zap.Error(mErr))
	}

	var bindIfaces []*registry.DormnetClientBindIface
	for index, rawBindIface := range rawBindIfaces {
		if len(offline) > 0 && (index > 0 || accIdx > 0 || ifaceIdx > 0) {
			break
		}
		var out registry.DormnetClientBindIface
		if err = registry.ParseExtraArgs(client.DefaultExtraArgs(), rawBindIface, &out); err != nil {
			return errx.NewExceptionWithCause(err, "failed to parse extra args for iface bind",
				zap.Int("iface_index", index),
				zap.String("account", client.AccountId()),
				zap.Error(err))
		}
		bindIfaces = append(bindIfaces, &out)
	}

	if len(bindIfaces) == 0 {
		return nil
	}
	log.Info("start process", zap.String("account", client.AccountId()),
		zap.String("iface", only))
	err = client.Process(bindIfaces)
	log.Info("process finished", zap.String("account", client.AccountId()),
		zap.String("iface", only))
	if err == nil {
		return nil
	}
	return errx.NewExceptionWithCause(err, "error occur during process",
		zap.String("account", client.AccountId()),
		zap.Error(err))
}

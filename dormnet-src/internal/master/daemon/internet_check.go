package daemon

import (
	"slices"
	"time"

	"github.com/openwrt-dormnet/dormnet/internal/master/targets/registry"
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	"github.com/openwrt-dormnet/dormnet/shared/utils/uci"
	"go.uber.org/zap"
)

func StartInternetCheck(allMemberOnline func() ([]string, errx.Exception)) errx.Exception {
	checkInterval := uci.GetInt("dormnet", "internet_check_interval", 120)
	for {
		dormnetAccounts := uci.GetSections("account")
		if len(dormnetAccounts) <= 0 {
			return errx.NewException("you must add at list one account")
		}
		selected := uci.GetList("basic", "login_account")
		// 兼容旧的单值 option：如果 list 为空，再读 string
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
			if len(dormnetAccounts) <= 0 {
				return errx.NewException("none of selected login accounts exist", zap.Strings("accounts", selected))
			}
		}

		offline, err := allMemberOnline()
		if err != nil {
			log.Warn("failed to check whether all members are online", zap.Error(err))
		}
		for accIdx, account := range dormnetAccounts {
			if err = doActionOnAccount(offline, accIdx, account); err != nil {
				log.Error("failed to deal with account", zap.String("account", account), zap.Error(err))
			}
		}
		<-time.After(time.Duration(checkInterval) * time.Second)
	}
}

func doActionOnAccount(offline []string, accIdx int, account string) errx.Exception {
	rawBindIfaces := uci.GetSections("bind_iface")
	rawBindIfaces = slices.DeleteFunc(rawBindIfaces, func(s string) bool {
		if value, ok := uci.LookupString(s, "parent_account"); ok {
			return value != account
		} else {
			return true
		}
	})
	if len(rawBindIfaces) <= 0 {
		return errx.NewException("a account must bind at list one iface")
	}
	username := uci.GetString(account, "username", "")
	password := uci.GetString(account, "password", "")
	loginIface := uci.GetString(account, "login_iface", "")
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

	var bindIfaces []*registry.DormnetClientBindIface
	for index, rawBindIface := range rawBindIfaces {
		if len(offline) > 0 && (index > 0 || accIdx > 0) {
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

	if len(bindIfaces) <= 0 {
		log.Info("no iface need to process, skip.")
		return nil
	}
	log.Info("start process", zap.String("account", client.AccountId()))
	err = client.Process(bindIfaces)
	log.Info("process finished", zap.String("account", client.AccountId()))
	if err == nil {
		return nil
	}
	return errx.NewExceptionWithCause(err, "error occur during process",
		zap.String("account", client.AccountId()),
		zap.Error(err))
}

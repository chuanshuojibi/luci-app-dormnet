package command

import (
	"github.com/openwrt-dormnet/dormnet/internal/master/targets/registry"
	"github.com/openwrt-dormnet/dormnet/shared/utils/uci"
)

func ExtraArgsOfTarget(target string) *StdJsonOutput {
	exist := registry.HasClient(target)
	if !exist {
		return &StdJsonOutput{
			Success: false,
			Message: "request target does not exist",
		}
	}

	client, err := registry.CreateClient(target, "", "", "", "")
	if err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: "failed to create empty target",
		}
	}

	args, err := registry.FormatExtraArgInfo(client.DefaultExtraArgs())
	if err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: err.Error(),
		}
	}
	return &StdJsonOutput{
		Success: true,
		Message: "",
		Data:    args,
	}
}

func ExtraArgsOfAccount(account string) *StdJsonOutput {
	target := uci.GetString(account, "type", "")
	if target == "" {
		return &StdJsonOutput{
			Success: false,
			Message: "account does not exist or the account type is not set",
		}
	}

	return ExtraArgsOfTarget(target)
}

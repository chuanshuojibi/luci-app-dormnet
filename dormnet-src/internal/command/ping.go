package command

import (
	"github.com/openwrt-dormnet/dormnet/shared/utils"
)

func Ping(iface string, target string) *StdJsonOutput {
	pinger := utils.NewPinger(iface)
	body, err := pinger.PingOnce(target)
	if err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: err.Error(),
		}
	} else {
		return &StdJsonOutput{
			Success: true,
			Message: "",
			Data:    body,
		}
	}
}

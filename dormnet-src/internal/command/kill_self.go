package command

import (
	"github.com/openwrt-dormnet/dormnet/internal/master"
)

func KillSelf() *StdJsonOutput {
	pid := master.DormnetPid()

	if err := pid.KillSelf(); err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: "failed to kill self",
		}
	}

	return &StdJsonOutput{
		Success: true,
		Message: "",
	}
}

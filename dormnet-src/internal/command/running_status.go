package command

import (
	"github.com/openwrt-dormnet/dormnet/internal/master"
	"github.com/openwrt-dormnet/dormnet/shared/utils/uci"
)

type RunningStatus struct {
	Running bool `json:"running"`
}

func GetRunningStatus() *StdJsonOutput {
	var status RunningStatus

	uci.InitUci("dormnet")
	pid := master.DormnetPid()
	isSelf, err := pid.IsSelf()
	if err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: "failed to load running status",
		}
	}
	status.Running = isSelf

	return &StdJsonOutput{
		Success: true,
		Message: "",
		Data:    status,
	}
}

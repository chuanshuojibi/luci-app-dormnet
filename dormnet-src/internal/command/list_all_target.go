package command

import (
	"github.com/openwrt-dormnet/dormnet/internal/master/targets/registry"
)

func ListTargets() *StdJsonOutput {
	return &StdJsonOutput{
		Success: true,
		Message: "",
		Data:    registry.ListAll(),
	}
}

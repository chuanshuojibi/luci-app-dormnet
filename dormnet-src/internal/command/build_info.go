package command

import (
	"github.com/openwrt-dormnet/dormnet/internal/constants"
)

type BuildInfos struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
}

func BuildInfo() *StdJsonOutput {
	return &StdJsonOutput{
		Success: true,
		Message: "",
		Data: BuildInfos{
			Version:   constants.Version(),
			BuildTime: constants.BuildTime(),
		},
	}
}

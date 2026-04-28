package command

import (
	"fmt"
	"strings"

	"github.com/openwrt-dormnet/dormnet/internal/master"
	"github.com/openwrt-dormnet/dormnet/shared/utils"
)

func Logs() *StdJsonOutput {
	pid := master.DormnetPid()

	if !pid.Exist() {
		return &StdJsonOutput{
			Success: true,
			Message: "",
			Data:    make([]string, 0),
		}
	}

	pidInt, err := pid.Pid()
	if err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: "failed to get pid",
		}
	}

	cli, err := utils.NewShellCli()
	if err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: "failed to open shell",
		}
	}

	_, content, err := cli.Run("logread", "|", "grep", fmt.Sprintf("\"dormnet\\[%d\\]\"", pidInt))
	if err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: "failed to read logs",
		}
	}

	var logs []string
	if content == "" {
		logs = make([]string, 0)
	} else {
		logs = strings.Split(content, "\n")
		if len(logs) > 100 {
			logs = logs[len(logs)-100:]
		}
	}

	return &StdJsonOutput{
		Success: true,
		Message: "",
		Data:    logs,
	}
}

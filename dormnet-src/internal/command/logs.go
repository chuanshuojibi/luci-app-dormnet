package command

import (
	"strings"

	"github.com/openwrt-dormnet/dormnet/shared/utils"
)

const maxLogLines = 500

func Logs() *StdJsonOutput {
	cli, err := utils.NewShellCli()
	if err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: "failed to open shell",
		}
	}
	defer cli.Close()

	// logread -e 能走 grep，但部分老版 logread 不支持，退一步用 pipe + grep。
	// 只按 syslog tag dormnet 过滤，不限 pid，这样重启后的老日志也能看到。
	_, content, err := cli.Run("logread", "|", "grep", "-F", "dormnet[")
	if err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: "failed to read logs",
		}
	}

	content = strings.TrimRight(content, "\n")
	logs := []string{}
	if content != "" {
		for _, line := range strings.Split(content, "\n") {
			if line == "" {
				continue
			}
			logs = append(logs, line)
		}
		if len(logs) > maxLogLines {
			logs = logs[len(logs)-maxLogLines:]
		}
	}

	return &StdJsonOutput{
		Success: true,
		Message: "",
		Data:    logs,
	}
}

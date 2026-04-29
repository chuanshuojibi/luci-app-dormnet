package command

import (
	"encoding/json"
	"os"

	"github.com/openwrt-dormnet/dormnet/internal/master/daemon"
)

// AccountStatus 读取 daemon 写在 /var/run/dormnet/status.json 的状态文件并返回。
// 如果文件不存在（dormnet 未运行）则返回空数组。
func AccountStatus() *StdJsonOutput {
	body, err := os.ReadFile(daemon.StatusFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &StdJsonOutput{
				Success: true,
				Message: "dormnet not running",
				Data:    []daemon.AccountState{},
			}
		}
		return &StdJsonOutput{
			Success: false,
			Message: "failed to read status file: " + err.Error(),
		}
	}
	var raw struct {
		States  []daemon.AccountState `json:"states"`
		Written int64                 `json:"written"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return &StdJsonOutput{
			Success: false,
			Message: "failed to parse status file: " + err.Error(),
		}
	}
	return &StdJsonOutput{
		Success: true,
		Message: "",
		Data: map[string]interface{}{
			"states":  raw.States,
			"written": raw.Written,
		},
	}
}

package master

import (
	"github.com/openwrt-dormnet/dormnet/shared/utils"
	"github.com/openwrt-dormnet/dormnet/shared/utils/uci"
)

var pidFile utils.PidFile

func DormnetPid() utils.PidFile {
	if pidFile == nil {
		pid := uci.GetString("basic", "pid_file", "/etc/dormnet/dormnet.pid")
		pidFile = utils.NewPidFile(pid)
	}
	return pidFile
}

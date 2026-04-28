package easytier

import (
	"slices"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/utils"
	"github.com/openwrt-dormnet/dormnet/shared/utils/uci"
)

type PeerItem struct {
	Hostname string `json:"hostname"`
}

func CheckWhetherAllMembersOnline(shell utils.ShellCli) ([]string, errx.Exception) {
	err := utils.CheckExecutableExist("easytier-cli")

	peers := make([]PeerItem, 0)
	_, err = shell.RunJson(&peers, "easytier-cli", "--output", "json", "peer")
	if err != nil {
		return nil, errx.NewExceptionWithCause(err, "failed to get peers form easytier-cli")
	}
	waitForMember := uci.GetList("dormnet", "peers")
	for _, peer := range peers {
		slices.DeleteFunc(waitForMember, func(s string) bool {
			return s == peer.Hostname
		})
	}
	return waitForMember, nil
}

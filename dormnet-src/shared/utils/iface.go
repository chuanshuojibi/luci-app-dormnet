package utils

import (
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	uci2 "github.com/openwrt-dormnet/dormnet/shared/utils/uci"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

type IfaceStatus struct {
	Up          bool                     `json:"up"`
	Available   bool                     `json:"available"`
	L3Device    string                   `json:"l3_device"`
	Device      string                   `json:"device"`
	Ipv4Address []IfaceStatusIpv4Address `json:"ipv4-address"`
}

type IfaceStatusIpv4Address struct {
	Address string `json:"address"`
	Mask    int    `json:"mask"`
}

type IfaceController interface {
	CreateDialer(timeout time.Duration) *net.Dialer

	Status() (*IfaceStatus, errx.Exception)

	QueryMac() (string, errx.Exception)
	QueryIp() (string, errx.Exception)
	WaitForIp(timeout time.Duration) (string, errx.Exception)

	Up() errx.Exception
	WaitForUp(timeout time.Duration) errx.Exception

	Down() errx.Exception
	WaitForDown(timeout time.Duration) errx.Exception
}

type ifaceController struct {
	name string
	uci  uci2.UciContext
}

func NewIfaceController(name string) IfaceController {
	return &ifaceController{
		name: name,
		uci:  uci2.NewUciConfig("network"),
	}
}

func (c *ifaceController) CreateDialer(timeout time.Duration) *net.Dialer {
	return &net.Dialer{
		Timeout: timeout,
		Control: func(network, address string, conn syscall.RawConn) error {
			var ctrlErr error
			err := conn.Control(func(fd uintptr) {
				ctrlErr = unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, c.name)
			})
			if err != nil {
				return err
			}
			return ctrlErr
		},
	}
}

func (c *ifaceController) QueryMac() (mac string, err errx.Exception) {
	deviceNames := []string{c.name}
	if status, statusErr := c.Status(); statusErr == nil {
		if status.L3Device != "" {
			deviceNames = append(deviceNames, status.L3Device)
		}
		if status.Device != "" {
			deviceNames = append(deviceNames, status.Device)
		}
	}

	for _, deviceName := range deviceNames {
		iface, err2 := net.InterfaceByName(deviceName)
		if err2 == nil && iface.HardwareAddr.String() != "" {
			return iface.HardwareAddr.String(), nil
		}
	}

	for i := 0; ; i++ {
		devName := c.uci.GetString(fmt.Sprintf("@device[%d]", i), "name", "")
		if devName == "" {
			err = errx.NewException("cannot find target iface", zap.String("iface", c.name))
			return
		}
		if devName != c.name {
			continue
		}
		mac = c.uci.GetString(fmt.Sprintf("@device[%d]", i), "macaddr", "")
		if mac == "" {
			err = errx.NewException("no macaddr set in iface", zap.String("iface", c.name))
		}
		return
	}
}
func (c *ifaceController) QueryIp() (string, errx.Exception) {
	status, err := c.Status()
	if err != nil {
		return "", errx.NewExceptionWithCause(err, "failed to query status", zap.String("iface", c.name))
	}
	if !status.Available || !status.Up {
		return "", errx.NewException("iface is not available of is not up", zap.String("iface", c.name))
	}
	if len(status.Ipv4Address) <= 0 {
		return "", errx.NewException("iface has no available ip", zap.String("iface", c.name))
	}
	return status.Ipv4Address[0].Address, nil
}

func (c *ifaceController) WaitForIp(timeout time.Duration) (string, errx.Exception) {
	start := time.Now()
	var ip string
	var err errx.Exception
	for time.Now().Sub(start) < timeout {
		<-time.After(time.Second / 2)
		ip, err = c.QueryIp()
		if err != nil {
			continue
		}
	}
	return ip, err
}

func (c *ifaceController) callUbus(command string) (int, string, errx.Exception) {
	return ShellRun("ubus", "call", fmt.Sprintf("network.interface.%s", c.name), command)
}

func (c *ifaceController) callUbusJson(output any, command string) (int, errx.Exception) {
	return ShellRunJson(output, "ubus", "call", fmt.Sprintf("network.interface.%s", c.name), command)
}

func (c *ifaceController) Down() errx.Exception {
	code, _, err := c.callUbus("down")
	if err != nil || code != 0 {
		return errx.NewExceptionWithCause(err, "error during command", zap.String("iface", c.name))
	}
	return nil
}

func (c *ifaceController) WaitForDown(timeout time.Duration) (err errx.Exception) {
	if err = c.Down(); err != nil {
		return
	}
	start := time.Now()
	var status *IfaceStatus
	for time.Now().Sub(start) < timeout {
		<-time.After(time.Second / 2)
		status, err = c.Status()
		if err != nil {
			log.Error("failed to query status", zap.String("iface", c.name), zap.Error(err))
			continue
		}
		if !status.Available || status.Up {
			err = errx.NewException("iface is not available or still up", zap.String("iface", c.name))
		}
	}
	return
}

func (c *ifaceController) Up() errx.Exception {
	code, _, err := c.callUbus("up")
	if err != nil || code != 0 {
		return errx.NewExceptionWithCause(err, "error during command", zap.String("iface", c.name))
	}
	return nil
}

func (c *ifaceController) WaitForUp(timeout time.Duration) (err errx.Exception) {
	if err = c.Up(); err != nil {
		return err
	}
	start := time.Now()
	var status *IfaceStatus
	for time.Now().Sub(start) < timeout {
		<-time.After(time.Second / 2)
		status, err = c.Status()
		if err != nil {
			log.Error("failed to query status", zap.String("iface", c.name), zap.Error(err))
			continue
		}
		if !status.Available || !status.Up {
			err = errx.NewException("iface is not available or still down", zap.String("iface", c.name))
		}
	}
	return
}

func (c *ifaceController) Status() (*IfaceStatus, errx.Exception) {
	var status IfaceStatus
	code, err := c.callUbusJson(&status, "status")
	if err != nil || code != 0 {
		return nil, errx.NewExceptionWithCause(err, "error during command", zap.String("iface", c.name))
	}
	return &status, nil
}

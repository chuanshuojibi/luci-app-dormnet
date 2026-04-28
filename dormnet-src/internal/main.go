package internal

import (
	"flag"

	"github.com/openwrt-dormnet/dormnet/internal/command"
	"github.com/openwrt-dormnet/dormnet/internal/master"
	"github.com/openwrt-dormnet/dormnet/internal/peer"
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	"github.com/openwrt-dormnet/dormnet/shared/utils"
	"github.com/openwrt-dormnet/dormnet/shared/utils/uci"
	"go.uber.org/zap"
)

func MainFunc() (code int, output *command.StdJsonOutput) {
	listTargets := flag.Bool("list-targets", false, "list all supported targets")

	extraArgs := flag.String("extra-args", "", "get extra args of target")

	extraArgsAccount := flag.String("extra-args-account", "", "get extra args of account")

	buildInfo := flag.Bool("build-info", false, "get build info of dormnet")

	killSelf := flag.Bool("kill-self", false, "kill self by configured pid file")

	status := flag.Bool("status", false, "get running status of dormnet")

	ping := flag.Bool("ping", false, "ping test on iface")
	iface := flag.String("iface", "", "ping on iface")
	target := flag.String("target", "", "ping to target")

	logs := flag.Bool("logs", false, "print syslog")

	daemon := flag.Bool("daemon", false, "daemon mode")

	flag.Parse()

	uci.InitUci("dormnet")
	pid := master.DormnetPid()

	if !*daemon {
		if *extraArgs != "" {
			output = command.ExtraArgsOfTarget(*extraArgs)
		} else if *extraArgsAccount != "" {
			output = command.ExtraArgsOfAccount(*extraArgsAccount)
		} else if *listTargets {
			output = command.ListTargets()
		} else if *killSelf {
			output = command.KillSelf()
		} else if *buildInfo {
			output = command.BuildInfo()
		} else if *status {
			output = command.GetRunningStatus()
		} else if *ping {
			output = command.Ping(*iface, *target)
		} else if *logs {
			output = command.Logs()
		} else {
			output = &command.StdJsonOutput{
				Success: false,
				Message: "no command provided",
			}
		}
		code = 0
		return
	}

	g := utils.NewErrQueue()
	defer g.DoDefer()

	g.Go(func() errx.Exception {
		return log.InitLogger()
	})
	g.Defer(func() {
		log.Sync()
	})

	enabled := uci.GetBool("basic", "enabled", false)

	g.Go(func() errx.Exception {
		if self, setupErr := pid.IsSelf(); setupErr == nil || !self {
			return nil
		}
		if enabled {
			log.Info("dormnet already running, stopping this one")
			code = 0
			return nil
		} else {
			log.Info("dormnet is disabled but running, stopping...")
			return pid.KillSelf()
		}
	})

	if enabled {
		g.Go(func() errx.Exception {
			return pid.Lock()
		})
		g.Defer(func() {
			_ = pid.Unlock()
		})
		g.Go(func() errx.Exception {
			return pid.Create()
		})
		g.Defer(func() {
			_ = pid.Remove()
		})

		g.Go(func() errx.Exception {
			return startDaemon()
		})
	}

	setupErr := g.Wait()

	if setupErr != nil {
		log.Error("dormnet stopped unexpectedly!", zap.Error(setupErr))
		code = 1
		return
	} else {
		log.Info("dormnet stopped")
		code = 0
		return
	}
}

func startDaemon() errx.Exception {
	workMode := uci.GetString("basic", "work_mode", "master")
	switch workMode {
	case "master":
		return master.DormMasterDaemon()
	case "peer":
		return peer.DormPeerDaemon()
	default:
		return errx.NewException("unknown work mode", zap.String("work_mode", workMode))
	}
}

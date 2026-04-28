package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"go.uber.org/zap"
)

type ShellCli interface {
	Run(command ...any) (int, string, errx.Exception)
	RunJson(output any, command ...any) (int, errx.Exception)
	Close() errx.Exception
}

type AbsShellCli struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	mutex  *sync.Mutex
}

func NewShellCli() (ShellCli, errx.Exception) {
	cmd := exec.Command("/bin/sh")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to open stdin pipe")
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to open stdout pipe")
	}
	err = cmd.Start()
	if err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to start shell")
	}

	return &AbsShellCli{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		mutex:  &sync.Mutex{},
	}, nil
}

func ShellRun(command ...any) (int, string, errx.Exception) {
	shell, err := NewShellCli()
	if err != nil {
		return 0, "", err
	}
	return shell.Run(command...)
}

func ShellRunJson(output any, command ...any) (int, errx.Exception) {
	shell, err := NewShellCli()
	if err != nil {
		return 0, err
	}
	return shell.RunJson(output, command...)
}

func (cli *AbsShellCli) Run(command ...any) (int, string, errx.Exception) {
	if len(command) <= 0 {
		return 0, "", errx.NewException("command is empty")
	}

	cli.mutex.Lock()
	defer cli.mutex.Unlock()

	endMark := "__CMD_DONE__"

	if _, err := fmt.Fprintln(cli.stdin, command...); err != nil {
		return 0, "", errx.NewExceptionWithError(err, "failed to execute command")
	}
	if _, err := fmt.Fprintln(cli.stdin, "echo", fmt.Sprintf("%s=$?", endMark)); err != nil {
		return 0, "", errx.NewExceptionWithError(err, "failed to get command return value")
	}

	scanner := bufio.NewScanner(cli.stdout)
	var output strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, endMark) {
			output.WriteString(line)
			output.WriteByte('\n')
			continue
		}
		code, err := strconv.Atoi(line[len(endMark)+1:])
		if err != nil {
			return 0, "", errx.NewExceptionWithError(err, "failed to parse command return value")
		}
		content := output.String()
		return code, content, nil
	}
	return 0, "", errx.IllegalStateError()
}

func (cli *AbsShellCli) RunJson(output any, command ...any) (int, errx.Exception) {
	code, content, err := cli.Run(command...)
	if err != nil {
		return code, err
	}
	if err := json.Unmarshal([]byte(content), output); err != nil {
		return code, errx.NewExceptionWithError(err, "failed to parse json output", zap.String("raw_content", content))
	}
	return code, nil
}

func (cli *AbsShellCli) Close() errx.Exception {
	_ = cli.stdin.Close()
	_ = cli.stdout.Close()
	_ = cli.cmd.Process.Kill()
	return nil
}

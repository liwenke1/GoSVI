package tool

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// bash
const ShellToUse = "bash"

// RunCommand 执行cmd命令
func RunCommand(name string, args []string, dir string) (bytes.Buffer, bytes.Buffer, error) {
	arg_list := strings.Join(args, " ")
	shell := strings.Join([]string{name, arg_list}, " ")
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", shell)
	cmd.Dir = dir
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		return stdoutBuf, stderrBuf, fmt.Errorf("cmd.Run() failed with %s", stderrBuf.String())
	}
	return stdoutBuf, stderrBuf, nil
}

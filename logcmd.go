package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// LogCmd enables Cmd with logging the executing command
type LogCmd struct {
	Cmd *exec.Cmd
}

// LogCommand constructs *LogCmd from *exec.Cmd
func LogCommand(cmd *exec.Cmd) *LogCmd {
	return &LogCmd{
		Cmd: cmd,
	}
}

// Run runs the underlying Cmd logging the CLI of the command
func (c *LogCmd) Run() error {
	args := []string{}

	for _, arg := range c.Cmd.Args {
		if strings.Contains(arg, " ") {
			args = append(args, "\""+arg+"\"")
		} else {
			args = append(args, arg)
		}
	}

	fmt.Printf("[CMD] %s\n", strings.Join(args, " "))
	return c.Cmd.Run()
}

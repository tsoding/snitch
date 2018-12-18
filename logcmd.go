package main

import (
	"os/exec"
	"strings"
	"fmt"
)

type LogCmd struct {
	Cmd *exec.Cmd
}

func LogCommand(cmd *exec.Cmd) *LogCmd {
	return &LogCmd {
		Cmd: cmd,
	}
}

func (c *LogCmd) Run() error {
	args := []string{}

	for _, arg := range c.Cmd.Args {
		if strings.Contains(arg, " ") {
			args = append(args, "\"" + arg + "\"")
		} else {
			args = append(args, arg)
		}
	}

	fmt.Printf("[CMD] %s\n", strings.Join(args, " "))
	return c.Cmd.Run()
}

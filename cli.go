package main

import (
	"errors"
	"flag"
	"fmt"
	"sort"
	"strings"
)

var commands []command

type command struct {
	name, desc string
	flags      *flag.FlagSet
	action     func() error
}

func addSubCommand(flags *flag.FlagSet, desc string, action func() error) {
	commands = append(commands, command{
		name:   flags.Name(),
		desc:   desc,
		flags:  flags,
		action: action,
	})
}

var (
	errNoCommandSpecified = errors.New("no command specified")
	errUknownCommand      = errors.New("unknown command")
	errCommandFailed      = errors.New("command failed")
)

func run(args []string) error {
	if len(args) <= 1 {
		printUsage()
		return errNoCommandSpecified
	}

	var cmd *command
	for _, v := range commands {
		if v.name == args[1] {
			cmd = &v
			break
		}
	}

	if cmd == nil {
		return fmt.Errorf("%s: %w", strings.Join(args, " "), errUknownCommand)
	}

	cmd.flags.Parse(args[2:])
	if err := cmd.action(); err != nil {
		return fmt.Errorf("%s: snitch %s: %w", err, cmd.name, errCommandFailed)
	}

	return nil
}

func printUsage() {
	sort.SliceStable(commands, func(i, j int) bool {
		return strings.Compare(commands[i].name, commands[j].name) == -1
	})

	fmt.Println("snitch [opt]")

	for _, cmd := range commands {
		fmt.Printf("\t%s", cmd.name)

		cmd.flags.VisitAll(func(f *flag.Flag) {
			name, _ := flag.UnquoteUsage(f)

			if len(name) == 0 {
				fmt.Printf(" [--%s]", f.Name)
			} else {
				fmt.Printf(" [--%s <%s>]", f.Name, name)
			}
		})

		fmt.Printf(": %s\n", cmd.desc)
	}
}

package main

import (
	"bufio"
	"fmt"
	"gopkg.in/go-ini/ini.v1"
	"os"
	"os/user"
	"path"
	"os/exec"
)

type GithubCredentials struct {
	PersonalToken string
}

func GithubCredentialsFromFile(filepath string) (GithubCredentials, error) {
	cfg, err := ini.Load(filepath)
	if err != nil {
		return GithubCredentials{}, err
	}

	return GithubCredentials {
		PersonalToken : cfg.Section("github").Key("personal_token").String(),
	}, nil
}

func ref_str(x string) *string {
	return &x
}

func ListSubcommand() error {
	return WalkTodosOfDir(".", func(todo Todo) error {
		fmt.Printf("%v\n", todo.LogString())
		return nil
	})
}

func ReportSubcommand(creds GithubCredentials, repo string) error {
	todosToReport := []Todo{}
	reader := bufio.NewReader(os.Stdin)

	err := WalkTodosOfDir(".", func(todo Todo) error {
		if todo.Id == nil {
			fmt.Printf("%v\n", todo.LogString());

			fmt.Printf("Do you want to report this? [y/n] ");
			text, err := reader.ReadString('\n')
			for err == nil && text != "y\n" && text != "n\n" {
				fmt.Printf("Do you want to report this? [y/n] ");
				text, err = reader.ReadString('\n')
			}

			if err != nil {
				return err
			}

			if text == "n\n" {
				return nil
			}

			todosToReport = append(todosToReport, todo)
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, todo := range todosToReport {
		reportedTodo, err := ReportTodo(todo, creds, repo)

		if err != nil {
			return err
		}

		fmt.Printf("[REPORTED] %v\n", reportedTodo.LogString())

		err = reportedTodo.UpdateInPlace()
		if err != nil {
			return err
		}
		
		err = exec.Command("git", "add", reportedTodo.Filename).Run()
		if err != nil {
			return err
		}

		err = exec.Command("git", "commit", "-m", reportedTodo.TodoString()).Run()
		if err != nil {
			return err
		}
	}

	return err
}

func usage() {
	// TODO(#9): implement a map for options instead of println'ing them all there
	fmt.Printf("snitch [opt]\n" +
		"\tlist: lists all todos of a dir recursively\n" +
		"\treport <owner/repo>: reports an issue to github\n")
}

func main() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	creds, err := GithubCredentialsFromFile(
		path.Join(usr.HomeDir, ".snitch/github.ini"))
	if err != nil {
		panic(err)
	}

	// TODO(#16): error results of subcommands are not handled
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			ListSubcommand()
		case "report":
			if len(os.Args) < 3 {
				usage()
				panic("Not enough arguments")
			}
			// TODO(#24): GitHub repo is not automatically derived from the git repo
			ReportSubcommand(creds, os.Args[2])
		default:
			panic(fmt.Sprintf("`%s` unknown command", os.Args[1]))
		}
	} else {
		usage()
	}
}

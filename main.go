package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
)

func listSubcommand() error {
	return WalkTodosOfDir(".", func(todo Todo) error {
		fmt.Printf("%v\n", todo.LogString())
		return nil
	})
}

func reportSubcommand(creds GithubCredentials, repo string, body string) error {
	todosToReport := []Todo{}
	reader := bufio.NewReader(os.Stdin)

	err := WalkTodosOfDir(".", func(todo Todo) error {
		if todo.ID == nil {
			fmt.Printf("%v\n", todo.LogString())

			fmt.Printf("Do you want to report this? [y/n] ")
			text, err := reader.ReadString('\n')
			for err == nil && text != "y\n" && text != "n\n" {
				fmt.Printf("Do you want to report this? [y/n] ")
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
		reportedTodo, err := ReportTodo(todo, creds, repo, body)

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

		err = exec.Command("git", "commit", "-m", reportedTodo.CommitMessage()).Run()
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
		"\treport <owner/repo> [issue-body]: reports all todos of a dir recursively as GitHub issues\n")
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

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			if err = listSubcommand(); err != nil {
				panic(err)
			}
		case "report":
			if len(os.Args) < 3 {
				usage()
				panic("Not enough arguments")
			}
			body := ""
			if len(os.Args) > 3 {
				body = os.Args[3]
			}
			// TODO(#24): GitHub repo is not automatically derived from the git repo
			if err = reportSubcommand(creds, os.Args[2], body); err != nil {
				panic(err)
			}
		default:
			panic(fmt.Sprintf("`%s` unknown command", os.Args[1]))
		}
	} else {
		usage()
	}
}

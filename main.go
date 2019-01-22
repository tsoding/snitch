package main

import (
	"bufio"
	"fmt"
	"gopkg.in/go-ini/ini.v1"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func yOrN(question string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [y/n] ", question)
	input, err := reader.ReadString('\n')
	text := strings.TrimSpace(input)

	for err == nil && text != "y" && text != "n" {
		fmt.Printf("%s [y/n] ", question)
		text, err = reader.ReadString('\n')
	}

	if err != nil || text == "n" {
		return false, err
	}

	return true, err
}

func listSubcommand(project Project) error {
	return project.WalkTodosOfDir(".", func(todo Todo) error {
		fmt.Printf("%v\n", todo.LogString())
		return nil
	})
}

func reportSubcommand(project Project, creds GithubCredentials, repo string, prependBody string) error {
	todosToReport := []Todo{}

	err := project.WalkTodosOfDir(".", func(todo Todo) error {
		if todo.ID == nil {
			fmt.Printf("%v\n", todo.LogString())
			fmt.Printf("Issue Title: %s\n", todo.Title)
			for _, bodyLine := range todo.Body {
				fmt.Printf("  %s\n", bodyLine)
			}

			yes, err := yOrN("Do you want to report this? ")

			if yes {
				todosToReport = append(todosToReport, todo)
			}

			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, todo := range todosToReport {
		reportedTodo, err := todo.ReportTodo(creds, repo, prependBody+"\n"+strings.Join(todo.Body, "\n"))

		if err != nil {
			return err
		}

		fmt.Printf("[REPORTED] %v\n", reportedTodo.LogString())

		err = reportedTodo.Update()
		if err != nil {
			return err
		}

		err = reportedTodo.GitCommit("Add")
		if err != nil {
			return err
		}
	}

	return err
}

func purgeSubcommand(project Project, creds GithubCredentials, repo string) error {
	todosToRemove := []Todo{}

	err := project.WalkTodosOfDir(".", func(todo Todo) error {
		if todo.ID == nil {
			return nil
		}

		status, err := todo.RetrieveGithubStatus(creds, repo)
		if err != nil {
			return err
		}

		if status == "closed" {
			fmt.Printf("[CLOSED] %v\n", todo.LogString())
			fmt.Printf("Issue link: https://github.com/%s/issues/%s\n", repo, (*todo.ID)[1:])

			yes, err := yOrN("This issue is closed. Do you want to remove the TODO?")

			if yes {
				todosToRemove = append(todosToRemove, todo)
			}

			if err != nil {
				return err
			}
		} else {
			fmt.Printf("[OPEN] %v\n", todo.LogString())
		}

		return err
	})

	sort.Slice(todosToRemove, func(i, j int) bool {
		if todosToRemove[i].Filename == todosToRemove[j].Filename {
			return todosToRemove[i].Line > todosToRemove[j].Line
		}

		return todosToRemove[i].Filename < todosToRemove[j].Filename
	})

	for _, todo := range todosToRemove {
		err = todo.Remove()
		if err != nil {
			return err
		}
		fmt.Printf("[REMOVED] %v\n", todo)

		err = todo.GitCommit("Remove")
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
		"\treport [--prepend-body <issue-body>]: reports all todos of a dir recursively as GitHub issues\n" +
		"\tpurge <owner/repo>: removes all of the reported TODOs that refer to closed issues\n")
}

func locateDotGit(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for absDir != "/" {
		dotGit := path.Join(absDir, ".git")

		if stat, err := os.Stat(dotGit); !os.IsNotExist(err) && stat.IsDir() {
			return dotGit, nil
		}

		absDir = filepath.Dir(absDir)
	}

	return "", fmt.Errorf("Couldn't find .git. Maybe you are not inside of a git repo")
}

func repoFromConfig(configPath string) (string, error) {
	cfg, err := ini.Load(configPath)
	if err != nil {
		return "", err
	}

	origin := cfg.Section("remote \"origin\"")
	if origin == nil {
		return "", fmt.Errorf("The git repo doesn't have any origin remote. " +
			"Please use `git remote add' command to add one.")
	}

	url := origin.Key("url")
	if url == nil {
		return "", fmt.Errorf("The origin remote doesn't have any URL's " +
			"associated with it.")
	}

	urlString := url.String()

	githubRepoRegexp := regexp.MustCompile(
		"github.com[:/]([-\\w]+)\\/([-\\w]+)(.git)?")
	groups := githubRepoRegexp.FindStringSubmatch(urlString)

	if groups != nil {
		return groups[1] + "/" + groups[2], nil
	}

	return "", fmt.Errorf("%s does not match %v",
		urlString, githubRepoRegexp)
}

func getGithubRepo(directory string) (string, error) {
	dotGit, err := locateDotGit(directory)
	if err != nil {
		return "", err
	}

	return repoFromConfig(path.Join(dotGit, "config"))
}

func parseParams(args []string) (map[string]string, error) {
	currentParam := ""
	result := map[string]string{}

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") { // Flag
			if len(currentParam) != 0 {
				result[currentParam] = ""
			}
			currentParam = arg[2:]
		} else { // Value
			if len(currentParam) == 0 {
				return nil, fmt.Errorf("Value %v is not associated with any flag", arg)
			}

			result[currentParam] = arg
			currentParam = ""
		}
	}

	if len(currentParam) != 0 {
		result[currentParam] = ""
	}

	return result, nil
}

func checkParams(params map[string]string, allowedParams []string) error {
	for param := range params {
		allowed := false
		for _, allowedParam := range allowedParams {
			if param == allowedParam {
				allowed = true
				break
			}
		}

		if !allowed {
			return fmt.Errorf("Unknown flag `%s'", param)
		}
	}

	return nil
}

func locateProject(directory string) (string, error) {
	dotGit, err := locateDotGit(directory)
	if err != nil {
		return "", err
	}

	return path.Join(filepath.Dir(dotGit), ".snitch.yaml"), nil
}

func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	creds, err := GithubCredentialsFromFile(
		path.Join(usr.HomeDir, ".snitch/github.ini"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	repo, err := getGithubRepo(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	projectPath, err := locateProject(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	project, err := NewProject(projectPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			if err = listSubcommand(*project); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		case "report":
			params, err := parseParams(os.Args[2:])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			err = checkParams(params, []string{"prepend-body"})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				usage()
				os.Exit(1)
			}

			prependBody, ok := params["prepend-body"]
			if !ok {
				prependBody = ""
			}

			fmt.Printf("Detected GitHub project: https://github.com/%s\n", repo)

			if err = reportSubcommand(*project, creds, repo, prependBody); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		case "purge":
			if err = purgeSubcommand(*project, creds, repo); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "`%s` unknown command\n", os.Args[1])
			os.Exit(1)
		}
	} else {
		usage()
	}
}

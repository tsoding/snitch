package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/go-ini/ini.v1"
)

func yOrN(question string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [y/n] ", question)
	input, err := reader.ReadString('\n')
	text := strings.TrimSpace(input)

	for err == nil && text != "y" && text != "n" {
		fmt.Printf("%s [y/n] ", question)
		text, err = reader.ReadString('\n')
		text = strings.TrimSpace(text)
	}

	if err != nil || text == "n" {
		return false, err
	}

	return true, err
}

func listSubcommand(project Project, filter func(todo Todo) bool) error {
	results, cancel, err := project.WalkTodosOfDir(".")
	if err != nil {
		return err
	}

	for v := range results {
		if v.err != nil {
			cancel()
			return v.err
		}
		if filter(*v.todo) {
			fmt.Println(v.todo.LogString())
		}
	}

	return nil
}

func reportSubcommand(project Project, creds IssueAPI, repo string, prependBody string) error {
	results, cancel, err := project.WalkTodosOfDir(".")
	if err != nil {
		return err
	}

	todosToReport := []*Todo{}
	for v := range results {
		if v.err != nil {
			cancel()
			return v.err
		}
		if v.todo.ID != nil {
			continue
		}

		fmt.Printf("%v\n", v.todo.LogString())
		fmt.Printf("Issue Title: %s\n", v.todo.Title)
		for _, bodyLine := range v.todo.Body {
			fmt.Printf("  %s\n", bodyLine)
		}

		yes, err := yOrN("Do you want to report this? ")
		if err != nil {
			cancel()
			return err
		} else if yes {
			todosToReport = append(todosToReport, v.todo)
		}
	}

	for _, todo := range todosToReport {
		reportedTodo, err := todo.Report(creds, repo,
			prependBody+"\n\n"+strings.Join(todo.Body, "\n\n"))

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

func purgeSubcommand(project Project, creds IssueAPI, repo string) error {
	results, cancel, err := project.WalkTodosOfDir(".")
	if err != nil {
		return err
	}

	todosToRemove := []*Todo{}
	for v := range results {
		if v.err != nil {
			cancel()
			return v.err
		}
		if v.todo.ID == nil {
			continue
		}

		status, err := v.todo.RetrieveStatus(creds, repo)
		if err != nil {
			cancel()
			return err
		}
		if status != "closed" {
			fmt.Printf("[OPEN] %v\n", v.todo.LogString())
			continue
		}

		fmt.Printf("[CLOSED] %v\n", v.todo.LogString())
		fmt.Printf("Issue link: https://%s/%s/issues/%s\n",
			creds.getHost(), repo, (*v.todo.ID)[1:])

		yes, err := yOrN("This issue is closed. Do you want to remove the TODO?")
		if err != nil {
			cancel()
			return err
		} else if yes {
			todosToRemove = append(todosToRemove, v.todo)
		}
	}

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

func locateDotGit(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	rooted := ""
	if err != nil {
		return "", err
	}

	for absDir != rooted {
		dotGit := path.Join(absDir, ".git")

		if stat, err := os.Stat(dotGit); !os.IsNotExist(err) && stat.IsDir() {
			return dotGit, nil
		}
		rooted = absDir
		absDir = filepath.Dir(absDir)
	}

	return "", fmt.Errorf("Couldn't find .git. Maybe you are not inside of a git repo")
}

func getURLAliases() (map[string]string, error) {
	usr, err := user.Current()
	if err != nil {
		return map[string]string{}, nil
	}

	path := path.Join(usr.HomeDir, ".gitconfig")

	cfg, err := ini.Load(path)
	if err != nil {
		return map[string]string{}, nil
	}

	sections := cfg.Sections()
	aliases := map[string]string{}

	for _, elem := range sections {
		sectionName := elem.Name()

		regex := regexp.MustCompile("url \"([-\\w]+@[github.com|gitlab.com][^\"]+)\"")
		urlSections := regex.FindAllStringSubmatch(sectionName, -1)

		for _, elem := range urlSections {
			section := cfg.Section(elem[0])
			alias, err := section.GetKey("insteadOf")
			if err != nil {
				return map[string]string{}, nil
			}

			aliases[alias.Value()] = elem[1]
		}
	}

	return aliases, nil
}

func getRepo(directory string) (string, IssueAPI, error) {
	credentials := getCredentials()
	if len(credentials) == 0 {
		return "", nil, fmt.Errorf("No credentials have been found. Read https://github.com/tsoding/snitch#credentials")
	}

	dotGit, err := locateDotGit(directory)
	if err != nil {
		return "", nil, err
	}

	configPath := path.Join(dotGit, "config")

	cfg, err := ini.Load(configPath)
	if err != nil {
		return "", nil, err
	}

	origin := cfg.Section("remote \"origin\"")
	if origin == nil {
		return "", nil, fmt.Errorf("The git repo doesn't have any origin remote. " +
			"Please use `git remote add' command to add one.")
	}

	url := origin.Key("url")
	if url == nil {
		return "", nil, fmt.Errorf("The origin remote doesn't have any URL's " +
			"associated with it.")
	}

	aliases, err := getURLAliases()
	if err != nil {
		return "", nil, err
	}

	urlString := url.String()

	for key, value := range aliases {
		if strings.Contains(urlString, key) {
			urlString = strings.Replace(urlString, key, value, 1)
			break
		}
	}

	for _, creds := range credentials {
		hostRegex := regexp.MustCompile(
			creds.getHost() + "[:/]([-\\w]+)\\/([-\\w]+)(.git)?")
		groups := hostRegex.FindStringSubmatch(urlString)

		if groups != nil {
			return groups[1] + "/" + groups[2], creds, nil
		}
	}

	return "", nil, fmt.Errorf("%s does not match any of the hosts", urlString)
}

func locateProject(directory string) (string, error) {
	dotGit, err := locateDotGit(directory)
	if err != nil {
		return "", err
	}

	return filepath.Dir(dotGit), nil
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getCredentials() []IssueAPI {
	creds := []IssueAPI{}

	if github, err := getGithubCredentials(); err == nil {
		creds = append(creds, github)
	}
	creds = getGitlabCredentials(creds)

	return creds
}

func main() {
	projectPath, err := locateProject(".")
	exitOnError(err)

	project, err := NewProject(projectPath)
	exitOnError(err)

	var (
		listCmd           = flag.NewFlagSet("list", flag.ExitOnError)
		listCmdReported   = listCmd.Bool("reported", false, "list reported todos")
		listCmdUnreported = listCmd.Bool("unreported", false, "list unreported todos")
	)
	addSubCommand(listCmd, "lists all todos of a dir recursively", func() error {
		return listSubcommand(*project, func(todo Todo) bool {
			return *listCmdReported == *listCmdUnreported ||
				(*listCmdReported && todo.ID != nil) ||
				(*listCmdUnreported && todo.ID == nil)
		})
	})

	var (
		reportCmd            = flag.NewFlagSet("report", flag.ExitOnError)
		reportCmdPrependBody = reportCmd.String("prepend-body", "", "prepend the `issue-body`")
	)
	addSubCommand(reportCmd, "reports all todos of a dir recursively as GitHub issues", func() error {
		repo, creds, err := getRepo(".")
		if err != nil {
			return err
		}

		fmt.Printf("Detected project: https://%s/%s\n", creds.getHost(), repo)
		return reportSubcommand(*project, creds, repo, *reportCmdPrependBody)
	})

	var (
		purgeCmd = flag.NewFlagSet("purge", flag.ExitOnError)
	)
	addSubCommand(purgeCmd, "removes all of the reported TODOs that refer to closed issues", func() error {
		repo, creds, err := getRepo(".")
		if err != nil {
			return err
		}

		return purgeSubcommand(*project, creds, repo)
	})

	if err = run(os.Args); err != nil && !errors.Is(err, errNoCommandSpecified) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

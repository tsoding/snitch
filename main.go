package main

import (
	"bufio"
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

func reportSubcommand(project Project, creds IssueAPI, repo string, prependBody string, verbosity string) error {
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

		yes := false
		if verbosity == "interactive" {
			yes, err = yOrN("Do you want to report this? ")
		} else if verbosity == "non-interactive" {
			yes = true
		}

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

func purgeSubcommand(project Project, creds IssueAPI, repo string, verbosity string) error {
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

		yes := false
		if verbosity == "interactive" {
			yes, err = yOrN("This issue is closed. Do you want to remove the TODO?")
		} else if verbosity == "non-interactive" {
			yes = true
		}
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

func usage() {
	// FIXME(#9): implement a map for options instead of println'ing them all there
	fmt.Printf("snitch [opt]\n" +
		"\tlist [--unreported] [--reported] [--y]: lists all todos of a dir recursively\n" +
		"\treport [--prepend-body <issue-body>] [--y]: reports all todos of a dir recursively \n\t\tas GitHub issues\n" +
		"\tpurge: removes all of the reported TODOs that refer to closed issues\n")
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

func parseParams(args []string) (map[string]string, error) {
	currentParam := ""
	result := map[string]string{}

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {       // Long Flag
			if len(currentParam) != 0 {
				result[currentParam] = ""
			}
			currentParam = arg[2:]
		} else if strings.HasPrefix(arg, "-") { // Short Flags
			if len(currentParam) != 0 {
				result[currentParam] = ""
			}
			currentParam = arg[1:]
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
	creds = getGiteaCredentials(creds)
	return creds
}

func main() {
	projectPath, err := locateProject(".")
	exitOnError(err)

	project, err := NewProject(projectPath)
	exitOnError(err)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			params, err := parseParams(os.Args[2:])
			exitOnError(err)

			err = checkParams(params, []string{"unreported", "reported"})
			exitOnError(err)
			_, unreported := params["unreported"]
			_, reported := params["reported"]

			err = listSubcommand(*project, func(todo Todo) bool {
				filter := reported == unreported

				if unreported {
					filter = filter || todo.ID == nil
				}

				if reported {
					filter = filter || todo.ID != nil
				}

				return filter
			})
			exitOnError(err)
		case "report":
			params, err := parseParams(os.Args[2:])
			exitOnError(err)

			err = checkParams(params, []string{"prepend-body", "y"})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				usage()
				os.Exit(1)
			}

			prependBody, ok := params["prepend-body"]
			if !ok {
				prependBody = ""
			}

			verbosity, ok := params["y"]
			if !ok {
				verbosity = "interactive"
			} else {
				verbosity = "non-interactive"
			}

			repo, creds, err := getRepo(".")
			exitOnError(err)

			fmt.Printf("Detected project: https://%s/%s\n", creds.getHost(), repo)

			if err = reportSubcommand(*project, creds, repo, prependBody, verbosity); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		case "purge":
			repo, creds, err := getRepo(".")
			exitOnError(err)

			params, err := parseParams(os.Args[2:])
			exitOnError(err)

			err = checkParams(params, []string{"y"})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				usage()
				os.Exit(1)
			}

			verbosity, ok := params["y"]
			if !ok {
				verbosity = "interactive"
			} else {
				verbosity = "non-interactive"
			}

			if err = purgeSubcommand(*project, creds, repo, verbosity); err != nil {
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

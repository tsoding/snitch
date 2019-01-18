package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Todo contains information about a TODO in the repo
type Todo struct {
	Prefix   string
	Suffix   string
	Keyword  string
	ID       *string
	Filename string
	Line     int
	Title    string
	Body     []string
}

// LogString formats TODO for compilation logging. Format is
// compatible with Emacs compilation mode, so you can easily jump
// between the todos.
func (todo Todo) LogString() string {
	if todo.ID == nil {
		return fmt.Sprintf("%s:%d: %s%s: %s",
			todo.Filename, todo.Line,
			todo.Prefix, todo.Keyword,
			todo.Suffix)
	}

	return fmt.Sprintf("%s:%d: %s%s(%s): %s",
		todo.Filename, todo.Line,
		todo.Prefix, todo.Keyword,
		*todo.ID, todo.Suffix)
}

func (todo Todo) String() string {
	if todo.ID == nil {
		return fmt.Sprintf("%s%s: %s",
			todo.Prefix, todo.Keyword, todo.Suffix)
	}

	return fmt.Sprintf("%s%s(%s): %s",
		todo.Prefix, todo.Keyword, *todo.ID,
		todo.Suffix)
}

// ParseBodyLine strips off the prefix of a body line of the TODO
func (todo Todo) ParseBodyLine(line string) *string {
	if strings.HasPrefix(line, todo.Prefix) {
		bodyLine := strings.TrimPrefix(line, todo.Prefix)
		return &bodyLine
	}

	return nil
}

func (todo Todo) updateToFile(outputFilename string, lineCallback func(int, string) (string, bool)) error {
	inputFile, err := os.Open(todo.Filename)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer func() {
		cerr := outputFile.Close()
		if cerr != nil {
			err = cerr
		}
	}()

	scanner := bufio.NewScanner(inputFile)
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()

		replace, remove := lineCallback(lineNumber, line)
		if !remove {
			fmt.Fprintln(outputFile, replace)
		}

		lineNumber = lineNumber + 1
	}

	return err
}

func (todo Todo) updateInPlace(lineCallback func(int, string) (string, bool)) error {
	outputFilename := todo.Filename + ".snitch"
	err := todo.updateToFile(outputFilename, lineCallback)
	if err != nil {
		return err
	}

	err = os.Rename(outputFilename, todo.Filename)

	if err != nil {
		return err
	}

	return err
}

// Update updates the file where the Todo is located in-place.
func (todo Todo) Update() error {
	return todo.updateInPlace(func(lineNumber int, line string) (string, bool) {
		if lineNumber == todo.Line {
			return todo.String(), false
		}

		return line, false
	})
}

// Remove removes the Todo from the file where it is located in-place.
func (todo Todo) Remove() error {
	// TODO(#124): Todo.Remove does not remove the body of the TODO
	// It should remove both:
	// - The TODO itself
	// - The body of the TODO
	return todo.updateInPlace(func(lineNumber int, line string) (string, bool) {
		if lineNumber == todo.Line {
			return "", true
		}

		return line, false
	})
}

// GitCommit commits the Todo location to the git repo
func (todo Todo) GitCommit(prefix string) error {
	// TODO(#96): there is no way to check that Todo is unreported at compile time
	if todo.ID == nil {
		panic(fmt.Sprintf("Trying to commit an unreported TODO! %v", todo))
	}

	if err := LogCommand(exec.Command("git", "add", todo.Filename)).Run(); err != nil {
		return err
	}

	if err := LogCommand(exec.Command("git", "commit", "-m", fmt.Sprintf("%s %s(%s)", prefix, todo.Keyword, *todo.ID))).Run(); err != nil {
		return err
	}

	return nil
}

// RetrieveGithubStatus retrieves the current status of TODOs issue
// from GitHub
func (todo Todo) RetrieveGithubStatus(creds GithubCredentials, repo string) (string, error) {
	json, err := creds.QueryGithubAPI(
		"GET",
		// TODO(#59): possible GitHub API injection attack
		"https://api.github.com/repos/"+repo+"/issues/"+(*todo.ID)[1:],
		nil)

	if err != nil {
		return "", err
	}

	return json["state"].(string), nil
}

// ReportTodo reports the todo as a Github Issue, updates the file
// where the todo is located and commits the changes to the git repo.
func (todo Todo) ReportTodo(creds GithubCredentials, repo string, body string) (Todo, error) {
	// TODO(#60): ReportTodo is not a Todo method
	json, err := creds.QueryGithubAPI(
		"POST",
		"https://api.github.com/repos/"+repo+"/issues",
		map[string]interface{}{
			"title": todo.Title,
			"body":  body,
		})
	if err != nil {
		return todo, err
	}

	id := "#" + strconv.Itoa(int(json["number"].(float64)))
	todo.ID = &id

	return todo, err
}

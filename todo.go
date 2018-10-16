package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Todo contains information about a TODO in the repo
type Todo struct {
	Prefix   string
	Suffix   string
	ID       *string
	Filename string
	Line     int
}

// CommitMessage formats a commit message for commiting the TODO
func (todo Todo) CommitMessage() string {
	if todo.ID == nil {
		return fmt.Sprintf("TODO")
	}

	return fmt.Sprintf("TODO(%s)", *todo.ID)
}

// LogString formats TODO for compilation logging. Format is
// compatible with Emacs compilation mode, so you can easily jump
// between the todos.
func (todo Todo) LogString() string {
	if todo.ID == nil {
		return fmt.Sprintf("%s:%d: %sTODO: %s",
			todo.Filename, todo.Line,
			todo.Prefix, todo.Suffix)
	}

	return fmt.Sprintf("%s:%d: %sTODO(%s): %s",
		todo.Filename, todo.Line,
		todo.Prefix, *todo.ID, todo.Suffix)
}

func (todo Todo) String() string {
	if todo.ID == nil {
		return fmt.Sprintf("%sTODO: %s",
			todo.Prefix, todo.Suffix)
	}

	return fmt.Sprintf("%sTODO(%s): %s",
		todo.Prefix, *todo.ID, todo.Suffix)
}

// UpdateToFile updates the file where the Todo is located without
// changing the original file. The newer version of the file is dumped
// to '<filepath>.snitch' in the same directory.
func (todo Todo) UpdateToFile(outputFilename string) error {
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
	line := 1

	for scanner.Scan() {
		text := scanner.Text()

		if todo.Line == line {
			fmt.Fprintln(outputFile, todo)
		} else {
			fmt.Fprintln(outputFile, text)
		}

		line = line + 1
	}

	return err
}

// UpdateInPlace updates the file where the Todo is located in-place.
func (todo Todo) UpdateInPlace() error {
	outputFilename := todo.Filename + ".snitch"
	err := todo.UpdateToFile(outputFilename)
	if err != nil {
		return err
	}

	err = os.Rename(outputFilename, todo.Filename)

	if err != nil {
		return err
	}

	return err
}

func (todo Todo) removeToFile(outputFilename string) error {
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
	line := 1

	for scanner.Scan() {
		text := scanner.Text()

		if todo.Line != line {
			fmt.Fprintln(outputFile, text)
		}

		line = line + 1
	}

	return err
}

func (todo Todo) Remove() error {
	// TODO: duplicate code in Todo.Remove() and Todo.Update()
	outputFilename := todo.Filename + ".snitch"
	err := todo.removeToFile(outputFilename)
	if err != nil {
		return err
	}

	err = os.Rename(outputFilename, todo.Filename)

	if err != nil {
		return err
	}

	return err
}

func lineAsUnreportedTodo(line string) *Todo {
	unreportedTodo := regexp.MustCompile("^(.*)TODO: (.*)$")
	groups := unreportedTodo.FindStringSubmatch(line)

	if groups != nil {
		return &Todo{
			Prefix:   groups[1],
			Suffix:   groups[2],
			ID:       nil,
			Filename: "",
			Line:     0,
		}
	}

	return nil
}

func lineAsReportedTodo(line string) *Todo {
	unreportedTodo := regexp.MustCompile("^(.*)TODO\\((.*)\\): (.*)$")
	groups := unreportedTodo.FindStringSubmatch(line)

	if groups != nil {
		return &Todo{
			Prefix:   groups[1],
			Suffix:   groups[3],
			ID:       &groups[2],
			Filename: "",
			Line:     0,
		}
	}

	return nil
}

// LineAsTodo constructs a Todo from a string
func LineAsTodo(line string) *Todo {
	if todo := lineAsUnreportedTodo(line); todo != nil {
		return todo
	}

	if todo := lineAsReportedTodo(line); todo != nil {
		return todo
	}

	return nil
}

// WalkTodosOfFile visits all of the TODOs in a particular file
func WalkTodosOfFile(path string, visit func(Todo) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	line := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		todo := LineAsTodo(scanner.Text())

		if todo != nil {
			todo.Filename = path
			todo.Line = line

			if err := visit(*todo); err != nil {
				return err
			}
		}

		line = line + 1
	}

	return scanner.Err()
}

// WalkTodosOfDir visits all of the TODOs in a particular directory
func WalkTodosOfDir(dirpath string, visit func(todo Todo) error) error {
	return filepath.Walk(dirpath, func(filepath string, info os.FileInfo, err error) error {
		if !info.IsDir() && !strings.HasPrefix(filepath, ".") {
			err := WalkTodosOfFile(filepath, visit)

			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (todo Todo) RetrieveGithubStatus(creds GithubCredentials, repo string) (string, error) {
	// TODO: duplicate code in Todo.RetrieveGithubStatus() and ReportTodo
	client := &http.Client{}
	url := "https://api.github.com/repos/" + repo + "/issues/" + (*todo.ID)[1:]

	req, err := http.NewRequest(
		// TODO: possible GitHub API injection attack
		"GET", url,
		nil)

	if err != nil {
		return "", nil
	}
	req.Header.Add("Authorization", "token "+creds.PersonalToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", err
	}

	return v["state"].(string), nil
}

// ReportTodo reports the todo as a Github Issue, updates the file
// where the todo is located and commits the changes to the git repo.
func ReportTodo(todo Todo, creds GithubCredentials, repo string, body string) (Todo, error) {
	// TODO: ReportTodo is not a Todo method
	client := &http.Client{}

	bodyBuffer := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuffer).Encode(map[string]interface{}{
		"title": todo.Suffix,
		"body":  body,
	})
	if err != nil {
		return todo, err
	}

	req, err := http.NewRequest(
		"POST", "https://api.github.com/repos/"+repo+"/issues",
		bodyBuffer)
	if err != nil {
		return todo, nil
	}
	req.Header.Add("Authorization", "token "+creds.PersonalToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return todo, err
	}
	defer resp.Body.Close()

	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return todo, err
	}

	id := "#" + strconv.Itoa(int(v["number"].(float64)))
	todo.ID = &id

	return todo, err
}

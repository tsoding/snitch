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

type Todo struct {
	Prefix   string
	Suffix   string
	Id       *string
	Filename string
	Line     int
}

func (todo Todo) TodoString() string {
	if todo.Id == nil {
		return fmt.Sprintf("TODO")
	} else {
		return fmt.Sprintf("TODO(%s)", *todo.Id)
	}
}

func (todo Todo) LogString() string {
	if todo.Id == nil {
		return fmt.Sprintf("%s:%d: %sTODO: %s",
			todo.Filename, todo.Line,
			todo.Prefix, todo.Suffix)
	} else {
		return fmt.Sprintf("%s:%d: %sTODO(%s): %s",
			todo.Filename, todo.Line,
			todo.Prefix, *todo.Id, todo.Suffix)
	}
}

func (todo Todo) String() string {
	if todo.Id == nil {
		return fmt.Sprintf("%sTODO: %s",
			todo.Prefix, todo.Suffix)
	} else {
		return fmt.Sprintf("%sTODO(%s): %s",
			todo.Prefix, *todo.Id, todo.Suffix)
	}
}

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

func LineAsUnreportedTodo(line string) *Todo {
	unreportedTodo := regexp.MustCompile("^(.*)TODO: (.*)$")
	groups := unreportedTodo.FindStringSubmatch(line)

	if groups != nil {
		return &Todo{
			Prefix:   groups[1],
			Suffix:   groups[2],
			Id:       nil,
			Filename: "",
			Line:     0,
		}
	}

	return nil
}

func LineAsReportedTodo(line string) *Todo {
	unreportedTodo := regexp.MustCompile("^(.*)TODO\\((.*)\\): (.*)$")
	groups := unreportedTodo.FindStringSubmatch(line)

	if groups != nil {
		return &Todo{
			Prefix:   groups[1],
			Suffix:   groups[3],
			Id:       &groups[2],
			Filename: "",
			Line:     0,
		}
	}

	return nil
}

func LineAsTodo(line string) *Todo {
	// TODO(#2): LineAsTodo has false positive result inside of string literals
	if todo := LineAsUnreportedTodo(line); todo != nil {
		return todo
	}

	if todo := LineAsReportedTodo(line); todo != nil {
		return todo
	}

	return nil
}

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

func ReportTodo(todo Todo, creds GithubCredentials, repo string) (Todo, error) {
	client := &http.Client{}

	// TODO(#28): ReportTodo doesn't use a proper json library to encode json
	var jsonBody = []byte(`{"title": "` + todo.Suffix + `"}`)

	req, err := http.NewRequest(
		"POST", "https://api.github.com/repos/"+repo+"/issues",
		bytes.NewBuffer(jsonBody))
	if err != nil {
		return Todo{}, nil
	}
	req.Header.Add("Authorization", "token "+creds.PersonalToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return todo, err
	}
	defer resp.Body.Close()

	var v map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&v); err != nil {
		return todo, err
	}

	todo.Id = ref_str("#" + strconv.Itoa(int(v["number"].(float64))))

	return todo, err
}

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"sync"
)

// TransformRule defines a title transformation rule
type TransformRule struct {
	Match   string
	Replace string
}

// Transform applies a title transformation rule
func (transformRule *TransformRule) Transform(title string) string {
	matchRegexp := regexp.MustCompile(transformRule.Match)
	return string(matchRegexp.ReplaceAll(
		[]byte(title), []byte(transformRule.Replace)))
}

// TitleConfig contains project level configuration related to issue titles
type TitleConfig struct {
	Transforms []*TransformRule
}

// Transform transforms the suffix into the title
func (titleConfig *TitleConfig) Transform(title string) string {
	for _, rule := range titleConfig.Transforms {
		title = rule.Transform(title)
	}

	return title
}

const defaultBodySeparator = "---"

// Project contains the project level configuration
type Project struct {
	Title         *TitleConfig
	Keywords      []string `yaml:"omitempty"`
	BodySeparator string
}

func unreportedTodoRegexp(keyword string) string {
	return "^(.*)" + regexp.QuoteMeta(keyword) + ": (.*)$"
}

func reportedTodoRegexp(keyword string) string {
	return "^(.*)" + regexp.QuoteMeta(keyword) + "\\((.*)\\): (.*)$"
}

func (project Project) lineAsUnreportedTodo(line string) *Todo {
	for _, keyword := range project.Keywords {
		unreportedTodo := regexp.MustCompile(
			unreportedTodoRegexp(keyword))
		groups := unreportedTodo.FindStringSubmatch(line)

		if groups != nil {
			prefix := groups[1]
			suffix := groups[2]
			title := project.Title.Transform(suffix)

			return &Todo{
				Prefix:        prefix,
				Suffix:        suffix,
				Keyword:       keyword,
				ID:            nil,
				Filename:      "",
				Line:          0,
				Title:         title,
				BodySeparator: project.BodySeparator,
			}
		}
	}

	return nil
}

func (project Project) lineAsReportedTodo(line string) *Todo {
	for _, keyword := range project.Keywords {
		unreportedTodo := regexp.MustCompile(reportedTodoRegexp(keyword))
		groups := unreportedTodo.FindStringSubmatch(line)

		if groups != nil {
			prefix := groups[1]
			suffix := groups[3]
			id := groups[2]
			title := project.Title.Transform(suffix)

			return &Todo{
				Prefix:        prefix,
				Suffix:        suffix,
				Keyword:       keyword,
				ID:            &id,
				Filename:      "",
				Line:          0,
				Title:         title,
				BodySeparator: project.BodySeparator,
			}
		}
	}

	return nil
}

// LineAsTodo constructs a Todo from a string
func (project Project) LineAsTodo(line string) *Todo {
	if todo := project.lineAsUnreportedTodo(line); todo != nil {
		return todo
	}

	if todo := project.lineAsReportedTodo(line); todo != nil {
		return todo
	}

	return nil
}

// WalkTodosOfFile visits all of the TODOs in a particular file
func (project Project) WalkTodosOfFile(path string, visit func(Todo) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var todo *Todo

	text, _, err := reader.ReadLine()
	for line := 1; err == nil; line = line + 1 {
		if todo == nil { // LookingForTodo
			todo = project.LineAsTodo(string(text))

			if todo != nil { // Switch to CollectingBody
				todo.Filename = path
				todo.Line = line
			}
		} else { // CollectingBody
			if possibleTodo := project.LineAsTodo(string(text)); possibleTodo != nil {
				if err := visit(*todo); err != nil {
					return err
				}

				todo = possibleTodo // Remain in CollectingBody but for the next todo
				todo.Filename = path
				todo.Line = line
			} else if todo.IsBodySeperator(string(text)) {
				if err := visit(*todo); err != nil {
					return err
				}
				todo = nil // Switch to LookingForTodo
			} else if bodyLine := todo.ParseBodyLine(string(text)); bodyLine != nil {
				todo.Body = append(todo.Body, *bodyLine)
			} else {
				if err := visit(*todo); err != nil {
					return err
				}

				todo = nil // Switch to LookingForTodo
			}
		}

		text, _, err = reader.ReadLine()
	}

	if todo != nil {
		if err := visit(*todo); err != nil {
			return err
		}
		todo = nil // Switch to LookingForTodo
	}

	if err != io.EOF {
		return err
	}

	return nil
}

// TodoResult is used by WalkTodosOfDir and contains a todo and an error.
// The error may be raised by WalkTodosOfFile() or os.Stat() i.e. when a
// file doesn't exist.
type TodoResult struct {
	todo *Todo
	err  error
}

// WalkTodosOfDir visits all of the TODOs in a particular directory
func (project Project) WalkTodosOfDir(dirpath string) (<-chan TodoResult, context.CancelFunc, error) {
	cmd := exec.Command("git", "ls-files", dirpath)
	var outb bytes.Buffer
	cmd.Stdout = &outb

	err := cmd.Run()
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	work := make(chan string)
	output := make(chan TodoResult)
	var workers sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()

			var workerErr error
			var stat os.FileInfo
			for filepath := range work {
				stat, workerErr = os.Stat(filepath)
				if workerErr != nil {
					output <- TodoResult{nil, workerErr}
					continue
				}
				if stat.IsDir() {
					// FIXME(#145): snitch should go inside of git submodules recursively
					fmt.Printf("[WARN] `%s` is probably a submodule. Skipping it for now...\n", filepath)
					continue
				}

				workerErr = project.WalkTodosOfFile(filepath, func(todo Todo) error {
					output <- TodoResult{&todo, nil}
					return nil
				})
				if workerErr != nil {
					output <- TodoResult{nil, workerErr}
				}
			}
		}()
	}

	go func() {
		defer cancel()
		defer close(output)
		defer workers.Wait()
		defer close(work)

		for scanner := bufio.NewScanner(&outb); scanner.Scan(); {
			select {
			case <-ctx.Done():
				return
			default:
				work <- scanner.Text()
			}
		}
	}()

	return output, cancel, nil
}

func yamlConfigPath(projectPath string) (string, bool) {
	for _, suffix := range [2]string{"yaml", "yml"} {
		path := path.Join(projectPath, fmt.Sprintf(".snitch.%s", suffix))

		if stat, err := os.Stat(path); !os.IsNotExist(err) && !stat.IsDir() {
			return path, true
		}
	}

	return "", false
}

// NewProject constructs the Project from a YAML file
func NewProject(filePath string) (*Project, error) {
	project := &Project{
		Title: &TitleConfig{
			Transforms: []*TransformRule{},
		},
		Keywords:      []string{},
		BodySeparator: defaultBodySeparator,
	}

	if configPath, ok := yamlConfigPath(filePath); ok {
		configFile, err := os.Open(configPath)
		if err != nil {
			return nil, err
		}
		defer configFile.Close()

		yamlDecoder := yaml.NewDecoder(configFile)
		err = yamlDecoder.Decode(&project)
		if err != nil {
			return nil, errors.Wrap(err, configPath)
		}
	}

	if len(project.Keywords) == 0 {
		project.Keywords = []string{"TODO"}
	}

	return project, nil
}

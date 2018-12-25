package main

import (
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
	"bufio"
	"io"
	"os/exec"
	"bytes"
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

// Project contains the project level configuration
type Project struct {
	Title *TitleConfig
}

func (project Project) lineAsUnreportedTodo(line string) *Todo {
	unreportedTodo := regexp.MustCompile("^(.*)TODO: (.*)$")
	groups := unreportedTodo.FindStringSubmatch(line)

	if groups != nil {
		prefix := groups[1]
		suffix := groups[2]
		title := project.Title.Transform(suffix)

		return &Todo{
			Prefix:   prefix,
			Suffix:   suffix,
			ID:       nil,
			Filename: "",
			Line:     0,
			Title:    title,
		}
	}

	return nil
}

func (project Project) lineAsReportedTodo(line string) *Todo {
	unreportedTodo := regexp.MustCompile("^(.*)TODO\\((.*)\\): (.*)$")
	groups := unreportedTodo.FindStringSubmatch(line)

	if groups != nil {
		prefix := groups[1]
		suffix := groups[3]
		id := groups[2]
		title := project.Title.Transform(suffix)

		return &Todo{
			Prefix:   prefix,
			Suffix:   suffix,
			ID:       &id,
			Filename: "",
			Line:     0,
			Title:    title,
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
func (project Project) WalkTodosOfFile (path string, visit func(Todo) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	text, _, err := reader.ReadLine()
	for line := 1; err == nil; line = line + 1 {
		todo := project.LineAsTodo(string(text))

		if todo != nil {
			todo.Filename = path
			todo.Line = line

			if err := visit(*todo); err != nil {
				return err
			}
		}

		text, _, err = reader.ReadLine()
	}

	if err != io.EOF {
		return err
	}

	return nil
}

// WalkTodosOfDir visits all of the TODOs in a particular directory
func (project Project) WalkTodosOfDir (dirpath string, visit func(todo Todo) error) error {
	cmd := exec.Command("git", "ls-files", dirpath)
	var outb bytes.Buffer
	cmd.Stdout = &outb

	err := cmd.Run()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(&outb)

	for scanner.Scan() {
		filepath := scanner.Text()
		err = project.WalkTodosOfFile(filepath, visit)
		if err != nil {
			return err
		}
	}

	return err
}

// NewProject constructs the Project from a YAML file
func NewProject(filePath string) (*Project, error) {
	if stat, err := os.Stat(filePath); os.IsNotExist(err) || stat.IsDir() {
		return &Project{
			Title: &TitleConfig{
				Transforms: []*TransformRule{},
			},
		}, nil
	}

	configFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	yamlDecoder := yaml.NewDecoder(configFile)
	var project Project
	err = yamlDecoder.Decode(&project)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

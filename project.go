package main

import (
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
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

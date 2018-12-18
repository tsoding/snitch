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

// ProjectConfig contains the project level configuration
type ProjectConfig struct {
	Title *TitleConfig
}

// NewProjectConfig constructs the ProjectConfig from a YAML file
func NewProjectConfig(filePath string) (*ProjectConfig, error) {
	if stat, err := os.Stat(filePath); os.IsNotExist(err) || stat.IsDir() {
		return &ProjectConfig{
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
	var projectConfig ProjectConfig
	err = yamlDecoder.Decode(&projectConfig)
	if err != nil {
		return nil, err
	}

	return &projectConfig, nil
}

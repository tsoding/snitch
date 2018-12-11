package main

import (
	"regexp"
	"os"
	"fmt"
	"gopkg.in/yaml.v2"
)

// TransformRule defines a title transformation rule
type TransformRule struct {
	Match   *regexp.Regexp
	Replace string
}

// Transform applies a title transformation rule
func (transformRule *TransformRule) Transform(title string) string {
	// TODO(#111): TransformRule.Transform()
	return title
}

// TitleConfig contains project level configuration related to issue titles
type TitleConfig struct {
	TransformRules []TransformRule
}

// Transform transforms the suffix into the title
func (titleConfig *TitleConfig) Transform(title string) string {
	for _, rule := range titleConfig.TransformRules {
		title = rule.Transform(title)
	}

	return title
}

func NewTitleConfig(v interface{}) (*TitleConfig, error) {
	// TODO: NewTitleConfig is not implemented
	return nil, nil
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
				TransformRules: []TransformRule{},
			},
		}, nil
	}

	configFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	yamlDecoder := yaml.NewDecoder(configFile)
	v := make(map[string]interface{})
	err = yamlDecoder.Decode(&v)
	if err != nil {
		return nil, err
	}

	titleV, ok := v["title"]
	if !ok {
		return nil, fmt.Errorf("%s doesn't have the `title` section", filePath)
	}

	titleConfig, err := NewTitleConfig(titleV)
	if err != nil {
		return nil, err
	}

	return &ProjectConfig {
		Title: titleConfig,
	}, nil
}

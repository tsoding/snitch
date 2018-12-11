package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
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

func NewTransformRule(v interface{}) (TransformRule, error) {
	// TODO: NewTransformRule is not implemented
	return TransformRule {
		Match: nil,
		Replace: "",
	}, nil
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

// NewTitleConfig constructs a new TitleConfig from a configuration tree
func NewTitleConfig(v interface{}) (*TitleConfig, error) {
	titleV, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("NewTitleConfig expected map[string]interface as an input")
	}

	transformV, ok := titleV["transform"]
	if !ok {
		return nil, fmt.Errorf("Project title config doesn't have the `transform` section")
	}

	transformVs, ok := transformV.([]interface{})
	if !ok {
		return nil, fmt.Errorf("`transform` section of project title config is not a list")
	}

	transformRules := []TransformRule{}

	for _, tv := range transformVs {
		transformRule, err := NewTransformRule(tv)
		if err != nil {
			return nil, err
		}

		transformRules = append(transformRules, transformRule)
	}

	return &TitleConfig{
		TransformRules: transformRules,
	}, nil
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

	return &ProjectConfig{
		Title: titleConfig,
	}, nil
}

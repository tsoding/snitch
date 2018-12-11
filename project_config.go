package main

import (
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

// ProjectConfig contains the project level configuration
type ProjectConfig struct {
	Title *TitleConfig
}

// NewProjectConfig constructs the ProjectConfig from a YAML file
func NewProjectConfig(filePath string) (*ProjectConfig, error) {
	// TODO(#107): NewProjectConfig() is not implemented
	return &ProjectConfig{
		Title: &TitleConfig{},
	}, nil
}

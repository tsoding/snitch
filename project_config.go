package main

import ()

// TitleConfig contains project level configuration related to issue titles
type TitleConfig struct {
}

// Transform transforms the suffix into the title
func (titleConfig *TitleConfig) Transform(title string) string {
	// TODO(#106): TitleConfig.Transform() is not implemented
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

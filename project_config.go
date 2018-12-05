package main

import ()

type TitleConfig struct {
}

func (titleConfig *TitleConfig) Transform(title string) string {
	// TODO: TitleConfig.Transform() is not implemented
	return title
}

type ProjectConfig struct {
	Title *TitleConfig
}

func NewProjectConfig(filePath string) (*ProjectConfig, error) {
	// TODO: NewProjectConfig() is not implemented
	return &ProjectConfig{
		Title: &TitleConfig{},
	}, nil
}

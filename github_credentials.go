package main

import (
	"gopkg.in/go-ini/ini.v1"
)

// GithubCredentials contains PersonalToken for GitHub API authorization
type GithubCredentials struct {
	PersonalToken string
}

// GithubCredentialsFromFile gets GithubCredentials from a filepath
func GithubCredentialsFromFile(filepath string) (GithubCredentials, error) {
	cfg, err := ini.Load(filepath)
	if err != nil {
		return GithubCredentials{}, err
	}

	return GithubCredentials{
		PersonalToken: cfg.Section("github").Key("personal_token").String(),
	}, nil
}

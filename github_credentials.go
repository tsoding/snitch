package main

import (
	"gopkg.in/go-ini/ini.v1"
)

type GithubCredentials struct {
	PersonalToken string
}

func GithubCredentialsFromFile(filepath string) (GithubCredentials, error) {
	cfg, err := ini.Load(filepath)
	if err != nil {
		return GithubCredentials{}, err
	}

	return GithubCredentials {
		PersonalToken : cfg.Section("github").Key("personal_token").String(),
	}, nil
}


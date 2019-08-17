package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/go-ini/ini.v1"
	"net/http"
)

// GithubCredentials contains PersonalToken for GitHub API authorization
type GithubCredentials struct {
	PersonalToken string
}

// QueryGithubAPI makes a GitHub API query
func (creds GithubCredentials) QueryGithubAPI(method, url string, jsonBody map[string]interface{}) (map[string]interface{}, error) {
	client := &http.Client{}

	bodyBuffer := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuffer).Encode(jsonBody)

	req, err := http.NewRequest(
		method, url, bodyBuffer)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "token "+creds.PersonalToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", buf.String())
	}

	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	return v, err
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

// GithubCredentialsFromToken returns a GithubCredentials from a string token
func GithubCredentialsFromToken(token string) GithubCredentials {
	return GithubCredentials{
		PersonalToken: token,
	}
}

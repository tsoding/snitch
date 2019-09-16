package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"

	"gopkg.in/go-ini/ini.v1"
)

// GithubCredentials contains PersonalToken for GitHub API authorization
// and Repository to be used in request
type GithubCredentials struct {
	PersonalToken string
	Repository    string
}

func (creds GithubCredentials) query(method, url string, jsonBody map[string]interface{}) (map[string]interface{}, error) {
	bodyBuffer := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuffer).Encode(jsonBody)

	req, err := http.NewRequest(method, url, bodyBuffer)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "token "+creds.PersonalToken)
	req.Header.Add("Content-Type", "application/json")

	return QueryHTTP(req)
}

func (creds GithubCredentials) getIssue(todo Todo) (map[string]interface{}, error) {
	json, err := creds.query(
		"GET",
		// FIXME(#59): possible GitHub API injection attack
		"https://api.github.com/repos/"+creds.Repository+"/issues/"+(*todo.ID)[1:],
		nil)

	if err != nil {
		return nil, err
	}

	return json, nil
}

func (creds GithubCredentials) postIssue(todo Todo, body string) (Todo, error) {
	json, err := creds.query(
		"POST",
		"https://api.github.com/repos/"+creds.Repository+"/issues",
		map[string]interface{}{
			"title": todo.Title,
			"body":  body,
		})
	if err != nil {
		return todo, err
	}

	id := "#" + strconv.Itoa(int(json["number"].(float64)))
	todo.ID = &id

	return todo, err
}

func (creds GithubCredentials) getHost() string {
	return "github.com"
}

func (creds GithubCredentials) getRepositoryAddress() string {
	return "https://github.com/" + creds.Repository
}

func (creds GithubCredentials) setRepository(repo string) Repo {
	return GithubCredentials{
		PersonalToken: creds.PersonalToken,
		Repository:    repo,
	}
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

func getGithubCredentials() (GithubCredentials, error) {
	tokenEnvar := os.Getenv("GITHUB_PERSONAL_TOKEN")
	xdgEnvar := os.Getenv("XDG_CONFIG_HOME")
	usr, err := user.Current()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(tokenEnvar) != 0 {
		return GithubCredentialsFromToken(tokenEnvar), nil
	}

	// custom XDG_CONFIG_HOME
	if len(xdgEnvar) != 0 {
		filePath := path.Join(xdgEnvar, "snitch/github.ini")
		if _, err := os.Stat(filePath); err == nil {
			return GithubCredentialsFromFile(filePath)
		}
	}

	// default XDG_CONFIG_HOME
	if len(xdgEnvar) == 0 {
		filePath := path.Join(usr.HomeDir, ".config/snitch/github.ini")
		if _, err := os.Stat(filePath); err == nil {
			return GithubCredentialsFromFile(filePath)
		}
	}

	filePath := path.Join(usr.HomeDir, ".snitch/github.ini")
	if _, err := os.Stat(filePath); err == nil {
		return GithubCredentialsFromFile(filePath)
	}

	return GithubCredentials{}, fmt.Errorf("GitHub token is missing")
}

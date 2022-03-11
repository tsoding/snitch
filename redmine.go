package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/ini.v1"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"
)

// RedmineCredentials contains PersonalToken for Redmine API authorization
type RedmineCredentials struct {
	PersonalToken string
}

func (creds RedmineCredentials) query(method, url string, jsonBody map[string]interface{}) (map[string]interface{}, error) {
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

func (creds RedmineCredentials) getIssue(repo string, todo Todo) (map[string]interface{}, error) {
	json, err := creds.query(
		"GET",
		// TODO: make this configurable via ini file
		fmt.Sprintf("https://redmine.sighup-prod.sighup.io/issues/%s.json", (*todo.ID)[1:]),
		nil)

	if err != nil {
		return nil, err
	}

	fmt.Println(json)

	return json, nil
}

func (creds RedmineCredentials) postIssue(repo string, todo Todo, body string) (Todo, error) {
	json, err := creds.query(
		"POST",
		"https://api.github.com/repos/"+repo+"/issues",
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

func (creds RedmineCredentials) getHost() string {
	return "github.com"
}

// RedmineCredentialsFromFile gets RedmineCredentials from a filepath
func RedmineCredentialsFromFile(filepath string) (RedmineCredentials, error) {
	cfg, err := ini.Load(filepath)
	if err != nil {
		return RedmineCredentials{}, err
	}

	return RedmineCredentials{
		PersonalToken: cfg.Section("redmine").Key("personal_token").String(),
	}, nil
}

// RedmineCredentialsFromToken returns a RedmineCredentials from a string token
func RedmineCredentialsFromToken(token string) RedmineCredentials {
	return RedmineCredentials{
		PersonalToken: token,
	}
}

func getRedmineCredentials() (RedmineCredentials, error) {
	tokenEnvar := os.Getenv("GITHUB_PERSONAL_TOKEN")
	xdgEnvar := os.Getenv("XDG_CONFIG_HOME")
	usr, err := user.Current()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(tokenEnvar) != 0 {
		return RedmineCredentialsFromToken(tokenEnvar), nil
	}

	// custom XDG_CONFIG_HOME
	if len(xdgEnvar) != 0 {
		filePath := path.Join(xdgEnvar, "snitch/redmine.ini")
		if _, err := os.Stat(filePath); err == nil {
			return RedmineCredentialsFromFile(filePath)
		}
	}

	// default XDG_CONFIG_HOME
	if len(xdgEnvar) == 0 {
		filePath := path.Join(usr.HomeDir, ".config/snitch/redmine.ini")
		if _, err := os.Stat(filePath); err == nil {
			return RedmineCredentialsFromFile(filePath)
		}
	}

	filePath := path.Join(usr.HomeDir, ".snitch/redmine.ini")
	if _, err := os.Stat(filePath); err == nil {
		return RedmineCredentialsFromFile(filePath)
	}

	return RedmineCredentials{}, fmt.Errorf("Redmine token is missing")
}

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

	"gopkg.in/ini.v1"
)

// RedmineSpec contains PersonalToken for Redmine API authorization
type RedmineSpec struct {
	PersonalToken string
	BaseURL       string
}

func (creds RedmineSpec) query(method, url string, jsonBody map[string]interface{}) (map[string]interface{}, error) {
	bodyBuffer := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuffer).Encode(jsonBody)

	req, err := http.NewRequest(method, url, bodyBuffer)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Redmine-API-Key", creds.PersonalToken)
	req.Header.Add("Content-Type", "application/json")

	return QueryHTTP(req)
}

func (creds RedmineSpec) getIssue(repo string, todo Todo) (map[string]interface{}, error) {
	json, err := creds.query(
		"GET",
		fmt.Sprintf("%s/issues/%s.json", creds.BaseURL, (*todo.ID)[1:]),
		nil,
	)

	if err != nil {
		return nil, err
	}

	fmt.Println(json)

	return json, nil
}

func (creds RedmineSpec) getProject() {
	//https://redmine.sighup-prod.sighup.io/search.json?q=fury-fleet-api&projects=1&titles_only=1
	json, err := creds.query(
		"GET",
		fmt.Sprintf("%s/search.json?q=fury-fleet-api&projects=1&titles_only=1", creds.BaseURL),
		map[string]interface{}{
			"subject":     todo.Title,
			"description": body,
			"project_id":  creds.getProject(),
		},
}

func (creds RedmineSpec) postIssue(repo string, todo Todo, body string) (Todo, error) {
	json, err := creds.query(
		"POST",
		fmt.Sprintf("%s/issues.json", creds.BaseURL),
		map[string]interface{}{
			"subject":     todo.Title,
			"description": body,
			"project_id":  creds.getProject(),
		},
	)
	if err != nil {
		return todo, err
	}

	id := "#" + strconv.Itoa(int(json["number"].(float64)))
	todo.ID = &id

	return todo, err
}

func (creds RedmineSpec) getHost() string {
	return "github.com"
}

// RedmineCredentialsFromFile gets RedmineSpec from a filepath
func RedmineCredentialsFromFile(filepath string) (RedmineSpec, error) {
	cfg, err := ini.Load(filepath)
	if err != nil {
		return RedmineSpec{}, err
	}

	return RedmineSpec{
		PersonalToken: cfg.Section("redmine").Key("personal_token").String(),
		BaseURL:       cfg.Section("redmine").Key("base_url").String(),
	}, nil
}

// RedmineCredentialsFromToken returns a RedmineSpec from a string token
func RedmineCredentialsFromToken(token string) RedmineSpec {
	return RedmineSpec{
		PersonalToken: token,
	}
}

func getRedmineCredentials() (RedmineSpec, error) {
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

	return RedmineSpec{}, fmt.Errorf("Redmine token is missing")
}

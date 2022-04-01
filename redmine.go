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

type SearchQuery struct {
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
	Results    []Issue `json:"results"`
	TotalCount int     `json:"total_count"`
}

type Issue struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Project     string `json:"project"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Datetime    string `json:"datetime"`
}

func (creds RedmineSpec) search(method, url string, jsonBody map[string]interface{}) (SearchQuery, error) {
	bodyBuffer := new(bytes.Buffer)
	err := json.NewEncoder(bodyBuffer).Encode(jsonBody)

	req, err := http.NewRequest(method, url, bodyBuffer)
	if err != nil {
		return SearchQuery{}, err
	}

	req.Header.Add("X-Redmine-API-Key", creds.PersonalToken)
	req.Header.Add("Content-Type", "application/json")

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

func (creds *RedmineSpec) checkIfIssueExists(issueID *string) (bool, error) {
	url := creds.BaseURL + "/issues.json?issue_id=" + *issueID

	resp, err := creds.query("GET", url, nil)
	if err != nil {
		return false, err
	}

	if resp["total_count"] == 0 {
		return false, nil
	}

	return true, nil
}

//PIPPO: this is a test
func (creds RedmineSpec) getIssue(repo string, todo Todo) (map[string]interface{}, error) {

	ok, err := creds.checkIfIssueExists(todo.ID)

	if !ok {
		return nil, fmt.Errorf("Issue %s not found", *(todo.ID))
	}

	if err != nil {
		return nil, err
	}

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

func (creds RedmineSpec) getProject(project string) (string, error) {
	query, err := creds.query(
		"GET",
		fmt.Sprintf("%s/search.json?q=%s&projects=1&titles_only=1", creds.BaseURL, project),
		nil,
	)

	if err != nil {
		return "", err
	}

	if query["total_count"] == 0 {
		return "", fmt.Errorf("Project %s not found", project)
	}

	return query["results"][0]["id"], nil

	//project :=
	//
	//return query["results"][0]
}

func (creds RedmineSpec) postIssue(repo string, todo Todo, body string) (Todo, error) {
	//project := strings.Split(repo, "/")[1]
	project := "test-gh-rm-integration"
	projectID, err := creds.getProject(project)

	if err != nil {
		return Todo{}, err
	}

	json, err := creds.query(
		"POST",
		fmt.Sprintf("%s/issues.json", creds.BaseURL),
		map[string]interface{}{
			"subject":     todo.Title,
			"description": body,
			"project_id":  projectID,
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

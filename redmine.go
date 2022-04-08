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
	"strings"

	"gopkg.in/ini.v1"
)

// RedmineSpec contains PersonalToken for Redmine API authorization
type RedmineSpec struct {
	PersonalToken string
	BaseURL       string
	TrackerID     string
}

type SearchQuery struct {
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
	Results    []Issue `json:"results"`
	TotalCount int     `json:"total_count"`
}

type Issue struct {
	ID          int            `json:"id"`
	Subject     string         `json:"subject"`
	Description string         `json:"description"`
	Project     RedmineProject `json:"project"`
	Type        string         `json:"type"`
	URL         string         `json:"url"`
	Datetime    string         `json:"datetime"`
}

type RedmineProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//FIXME: переделать на простой массив и проверку на наличие в нем значения
func (creds RedmineSpec) search(url string) (SearchQuery, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return SearchQuery{}, err
	}

	req.Header.Add("X-Redmine-API-Key", creds.PersonalToken)
	req.Header.Add("Content-Type", "application/json")

	return SearchQueryHTTP(req)
}

func (creds RedmineSpec) postIssueQuery(method, url string, jsonBody map[string]interface{}) (map[string]interface{}, error) {
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

func (creds RedmineSpec) IsClosed(status string) bool {
	if status == "Done" {
		return true
	}
	return false
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

//PIPPO(#2027): this is a test
func (creds RedmineSpec) getIssue(repo string, todo Todo) (map[string]interface{}, error) {

	id := (*todo.ID)[1:]
	ok, err := creds.checkIfIssueExists(&id)

	if !ok {
		return nil, fmt.Errorf("Redmine Issue %s not found", *(todo.ID))
	}

	if err != nil {
		return nil, err
	}

	json, err := creds.query(
		"GET",
		fmt.Sprintf("%s/issues/%s.json", creds.BaseURL, id),
		nil,
	)

	if err != nil {
		return nil, err
	}

	json["state"] = json["issue"].(map[string]interface{})["status"].(map[string]interface{})["name"]

	return json, nil
}

func (creds RedmineSpec) getProject(project string) (int, error) {
	query, err := creds.search(
		fmt.Sprintf("%s/search.json?q=%s&projects=1&titles_only=1", creds.BaseURL, project),
	)
	if err != nil {
		return 0, err
	}

	if query.TotalCount == 0 {
		return 0, fmt.Errorf("Project %s not found", project)
	}

	return query.Results[0].ID, nil
}

func (creds RedmineSpec) postIssue(repo string, todo Todo, body string) (Todo, error) {
	project := strings.Split(repo, "/")[1]
	projectID, err := creds.getProject(project)

	if err != nil {
		return Todo{}, err
	}

	json, err := creds.postIssueQuery(
		"POST",
		fmt.Sprintf("%s/issues.json", creds.BaseURL),
		map[string]interface{}{
			"issue": map[string]interface{}{
				"subject":     todo.Title,
				"description": body,
				"project_id":  projectID,
				"tracker_id":  creds.TrackerID,
			},
		},
	)
	if err != nil {
		return todo, err
	}

	id := "#" + strconv.Itoa(json["issue"].(map[string]interface{})["id"].(int))
	todo.ID = &id

	return todo, err
}

func (creds RedmineSpec) getHost() string {
	return ""
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
		TrackerID:     cfg.Section("redmine").Key("tracker_id").String(),
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

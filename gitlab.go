package main

import (
	"fmt"
	"gopkg.in/go-ini/ini.v1"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"strconv"
)

// GitlabCredentials contains PersonalToken for GitLab API authorization
// and Host for possibly implementing support for self-hosted instances
type GitlabCredentials struct {
	// TODO: Self-hosted GitLab instance
	Host string
	PersonalToken string
}

func (creds GitlabCredentials) query(method, url string) (map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("PRIVATE-TOKEN", creds.PersonalToken)

	return QueryHTTP(req)
}

func (creds GitlabCredentials) getIssue(repo string, todo Todo) (map[string]interface{}, error) {
	json, err := creds.query(
		"GET",
		// FIXME: possible GitLab API injection attack
		"https://"+creds.Host+"/api/v4/projects/"+url.QueryEscape(repo)+"/issues/"+(*todo.ID)[1:]) // self-hosted

	if err != nil {
		return nil, err
	}

	return json, nil
}

func (creds GitlabCredentials) postIssue(repo string, todo Todo, body string) (Todo, error) {
	params := url.Values{}
	params.Add("title", todo.Title)
	params.Add("description", body)

	json, err := creds.query(
		"POST",
		"https://"+creds.Host+"/api/v4/projects/"+url.QueryEscape(repo)+"/issues" + params.Encode()) // self-hosted
	if err != nil {
		return todo, err
	}

	id := "#" + strconv.Itoa(int(json["iid"].(float64)))
	todo.ID = &id

	return todo, err
}

func (creds GitlabCredentials) getHost() string {
	return creds.Host
}


// GitlabCredentialsFromFile gets GitlabCredentials from a filepath
func GitlabCredentialsFromFile(filepath string) []GitlabCredentials {
	credentials := []GitlabCredentials{}

	cfg, err := ini.Load(filepath)
	if err != nil {
		return credentials
	}

	sections := cfg.Sections()
	for i := 1; i < len(sections); i++ {
		fmt.Printf("[%d]: %s\n", i, sections[i].Name())
		credentials = append(credentials, GitlabCredentials{
			Host: sections[i].Name(),
			PersonalToken: sections[i].Key("personal_token").String(),
		})
	}

	return credentials
}

// GitlabCredentialsFromToken returns a GitlabCredentials from a string token
func GitlabCredentialsFromToken(token string) GitlabCredentials {
	return GitlabCredentials{
		Host: "gitlab.com",
		PersonalToken: token,
	}
}

func getGitlabCredentials(creds []IssueAPI) []IssueAPI {
	tokenEnvar := os.Getenv("GITLAB_PERSONAL_TOKEN")
	xdgEnvar := os.Getenv("XDG_CONFIG_HOME")
	usr, err := user.Current()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(tokenEnvar) != 0 {
		// TODO: Support multiple hosts from ENV
		creds = append(creds, GitlabCredentialsFromToken(tokenEnvar))
	}

	// custom XDG_CONFIG_HOME
	if len(xdgEnvar) != 0 {
		filePath := path.Join(xdgEnvar, "snitch/gitlab.ini")
		if _, err := os.Stat(filePath); err == nil {
			for _, cred := range GitlabCredentialsFromFile(filePath) {
				creds = append(creds, cred)
			}
		}
	}

	// default XDG_CONFIG_HOME
	if len(xdgEnvar) == 0 {
		filePath := path.Join(usr.HomeDir, ".config/snitch/gitlab.ini")
		if _, err := os.Stat(filePath); err == nil {
			for _, cred := range GitlabCredentialsFromFile(filePath) {
				creds = append(creds, cred)
			}
		}
	}

	filePath := path.Join(usr.HomeDir, ".snitch/gitlab.ini")
	if _, err := os.Stat(filePath); err == nil {
		for _, cred := range GitlabCredentialsFromFile(filePath) {
			creds = append(creds, cred)
		}
	}

	return creds
}
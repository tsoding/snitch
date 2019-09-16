package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"strconv"

	"gopkg.in/go-ini/ini.v1"
)

// GitlabCredentials contains PersonalToken for GitLab API authorization,
// Host for implementing support for self-hosted instances and
// Repository
type GitlabCredentials struct {
	PersonalToken string
	Host          string
	Repository    string
}

func (creds GitlabCredentials) query(method, url string) (map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("PRIVATE-TOKEN", creds.PersonalToken)

	return QueryHTTP(req)
}

func (creds GitlabCredentials) getIssue(todo Todo) (map[string]interface{}, error) {
	json, err := creds.query(
		"GET",
		// FIXME(#156): possible GitLab API injection attack
		"https://"+creds.Host+"/api/v4/projects/"+url.QueryEscape(creds.Repository)+"/issues/"+(*todo.ID)[1:]) // self-hosted

	if err != nil {
		return nil, err
	}

	return json, nil
}

func (creds GitlabCredentials) postIssue(todo Todo, body string) (Todo, error) {
	params := url.Values{}
	params.Add("title", todo.Title)
	params.Add("description", body)

	json, err := creds.query(
		"POST",
		"https://"+creds.Host+"/api/v4/projects/"+url.QueryEscape(creds.Repository)+"/issues?"+params.Encode()) // self-hosted
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

func (creds GitlabCredentials) getRepositoryAddress() string {
	return "https://" + creds.Host + "/" + creds.Repository
}

func (creds GitlabCredentials) setRepository(repo string) Repo {
	return GitlabCredentials{
		Host:          creds.Host,
		PersonalToken: creds.PersonalToken,
		Repository:    repo,
	}
}

// GitlabCredentialsFromFile gets GitlabCredentials from a filepath
func GitlabCredentialsFromFile(filepath string) []GitlabCredentials {
	credentials := []GitlabCredentials{}

	cfg, err := ini.Load(filepath)
	if err != nil {
		return credentials
	}

	for _, section := range cfg.Sections()[1:] {
		credentials = append(credentials, GitlabCredentials{
			Host:          section.Name(),
			PersonalToken: section.Key("personal_token").String(),
		})
	}

	return credentials
}

// GitlabCredentialsFromToken returns a GitlabCredentials from a string token
func GitlabCredentialsFromToken(token string) GitlabCredentials {
	return GitlabCredentials{
		Host:          "gitlab.com",
		PersonalToken: token,
	}
}

func getGitlabCredentials(creds []Repo) []Repo {
	tokenEnvar := os.Getenv("GITLAB_PERSONAL_TOKEN")
	xdgEnvar := os.Getenv("XDG_CONFIG_HOME")
	usr, err := user.Current()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(tokenEnvar) != 0 {
		// FIXME(#157): Support multiple GitLab hosts from environment variables
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

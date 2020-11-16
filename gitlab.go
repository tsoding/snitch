package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

// GitlabCredentials contains PersonalToken for GitLab API authorization
// and Host for possibly implementing support for self-hosted instances
type GitlabCredentials struct {
	Host          string
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
		// FIXME(#156): possible GitLab API injection attack
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
		"https://"+creds.Host+"/api/v4/projects/"+url.QueryEscape(repo)+"/issues?"+params.Encode()) // self-hosted
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

	for _, section := range cfg.Sections()[1:] {
		credentials = append(credentials, GitlabCredentials{
			Host:          section.Name(),
			PersonalToken: section.Key("personal_token").String(),
		})
	}

	return credentials
}

// GitlabCredentialsFromToken returns a GitlabCredentials from a string token
func GitlabCredentialsFromToken(token string) (GitlabCredentials, error) {
	credentials := strings.Split(token, ":")

	switch len(credentials) {
	case 1:
		return GitlabCredentials{
			Host:          "gitlab.com",
			PersonalToken: credentials[0],
		}, nil
	case 2:
		return GitlabCredentials{
			Host:          credentials[0],
			PersonalToken: credentials[1],
		}, nil
	default:
		return GitlabCredentials{},
			fmt.Errorf("Couldn't parse GitLab credentials from ENV: %s", token)
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
		for _, credential := range strings.Split(tokenEnvar, ",") {
			parsedCredentials, err := GitlabCredentialsFromToken(credential)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			creds = append(creds, parsedCredentials)
		}
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

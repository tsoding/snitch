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

// GiteaCredentials contains PersonalToken for gitea API authorization
// and Host for possibly implementing support for self-hosted instances
type GiteaCredentials struct {
	Host          string
	PersonalToken string
}

func (creds GiteaCredentials) query(method, url string, jsonBody map[string]interface{}) (map[string]interface{}, error) {
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

func (creds GiteaCredentials) getIssue(repo string, todo Todo) (map[string]interface{}, error) {
	// FIXME(#187): gitea integration does not support http instances.
	json, err := creds.query(
		"GET",
		"https://"+creds.Host+"/api/v1/repos/"+repo+"/issues/"+(*todo.ID)[1:],
		nil) // self-hosted

	if err != nil {
		return nil, err
	}

	return json, nil
}

func (creds GiteaCredentials) postIssue(repo string, todo Todo, body string) (Todo, error) {
	json, err := creds.query(
		"POST",
		"https://"+creds.Host+"/api/v1/repos/"+repo+"/issues",
		map[string]interface{}{
			"title": todo.Title,
			"body":  body,
		}) // self-hosted
	if err != nil {
		return todo, err
	}

	id := "#" + strconv.Itoa(int(json["number"].(float64)))
	todo.ID = &id

	return todo, err
}

func (creds GiteaCredentials) getHost() string {
	return creds.Host
}

// GiteaCredentialsFromFile gets GiteaCredentials from a filepath
func GiteaCredentialsFromFile(filepath string) []GiteaCredentials {
	credentials := []GiteaCredentials{}

	cfg, err := ini.Load(filepath)
	if err != nil {
		return credentials
	}

	for _, section := range cfg.Sections()[1:] {
		credentials = append(credentials, GiteaCredentials{
			Host:          section.Name(),
			PersonalToken: section.Key("access_token").String(),
		})
	}

	return credentials
}

// GiteaCredentialsFromToken returns a GiteaCredentials from a string token
func GiteaCredentialsFromToken(token string) (GiteaCredentials, error) {
	credentials := strings.Split(token, ":")

	switch len(credentials) {
	case 2:
		return GiteaCredentials{
			Host:          credentials[0],
			PersonalToken: credentials[1],
		}, nil
	case 3:
		return GiteaCredentials{
			Host:          credentials[0] + ":" + credentials[1],
			PersonalToken: credentials[2],
		}, nil
	default:
		return GiteaCredentials{},
			fmt.Errorf("Couldn't parse gitea credentials from ENV: %s", token)
	}

}

func getGiteaCredentials(creds []IssueAPI) []IssueAPI {
	tokenEnvar := os.Getenv("GITEA_ACCESS_TOKEN")
	xdgEnvar := os.Getenv("XDG_CONFIG_HOME")
	usr, err := user.Current()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(tokenEnvar) != 0 {
		for _, credential := range strings.Split(tokenEnvar, ",") {
			parsedCredentials, err := GiteaCredentialsFromToken(credential)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			creds = append(creds, parsedCredentials)
		}
	}

	// custom XDG_CONFIG_HOME
	if len(xdgEnvar) != 0 {
		filePath := path.Join(xdgEnvar, "snitch/gitea.ini")
		if _, err := os.Stat(filePath); err == nil {
			for _, cred := range GiteaCredentialsFromFile(filePath) {
				creds = append(creds, cred)
			}
		}
	}

	// default XDG_CONFIG_HOME
	if len(xdgEnvar) == 0 {
		filePath := path.Join(usr.HomeDir, ".config/snitch/gitea.ini")
		if _, err := os.Stat(filePath); err == nil {
			for _, cred := range GiteaCredentialsFromFile(filePath) {
				creds = append(creds, cred)
			}
		}
	}

	filePath := path.Join(usr.HomeDir, ".snitch/gitea.ini")
	if _, err := os.Stat(filePath); err == nil {
		for _, cred := range GiteaCredentialsFromFile(filePath) {
			creds = append(creds, cred)
		}
	}

	return creds
}

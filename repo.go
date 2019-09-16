package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Repo requires implementing common API for querying and posting issues
// regardless of service that's being used.
type Repo interface {
	getIssue(todo Todo) (map[string]interface{}, error)
	postIssue(todo Todo, body string) (Todo, error)
	getHost() string
	getRepositoryAddress() string
	setRepository(repo string) Repo
}

// QueryHTTP makes an API query
func QueryHTTP(req *http.Request) (map[string]interface{}, error) {
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf("API error: %s", buf.String())
	}

	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	return v, err
}

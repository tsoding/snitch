package main

type IssueAPI interface {
	getIssue(repo string, todo Todo) (map[string]interface{}, error)
	postIssue(repo string, todo Todo, body string) (Todo, error)
}
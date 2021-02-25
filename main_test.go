package main

import (
	"testing"
)

func TestMain_GetRepoFromRemoteURL(t *testing.T) {
	serviceURL := "github.com"
	tests := []struct {
		in  string
		out *string
	}{
		{"https://bitbucket.org/user/project.git", nil},
		{"https://github.com/user/project", stringPtr("user/project")},
		{"https://github.com/user/project/", stringPtr("user/project")},
		{"https://github.com/user/project.git", stringPtr("user/project")},
		{"https://github.com/user/project.with.dot.git", stringPtr("user/project.with.dot")},
		{"https://github.com/long/path/to/project", stringPtr("long/path/to/project")},
		{"https://github.com/long/path/to/project.git", stringPtr("long/path/to/project")},
		{"https://github.com/long/path/to/project.git/", stringPtr("long/path/to/project")},
		{"ssh://git@github.com:22/user/project.git", stringPtr("user/project")},
		{"ssh://git@github.com:22/long/path/to/project.git", stringPtr("long/path/to/project")},
		{"ssh://git@github.com:user/project.git", stringPtr("user/project")},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := getRepoFromRemoteURL(serviceURL, tt.in); !stringPtrEqual(got, tt.out) {
				t.Errorf("got %q, want %q", derefString(got), derefString(tt.out))
			}
		})
	}
}

func TestMain_GetRepoFromRemoteURLWithPort(t *testing.T) {
	serviceURL := "gitea.com:80"
	gitRemote := "https://gitea.com:80/user/path.git"
	expected := stringPtr("user/path")
	if got := getRepoFromRemoteURL(serviceURL, gitRemote); !stringPtrEqual(got, expected) {
		t.Errorf("got %q, want %q", derefString(got), derefString(expected))
	}
}

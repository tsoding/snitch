package main

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestTodo_IsBodySeperator(t *testing.T) {
	tests := []struct {
		in  string
		out bool
	}{
		{"Kappa ---", true},
		{"--- Kappa", false},
		{"", false},
		{"Kappa ---            ", false},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			todo := Todo{BodySeparator: "---"}

			if got := todo.IsBodySeperator(tt.in); got != tt.out {
				t.Errorf("got %t, want %t", got, tt.out)
			}
		})
	}
}

func TestTodo_ParseBodyLine(t *testing.T) {
	tests := []struct {
		in  string
		out *string
	}{
		{"TODO: PogChamp", stringPtr(": PogChamp")},
		{"PogChamp", nil},
		{"", nil},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			todo := Todo{Prefix: "TODO"}

			if got := todo.ParseBodyLine(tt.in); !stringPtrEqual(got, tt.out) {
				t.Errorf("got %q, want %q", derefString(got), derefString(tt.out))
			}
		})
	}
}

func TestTodo_RemoveShouldWork(t *testing.T) {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		log.Fatal(err)
	}

	fileContent := `package main

import "fmt"

// TODO: Rewrite this in rust
//   No really.
func main() {
	fmt.Println("Hello world")
}`
	if _, err := tmp.WriteString(fileContent); err != nil {
		log.Fatal(err)
	}
	tmp.Close()

	wantFileContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello world")
}
`

	todo := Todo{
		Filename: tmp.Name(),
		Prefix:   "TODO",
		Line:     5,
		Body:     []string{""},
	}

	err = todo.Remove()
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadFile(tmp.Name())
	if err != nil {
		log.Fatal(err)
	}

	if got := string(b); got != wantFileContent {
		t.Errorf("got:\n%s\nwant:\n%s", got, wantFileContent)
	}
}

func stringPtrEqual(s1, s2 *string) bool {
	return derefString(s1) == derefString(s2)
}

func stringPtr(s string) *string {
	return &s
}

func derefString(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

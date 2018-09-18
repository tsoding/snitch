package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	// "regexp"
)

type Todo struct {
	Prefix string
	Suffix string
	Id *string
	Filename string
	Line int
}

func (todo Todo) String() string {
	if todo.Id == nil {
		return fmt.Sprintf("%s:%d: %sTODO: %s\n",
			todo.Filename, todo.Line,
			todo.Prefix, todo.Suffix)
	} else {
		return fmt.Sprintf("%s:%d: %sTODO(%s): %s\n",
			todo.Filename, todo.Line,
			todo.Prefix, *todo.Id, todo.Suffix)
	}
}

func ref_str(x string) *string {
	return &x
}

func LineAsTodo(line string) *Todo {
	// return Todo {
	// 	Prefix: "// ",
	// 	Suffix: "khooy",
	// 	Id: ref_str("#42"),
	// 	Filename: "./main.go",
	// 	Line: 10,
	// }
	return nil
}

func TodosOfFile(path string) ([]Todo, error) {
	result := []Todo{}

	file, err := os.Open(path)
	if err != nil {
		return []Todo{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		todo := LineAsTodo(scanner.Text())
		if todo != nil {
			result = append(result, *todo)
		}
	}

	return result, scanner.Err()
}

func TodosOfDir(dirpath string) ([]Todo, error) {
	result := []Todo{}

	err := filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			todos, err := TodosOfFile(path)
			
			if err != nil {
				return err
			}

			for _, todo := range todos {
				result = append(result, todo)
			}
		}

		return nil
	})

	return result, err
}

func ListSubcommand() {
	// TODO: ListSubcommand doesn't handle error from TodosOfDir
	todos, _ := TodosOfDir(".")

	for _, todo := range todos {
		fmt.Printf("%v", todo)
	}
}

func ReportSubcommand() {
	// TODO: ReportSubcommand is not implemented
	panic("report is not implemented")
}

func main() {
	// TODO: index out of range error when no subcommands are provided
	switch os.Args[1] {
	case "list":
		ListSubcommand()
	case "report":
		ReportSubcommand()
	default:
		panic(fmt.Sprintf("`%s` unknown command", os.Args[1]))
	}
}

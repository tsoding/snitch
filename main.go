package main

import "fmt"
import "os"

type Todo struct {
	Prefix string
	Suffix string
	Id *string
	Filename string
	Line int
}

func (todo Todo) String() string {
	// TODO: Todo.String doesn't print id
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

func TodosOfDir(dirpath string) []Todo {
	// TODO: TodosOfDir is not implemented
	return []Todo {
		Todo {
			Prefix: "// ",
			Suffix: "khooy",
			Id: ref_str("#42"),
			Filename: "./main.go",
			Line: 10,
		},
		Todo {
			Prefix: "// ",
			Suffix: "foo",
			Id: nil,
			Filename: "./src/foo.go",
			Line: 0,
		},
	}
}

func ListSubcommand() {
	for _, todo := range TodosOfDir(".") {
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

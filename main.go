package main

import (
	"fmt"
	"os"

	"github.com/jakedgy/shelm/internal/shelm"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("shelm - A REPL for Go text/template with Sprig functions")
		fmt.Println()
		fmt.Println("Usage: shelm")
		fmt.Println()
		fmt.Println("Start an interactive REPL for evaluating template expressions.")
		fmt.Println("Type :help within the REPL for more information.")
		os.Exit(0)
	}

	repl := shelm.NewREPL()
	repl.Run()
}

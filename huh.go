package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var commands = map[string]func([]string) int{
	"echo": echo,
}

func main() {
	if len(os.Args) < 2 {
		errf("%s requires a command name to run!", os.Args[0])
	}

	cmd, ok := commands[os.Args[1]]
	if !ok {
		errf("Unknown command '%s'", os.Args[1])
	}

	os.Exit(cmd(os.Args[2:]))
}

func echo(args []string) int {
	flags, args, err := parseFlags(args, "n")

	if err != nil {
		errf("Unrecognised flag '%s'", err)
	}

	w := bufio.NewWriter(os.Stdout)
	for i, s := range args {
		w.WriteString(s)
		if i != len(args)-1 {
			w.WriteString(" ")
		}
	}
	if !flags["n"] {
		w.WriteString("\n")
	}
	w.Flush()
	return 0
}

func parseFlags(args []string, expected string) (map[string]bool, []string, error) {
	var (
		result  = make(map[string]bool)
		allowed = make(map[rune]bool)
		start   = 0
	)

	for _, v := range expected {
		allowed[v] = true
	}

	for i, s := range args {
		if !strings.HasPrefix(s, "-") {
			break
		}
		start = i + 1
		for idx, r := range []rune(s)[1:] {
			if !allowed[r] {
				return nil, nil, errors.New(string(r))
			}
			result[string(r)] = true
		}
	}

	return result, args[start:], nil
}

func errf(format string, xs ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", xs...)
	os.Exit(1)
}

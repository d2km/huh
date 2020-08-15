package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

var commands = map[string]func([]string) int{
	"echo": echo,
	"wc":   wc,
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
		errf("illegal option '%s'", err)
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

type wcJob struct {
	file *os.File
	name string
}

type wcResult struct {
	chars uint64
	runes uint64
	words uint64
	lines uint64
}

func wc(args []string) int {
	flags, args, err := parseFlags(args, "clmw")

	if err != nil {
		errf("illegal option '%s'", err)
	}

	if len(flags) == 0 {
		flags["c"] = true
		flags["w"] = true
		flags["l"] = true
	}

	jobs := []wcJob{}

	if len(args) == 0 {
		jobs = append(jobs, wcJob{os.Stdin, ""})
	} else {
		for _, name := range args {
			if file, err := os.Open(name); err != nil {
				log.Fatal(err)
			} else {
				jobs = append(jobs, wcJob{file, name})
			}
		}
	}

	channels := make([]chan string, len(jobs))
	for i, job := range jobs {
		c := make(chan string)
		channels[i] = c
		go wcFile(job, flags, c)
	}

	for _, c := range channels {
		fmt.Printf("%s\n", <-c)
	}

	return 0
}

func wcFile(job wcJob, flags map[string]bool, c chan string) {
	var result wcResult

	scanner := bufio.NewScanner(bufio.NewReader(job.file))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		result.lines++
		if flags["c"] {
			result.chars += uint64(len(scanner.Bytes())) + 1 // add line break
		} else if flags["m"] {
			rs := bufio.NewScanner(strings.NewReader(scanner.Text()))
			rs.Split(bufio.ScanRunes)
			for rs.Scan() {
				result.runes++
			}
			result.runes++ // add line break
		}

		if flags["w"] {
			ws := bufio.NewScanner(strings.NewReader(scanner.Text()))
			ws.Split(bufio.ScanWords)
			for ws.Scan() {
				result.words++
			}
		}
	}

	c <- wcPrintStr(&result, job.name, flags)
}

func wcPrintStr(r *wcResult, name string, flags map[string]bool) string {
	var s string

	if flags["l"] {
		s += fmt.Sprintf("%8d", r.lines)
	}
	if flags["w"] {
		s += fmt.Sprintf("%8d", r.words)
	}
	if flags["c"] {
		s += fmt.Sprintf("%8d", r.chars)
	} else if flags["m"] {
		s += fmt.Sprintf("%8d", r.runes)
	}

	s += fmt.Sprintf(" %s", name)

	return s
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
		for _, r := range []rune(s)[1:] {
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

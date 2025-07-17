package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const historySize = 10
var commandHistory []string

func addCommandToHistory(cmd string) {
	if len(commandHistory) >= historySize {
		commandHistory = commandHistory[1:]
	}
	commandHistory = append(commandHistory, cmd)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		// Remove the newline character
		input = strings.TrimSuffix(input, "\n")

		if len(input) == 0 {
			continue
		}

		addCommandToHistory(input)

		// Handle built-in commands
		args := strings.Split(input, " ")
		switch args[0] {
		case "exit":
			return
		case "cd":
			if len(args) < 2 {
				home, err := os.UserHomeDir()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					continue
				}
				err = os.Chdir(home)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			} else {
				err := os.Chdir(args[1])
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}
			continue
		case "history":
			for i, cmd := range commandHistory {
				fmt.Printf("%d  %s\n", i+1, cmd)
			}
			continue
		case "export":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "export: usage: export NAME=VALUE")
				continue
			}
			for _, arg := range args[1:] {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) == 2 {
					os.Setenv(parts[0], parts[1])
				} else {
					fmt.Fprintf(os.Stderr, "export: invalid argument: %s\n", arg)
				}
			}
			continue
		}

		// Expand environment variables in arguments
		for i, arg := range args {
			if strings.HasPrefix(arg, "$") {
				args[i] = os.Getenv(arg[1:])
			}
		}

		// I/O redirection
		var stdin io.Reader = os.Stdin
		var stdout io.Writer = os.Stdout

		if i := strings.Index(input, "<"); i != -1 {
			filename := strings.TrimSpace(input[i+1:])
			file, err := os.Open(filename)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			defer file.Close()
			stdin = file
			input = input[:i]
		}

		if i := strings.Index(input, ">"); i != -1 {
			filename := strings.TrimSpace(input[i+1:])
			file, err := os.Create(filename)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			defer file.Close()
			stdout = file
			input = input[:i]
		}

		// Handle pipes
		if strings.Contains(input, "|") {
			commands := strings.Split(input, "|")
			var cmds []*exec.Cmd

			for _, cmdStr := range commands {
				cmdStr = strings.TrimSpace(cmdStr)
				args := strings.Split(cmdStr, " ")
				cmd := exec.Command(args[0], args[1:]...)
				cmds = append(cmds, cmd)
			}

			for i := 0; i < len(cmds)-1; i++ {
				r, w := io.Pipe()
				cmds[i].Stdout = w
				cmds[i+1].Stdin = r
			}

			cmds[0].Stdin = stdin
			cmds[len(cmds)-1].Stdout = stdout

			for _, cmd := range cmds {
				if err := cmd.Start(); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}

			for _, cmd := range cmds {
				if err := cmd.Wait(); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
\t		}

			continue
		}

		args = strings.Split(input, " ")

		// Check for background process
		isBackground := false
		if len(args) > 0 && args[len(args)-1] == "&" {
			isBackground = true
			args = args[:len(args)-1]
		}

		cmd := exec.Command(args[0], args[1:]...)

		cmd.Stdin = stdin
		cmd.Stdout = stdout
		cmd.Stderr = os.Stderr

		if isBackground {
			err = cmd.Start()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		} else {
			err = cmd.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}
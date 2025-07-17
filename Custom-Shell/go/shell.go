package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

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
			}

			continue
		}

		args = strings.Split(input, " ")
		cmd := exec.Command(args[0], args[1:]...)

		cmd.Stdin = stdin
		cmd.Stdout = stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

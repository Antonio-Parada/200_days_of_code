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

			cmds[len(cmds)-1].Stdout = os.Stdout

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

		cmd := exec.Command(args[0], args[1:]...)

		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		err = cmd.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
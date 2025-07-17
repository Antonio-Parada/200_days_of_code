package main

import (
	"bufio"
	"fmt"
	os"
	os/exec"
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

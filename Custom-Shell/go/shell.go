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

		if input == "exit" {
			break
		}

		args := strings.Split(input, " ")

		cmd := exec.Command(args[0], args[1:]...)

		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		err = cmd.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

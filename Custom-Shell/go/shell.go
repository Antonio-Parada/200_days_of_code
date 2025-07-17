package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const historySize = 10
var commandHistory []string

type Job struct {
	ID      int
	Cmd     *exec.Cmd
	Command string
	Status  string // "Running", "Done", "Stopped"
}

var jobs []Job
var nextJobID int = 1

func addCommandToHistory(cmd string) {
	if len(commandHistory) >= historySize {
		commandHistory = commandHistory[1:]
	}
	commandHistory = append(commandHistory, cmd)
}

func addJob(cmd *exec.Cmd, commandStr string) {
	job := Job{
		ID:      nextJobID,
		Cmd:     cmd,
		Command: commandStr,
		Status:  "Running",
	}
	jobs = append(jobs, job)
	nextJobID++
	fmt.Printf("[%d] %d\n", job.ID, cmd.Process.Pid)
}

func updateJobs() {
	for i := 0; i < len(jobs); i++ {
		select {
		case <-time.After(10 * time.Millisecond):
			// Non-blocking check
			processState := jobs[i].Cmd.ProcessState
			if processState != nil && processState.Exited() {
				fmt.Printf("[%d] Done %s\n", jobs[i].ID, jobs[i].Command)
				jobs = append(jobs[:i], jobs[i+1:]...)
				i-- // Adjust index after removal
			}
		}
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		updateJobs()

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
		case "jobs":
			if len(jobs) == 0 {
				fmt.Println("No background jobs.")
			} else {
				for _, job := range jobs {
					fmt.Printf("[%d] %s %s\n", job.ID, job.Status, job.Command)
				}
			}
			continue
		case "fg":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "fg: usage: fg <job_id>")
				continue
			}
			jobID, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, "fg: invalid job ID")
				continue
			}
			foundJob := -1
			for i, job := range jobs {
				if job.ID == jobID {
					foundJob = i
					break
				}
			}
			if foundJob != -1 {
				job := jobs[foundJob]
				fmt.Printf("Bringing job %d to foreground: %s\n", job.ID, job.Command)
				// Bring process to foreground
				_, err := syscall.Wait4(job.Cmd.Process.Pid, nil, 0, nil)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				// Remove job from list after it finishes
				jobs = append(jobs[:foundJob], jobs[foundJob+1:]...)
			} else {
				fmt.Fprintln(os.Stderr, "fg: job not found")
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
			}

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
			addJob(cmd, input) // Add job to list
		} else {
			err = cmd.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}

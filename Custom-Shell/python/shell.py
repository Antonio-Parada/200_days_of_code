#!/usr/bin/env python3

import os
import sys
import subprocess
import readline
import glob
import atexit
import signal

# List of built-in commands for tab completion
BUILTIN_COMMANDS = ["cd", "exit", "jobs", "fg"]

# List to store background processes
background_jobs = []
job_id_counter = 1

def cleanup_jobs(signum, frame):
    global background_jobs
    for job in background_jobs[:]:
        if job['process'].poll() is not None: # Process has terminated
            print(f"\n[{job['id']}] Done {job['command']}")
            background_jobs.remove(job)

def completer(text, state):
    line = readline.get_line_buffer().split()

    # If no command has been typed yet, suggest built-in commands and executables in PATH
    if not line or (len(line) == 1 and not text):
        options = [cmd for cmd in BUILTIN_COMMANDS if cmd.startswith(text)]
        for path_dir in os.environ.get("PATH", "").split(os.pathsep):
            for exe in glob.glob(os.path.join(path_dir, text + "*")):
                if os.path.isfile(exe) and os.access(exe, os.X_OK):
                    options.append(os.path.basename(exe))
        options = sorted(list(set(options)))
        return options[state] if state < len(options) else None

    # If a command has been typed, suggest file paths
    if len(line) > 0:
        # If the current word is empty, suggest files in the current directory
        if not text:
            options = [f for f in os.listdir(".") if os.path.isdir(f) or os.path.isfile(f)]
            options = sorted(list(set(options)))
            return options[state] if state < len(options) else None
        
        # Otherwise, complete the current word as a file path
        matches = glob.glob(text + "*")
        matches = sorted(list(set(matches)))
        return matches[state] if state < len(matches) else None

    return None

def main():
    """Main loop for the shell."""
    # Set up signal handler for child processes
    signal.signal(signal.SIGCHLD, cleanup_jobs)

    # Set up command history
    histfile = os.path.join(os.path.expanduser("~"), ".python_shell_history")
    try:
        readline.read_history_file(histfile)
        # default history len is -1 (infinite); this avoids memory issues
        readline.set_history_length(1000)
    except FileNotFoundError:
        pass

    atexit.register(readline.write_history_file, histfile)

    # Set up tab completion
    readline.set_completer(completer)
    readline.set_completer_delims(' \t\n`~!@#$%^&*()=+[{]}|;:"',<>/?')
    readline.parse_and_bind("tab: complete")

    while True:
        try:
            # Print the prompt and read a line of input
            line = input("> ").strip()

            # If the line is empty, continue
            if not line:
                continue

            # Exit the shell if the user types 'exit'
            if line == "exit":
                break

            # Split the line into command and arguments
            args = line.split()

            # Handle built-in commands
            if args[0] == "cd":
                if len(args) > 1:
                    try:
                        os.chdir(args[1])
                    except FileNotFoundError:
                        print(f"cd: no such file or directory: {args[1]}")
                else:
                    # cd to home directory if no argument is given
                    os.chdir(os.path.expanduser("~"))
                continue
            elif args[0] == "export":
                if len(args) > 1:
                    for arg in args[1:]:
                        if "=" in arg:
                            key, value = arg.split("=", 1)
                            os.environ[key] = value
                        else:
                            print(f"export: invalid argument: {arg}")
                continue
            elif args[0] == "jobs":
                if not background_jobs:
                    print("No background jobs.")
                else:
                    for job in background_jobs:
                        status = "Running" if job['process'].poll() is None else "Done"
                        print(f"[{job['id']}] {status} {job['command']}")
                continue
            elif args[0] == "fg":
                if len(args) < 2:
                    print("fg: usage: fg <job_id>")
                    continue
                try:
                    job_id = int(args[1])
                    found_job = None
                    for job in background_jobs:
                        if job['id'] == job_id:
                            found_job = job
                            break
                    if found_job:
                        print(f"Bringing job {job_id} to foreground: {found_job['command']}")
                        found_job['process'].wait() # Wait for the process to complete
                        if found_job['process'].poll() is not None: # Process has terminated
                            background_jobs.remove(found_job)
                    else:
                        print(f"fg: job not found: {job_id}")
                except ValueError:
                    print("fg: invalid job ID")
                continue

            # Expand environment variables in arguments
            expanded_args = []
            for arg in args:
                if arg.startswith("$"):
                    expanded_args.append(os.environ.get(arg[1:], ""))
                else:
                    expanded_args.append(arg)
            args = expanded_args

            # Handle I/O redirection
            stdin_redir = None
            stdout_redir = None

            if "<" in args:
                i = args.index("<")
                stdin_redir = open(args[i+1], "r")
                args = args[:i]

            if ">" in args:
                i = args.index(">")
                stdout_redir = open(args[i+1], "w")
                args = args[:i]

            # Handle pipes
            if "|" in line:
                # Split the line into commands
                commands = [cmd.strip().split() for cmd in line.split("|")]
                
                # Create the first process
                p1 = subprocess.Popen(commands[0], stdout=subprocess.PIPE, stdin=stdin_redir)
                
                # Chain the rest of the commands
                for i in range(1, len(commands)):
                    p2 = subprocess.Popen(commands[i], stdin=p1.stdout, stdout=subprocess.PIPE)
                    p1.stdout.close()  # Allow p1 to receive a SIGPIPE if p2 exits.
                    p1 = p2

                # Get the output of the last command
                output, err = p1.communicate()
                if output:
                    if stdout_redir:
                        stdout_redir.write(output.decode())
                        stdout_redir.close()
                    else:
                        print(output.decode().strip())
                if err:
                    print(err.decode().strip(), file=sys.stderr)

            else:
                # Check for background process
                is_background = False
                if args and args[-1] == "&":
                    is_background = True
                    args = args[:-1]

                # Execute a simple command
                if is_background:
                    process = subprocess.Popen(args, stdin=stdin_redir, stdout=stdout_redir, stderr=sys.stderr)
                    global job_id_counter
                    background_jobs.append({'id': job_id_counter, 'process': process, 'command': ' '.join(args)})
                    print(f"[{job_id_counter}] {process.pid}")
                    job_id_counter += 1
                else:
                    subprocess.run(args, stdin=stdin_redir, stdout=stdout_redir, stderr=sys.stderr)

            if stdin_redir:
                stdin_redir.close()
            if stdout_redir:
                stdout_redir.close()

        except FileNotFoundError:
            print(f"Command not found: {line.split()[0]}")
        except Exception as e:
            print(f"An error occurred: {e}")

if __name__ == "__main__":
    main()

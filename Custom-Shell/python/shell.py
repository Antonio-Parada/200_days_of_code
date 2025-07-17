#!/usr/bin/env python3

import os
import sys
import subprocess

def main():
    """Main loop for the shell."""
    while True:
        # Print the prompt
        sys.stdout.write("> ")
        sys.stdout.flush()

        # Read a line of input
        line = sys.stdin.readline().strip()

        # If the line is empty, continue
        if not line:
            continue

        # Exit the shell if the user types 'exit'
        if line == "exit":
            break

        # Split the line into command and arguments
        args = line.split()

        try:
            # Handle pipes
            if "|" in line:
                # Split the line into commands
                commands = [cmd.strip().split() for cmd in line.split("|")]
                
                # Create the first process
                p1 = subprocess.Popen(commands[0], stdout=subprocess.PIPE)
                
                # Chain the rest of the commands
                for i in range(1, len(commands)):
                    p2 = subprocess.Popen(commands[i], stdin=p1.stdout, stdout=subprocess.PIPE)
                    p1.stdout.close()  # Allow p1 to receive a SIGPIPE if p2 exits.
                    p1 = p2

                # Get the output of the last command
                output, err = p1.communicate()
                if output:
                    print(output.decode().strip())
                if err:
                    print(err.decode().strip(), file=sys.stderr)

            else:
                # Execute a simple command
                args = line.split()
                # Handle the 'cd' command
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
                
                subprocess.run(args)

        except FileNotFoundError:
            print(f"Command not found: {line.split()[0]}")
        except Exception as e:
            print(f"An error occurred: {e}")

if __name__ == "__main__":
    main()

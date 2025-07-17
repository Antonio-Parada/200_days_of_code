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

        try:
            # Execute the command
            subprocess.run(args)
        except FileNotFoundError:
            print(f"Command not found: {args[0]}")
        except Exception as e:
            print(f"An error occurred: {e}")

if __name__ == "__main__":
    main()

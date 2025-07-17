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
            # Execute the command
            subprocess.run(args)
        except FileNotFoundError:
            print(f"Command not found: {args[0]}")
        except Exception as e:
            print(f"An error occurred: {e}")

if __name__ == "__main__":
    main()

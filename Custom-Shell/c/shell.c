#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>

#define MAX_LINE 80 /* The maximum length command */

int main(void)
{
    char *args[MAX_LINE/2 + 1]; /* command line arguments */
    int should_run = 1; /* flag to determine when to exit program */

    while (should_run) {
        printf("> ");
        fflush(stdout);

        char line[MAX_LINE];
        fgets(line, MAX_LINE, stdin);

        // Remove trailing newline character
        line[strcspn(line, "\n")] = 0;

        if (strlen(line) == 0) {
            continue;
        }

        // Tokenize the input string
        char *token = strtok(line, " ");
        int i = 0;
        while (token != NULL) {
            args[i++] = token;
            token = strtok(NULL, " ");
        }
        args[i] = NULL;

        if (args[0] == NULL) {
            continue;
        }

        if (strcmp(args[0], "exit") == 0) {
            should_run = 0;
            continue;
        }

        if (strcmp(args[0], "cd") == 0) {
            if (args[1] == NULL) {
                char *home = getenv("HOME");
                if (home != NULL) {
                    if (chdir(home) != 0) {
                        perror("cd");
                    }
                }
            } else {
                if (chdir(args[1]) != 0) {
                    perror("cd");
                }
            }
            continue;
        }

        pid_t pid = fork();

        if (pid < 0) { // Error
            fprintf(stderr, "Fork failed\n");
            return 1;
        } else if (pid == 0) { // Child process
            execvp(args[0], args);
            // execvp only returns if there is an error
            perror("execvp");
            exit(EXIT_FAILURE);
        } else { // Parent process
            wait(NULL);
        }
    }

    return 0;
}

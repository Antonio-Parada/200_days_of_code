#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>
#include <fcntl.h>

#define MAX_LINE 80 /* The maximum length command */

// Function to parse a command string into arguments
void parse_command(char* command, char** args) {
    char* token = strtok(command, " ");
    int i = 0;
    while (token != NULL) {
        args[i++] = token;
        token = strtok(NULL, " ");
    }
    args[i] = NULL;
}

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

        if (strcmp(line, "exit") == 0) {
            should_run = 0;
            continue;
        }

        // I/O redirection
        char* input_file = NULL;
        char* output_file = NULL;
        char* redirect_pos = strchr(line, '<');
        if (redirect_pos != NULL) {
            *redirect_pos = '\0';
            input_file = strtok(redirect_pos + 1, " ");
        }

        redirect_pos = strchr(line, '>');
        if (redirect_pos != NULL) {
            *redirect_pos = '\0';
            output_file = strtok(redirect_pos + 1, " ");
        }

        // Check for pipes
        char* pipe_pos = strchr(line, '|');
        if (pipe_pos != NULL) {
            // Pipe found, handle it
            char* command1 = strtok(line, "|");
            char* command2 = strtok(NULL, "|");

            char* args1[MAX_LINE/2 + 1];
            char* args2[MAX_LINE/2 + 1];

            parse_command(command1, args1);
            parse_command(command2, args2);

            int pipefd[2];
            pid_t p1, p2;

            if (pipe(pipefd) < 0) {
                perror("pipe");
                continue;
            }

            p1 = fork();
            if (p1 < 0) {
                perror("fork");
                continue;
            }

            if (p1 == 0) { // Child 1
                close(pipefd[0]);
                dup2(pipefd[1], STDOUT_FILENO);
                close(pipefd[1]);
                execvp(args1[0], args1);
                perror("execvp");
                exit(EXIT_FAILURE);
            }

            p2 = fork();
            if (p2 < 0) {
                perror("fork");
                continue;
            }

            if (p2 == 0) { // Child 2
                close(pipefd[1]);
                dup2(pipefd[0], STDIN_FILENO);
                close(pipefd[0]);
                execvp(args2[0], args2);
                perror("execvp");
                exit(EXIT_FAILURE);
            }

            close(pipefd[0]);
            close(pipefd[1]);
            waitpid(p1, NULL, 0);
            waitpid(p2, NULL, 0);

        } else {
            // No pipe, execute a simple command
            parse_command(line, args);

            if (args[0] == NULL) {
                continue;
            }

            // Check for background process
            int is_background = 0;
            int arg_count = 0;
            while (args[arg_count] != NULL) {
                arg_count++;
            }
            if (arg_count > 0 && strcmp(args[arg_count - 1], "&") == 0) {
                is_background = 1;
                args[arg_count - 1] = NULL; // Remove the '&'
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
                if (input_file != NULL) {
                    int fd_in = open(input_file, O_RDONLY);
                    if (fd_in < 0) {
                        perror("open");
                        exit(EXIT_FAILURE);
                    }
                    dup2(fd_in, STDIN_FILENO);
                    close(fd_in);
                }
                if (output_file != NULL) {
                    int fd_out = open(output_file, O_WRONLY | O_CREAT | O_TRUNC, 0644);
                    if (fd_out < 0) {
                        perror("open");
                        exit(EXIT_FAILURE);
                    }
                    dup2(fd_out, STDOUT_FILENO);
                    close(fd_out);
                }

                execvp(args[0], args);
                // execvp only returns if there is an error
                perror("execvp");
                exit(EXIT_FAILURE);
            } else { // Parent process
                if (!is_background) {
                    wait(NULL);
                }
            }
        }
    }

    return 0;
}

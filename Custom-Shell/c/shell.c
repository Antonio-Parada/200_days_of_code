#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>
#include <fcntl.h>
#include <termios.h>
#include <errno.h>
#include <dirent.h>

#define MAX_LINE 80 /* The maximum length command */
#define HISTORY_SIZE 10

char history[HISTORY_SIZE][MAX_LINE];
int history_count = 0;
int history_index = 0;

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

// Function to add a command to history
void add_to_history(const char* command) {
    if (history_count < HISTORY_SIZE) {
        strcpy(history[history_count], command);
        history_count++;
    } else {
        for (int i = 0; i < HISTORY_SIZE - 1; i++) {
            strcpy(history[i], history[i+1]);
        }
        strcpy(history[HISTORY_SIZE - 1], command);
    }
    history_index = history_count;
}

int main(void)
{
    char *args[MAX_LINE/2 + 1]; /* command line arguments */
    int should_run = 1; /* flag to determine when to exit program */

    struct termios old_tio, new_tio;

    // Get the current terminal settings
    tcgetattr(STDIN_FILENO, &old_tio);

    // Set new terminal settings for raw mode
    new_tio = old_tio;
    new_tio.c_lflag &= (~ICANON & ~ECHO);
    tcsetattr(STDIN_FILENO, TCSANOW, &new_tio);

    while (should_run) {
        printf("> ");
        fflush(stdout);

        char line[MAX_LINE];
        memset(line, 0, MAX_LINE);
        int line_len = 0;
        char c;

        while (read(STDIN_FILENO, &c, 1) == 1) {
            if (c == '\n') { // Enter key
                break;
            } else if (c == 127) { // Backspace
                if (line_len > 0) {
                    line_len--;
                    printf("\b \b"); // Erase character from screen
                    fflush(stdout);
                }
            } else if (c == '\t') { // Tab key
                // Basic tab completion for files/directories in current dir
                DIR *d;
                struct dirent *dir;
                char current_word[MAX_LINE];
                int word_start = 0;
                for (int i = line_len - 1; i >= 0; i--) {
                    if (line[i] == ' ' || line[i] == '/') {
                        word_start = i + 1;
                        break;
                    }
                }
                strncpy(current_word, line + word_start, line_len - word_start);
                current_word[line_len - word_start] = '\0';

                char *matches[MAX_LINE];
                int match_count = 0;

                d = opendir(".");
                if (d) {
                    while ((dir = readdir(d)) != NULL) {
                        if (strncmp(dir->d_name, current_word, strlen(current_word)) == 0) {
                            matches[match_count++] = strdup(dir->d_name);
                        }
                    }
                    closedir(d);
                }

                if (match_count == 1) {
                    // Complete the word
                    for (int i = 0; i < line_len - word_start; i++) printf("\b \b"); // Erase current word
                    printf("%s", matches[0]);
                    strcpy(line + word_start, matches[0]);
                    line_len = strlen(line);
                    fflush(stdout);
                } else if (match_count > 1) {
                    // Print matches and redraw prompt
                    printf("\n");
                    for (int i = 0; i < match_count; i++) {
                        printf("%s  ", matches[i]);
                    }
                    printf("\n> %s", line);
                    fflush(stdout);
                }

                for (int i = 0; i < match_count; i++) {
                    free(matches[i]);
                }

            } else if (c == '\033') { // Escape sequence (arrow keys)
                read(STDIN_FILENO, &c, 1);
                read(STDIN_FILENO, &c, 1);
                if (c == 'A') { // Up arrow
                    if (history_index > 0) {
                        history_index--;
                        // Clear current line
                        for (int i = 0; i < line_len; i++) printf("\b \b");
                        printf("%s", history[history_index]);
                        strcpy(line, history[history_index]);
                        line_len = strlen(line);
                        fflush(stdout);
                    }
                } else if (c == 'B') { // Down arrow
                    if (history_index < history_count - 1) {
                        history_index++;
                        // Clear current line
                        for (int i = 0; i < line_len; i++) printf("\b \b");
                        printf("%s", history[history_index]);
                        strcpy(line, history[history_index]);
                        line_len = strlen(line);
                        fflush(stdout);
                    } else if (history_index == history_count - 1) {
                        history_index++;
                        // Clear current line
                        for (int i = 0; i < line_len; i++) printf("\b \b");
                        memset(line, 0, MAX_LINE);
                        line_len = 0;
                        fflush(stdout);
                    }
                }
            } else {
                if (line_len < MAX_LINE - 1) {
                    line[line_len++] = c;
                    printf("%c", c);
                    fflush(stdout);
                }
            }
        }
        line[line_len] = '\0'; // Null-terminate the string

        if (strlen(line) == 0) {
            continue;
        }

        add_to_history(line);

        if (strcmp(line, "exit") == 0) {
            should_run = 0;
            continue;
        }

        // I/O redirection
        char temp_line[MAX_LINE];
        strcpy(temp_line, line);
        char* input_file = NULL;
        char* output_file = NULL;
        char* redirect_pos = strchr(temp_line, '<');
        if (redirect_pos != NULL) {
            *redirect_pos = '\0';
            input_file = strtok(redirect_pos + 1, " ");
        }

        redirect_pos = strchr(temp_line, '>');
        if (redirect_pos != NULL) {
            *redirect_pos = '\0';
            output_file = strtok(redirect_pos + 1, " ");
        }

        // Check for pipes
        char* pipe_pos = strchr(temp_line, '|');
        if (pipe_pos != NULL) {
            // Pipe found, handle it
            char* command1 = strtok(temp_line, "|");
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
            parse_command(temp_line, args);

            if (args[0] == NULL) {
                continue;
            }

            // Handle 'export' command
            if (strcmp(args[0], "export") == 0) {
                if (args[1] != NULL) {
                    char *eq_pos = strchr(args[1], '=');
                    if (eq_pos != NULL) {
                        *eq_pos = '\0';
                        setenv(args[1], eq_pos + 1, 1);
                    } else {
                        fprintf(stderr, "export: usage: export NAME=VALUE\n");
                    }
                }
                continue;
            }

            // Expand environment variables
            for (int j = 0; args[j] != NULL; j++) {
                if (args[j][0] == '

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

    // Restore original terminal settings
    tcsetattr(STDIN_FILENO, TCSANOW, &old_tio);

    return 0;
}
) {
                    char *env_val = getenv(args[j] + 1);
                    if (env_val != NULL) {
                        args[j] = env_val;
                    } else {
                        args[j] = ""; // Replace with empty string if not found
                    }
                }
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

    // Restore original terminal settings
    tcsetattr(STDIN_FILENO, TCSANOW, &old_tio);

    return 0;
}

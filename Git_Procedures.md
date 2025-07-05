# Git Procedures for 200 Days of Code

This document outlines the essential Git commands and workflow for managing your 200 Days of Code projects.

## 1. Cloning the Repository

If you haven't already, clone the main repository to your local machine:

```bash
git clone <repository_url>
cd 200_days_of_code
```

## 2. Checking Status

Before making changes or committing, always check the status of your repository:

```bash
git status
```

## 3. Staging Changes

After making changes to your project files, you need to stage them for commit. This adds them to the staging area:

```bash
git add .
```
Or to stage specific files:
```bash
git add path/to/your/file.md
```

## 4. Committing Changes

Once your changes are staged, commit them with a descriptive message:

```bash
git commit -m "Completed [Project Name]: Brief description of changes"
```

## 5. Pushing Changes

To upload your local commits to the remote repository (e.g., GitHub):

```bash
git push
```

## 6. Pulling Latest Changes

Before starting new work, always pull the latest changes from the remote repository to avoid merge conflicts:

```bash
git pull
```

## 7. Branching (Optional but Recommended)

For larger projects or when experimenting, consider creating new branches:

```bash
git checkout -b feature/my-new-feature
```

Switch back to main:

```bash
git checkout main
```

Merge a branch:

```bash
git merge feature/my-new-feature
```

## 8. Resolving Merge Conflicts

If `git pull` results in conflicts, you'll need to manually resolve them. Git will mark the conflicted files. Edit them to resolve the conflicts, then `git add` and `git commit`.

## 9. Viewing History

To see a log of your commits:

```bash
git log
```

#!/bin/bash
# Simple push to current branch
set -e

# Get current branch
branch=$(git rev-parse --abbrev-ref HEAD)

# Check for uncommitted changes
status=$(git status --porcelain)
if [ -z "$status" ]; then
    echo -e "\033[33mNo changes to commit.\033[0m"
    exit 0
fi

# Show status
echo -e "\033[36mChanges to commit:\033[0m"
git status --short

echo ""
read -p "Commit message: " message

if [ -z "$message" ]; then
    echo -e "\033[31mAborted: No commit message provided.\033[0m"
    exit 1
fi

# Commit and push
echo ""
echo -e "\033[36mCommitting and pushing to $branch...\033[0m"

git add .
git commit -m "$message"

if [ $? -ne 0 ]; then
    echo -e "\033[31mCommit failed.\033[0m"
    exit 1
fi

git push origin "$branch"

if [ $? -eq 0 ]; then
    echo -e "\033[32mPushed to $branch successfully!\033[0m"
else
    echo -e "\033[31mPush failed.\033[0m"
    exit 1
fi

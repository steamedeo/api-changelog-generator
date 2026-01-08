#!/bin/bash
# Create and push a release tag
set -e

# Get current branch
branch=$(git rev-parse --abbrev-ref HEAD)

# Check for uncommitted changes
status=$(git status --porcelain)
skip_commit=false

if [ -z "$status" ]; then
    echo -e "\033[33mNo changes to commit.\033[0m"
    read -p "Tag current commit as release? (y/N): " tag_only
    if [ "$tag_only" != "y" ] && [ "$tag_only" != "Y" ]; then
        exit 0
    fi
    skip_commit=true
else
    # Show status
    echo -e "\033[36mChanges to commit:\033[0m"
    git status --short
    echo ""
fi

# Get current version (suppress error output)
current_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [ -z "$current_version" ]; then
    echo -e "\033[33mNo existing tags found.\033[0m"
    current_version="v0.0.0"
else
    echo -e "\033[36mCurrent version: $current_version\033[0m"
fi

# Parse semantic version
if [[ $current_version =~ ^v?([0-9]+)\.([0-9]+)\.([0-9]+) ]]; then
    major="${BASH_REMATCH[1]}"
    minor="${BASH_REMATCH[2]}"
    patch="${BASH_REMATCH[3]}"
else
    major=0
    minor=1
    patch=0
fi

# Suggest next versions
echo ""
echo -e "\033[36mSuggested versions:\033[0m"
echo "  1) v$major.$minor.$((patch + 1))  (patch - bug fixes)"
echo "  2) v$major.$((minor + 1)).0  (minor - new features)"
echo "  3) v$((major + 1)).0.0  (major - breaking changes)"
echo "  4) Custom version"
echo ""

read -p "Select version (1-4): " choice

case $choice in
    1)
        new_version="v$major.$minor.$((patch + 1))"
        ;;
    2)
        new_version="v$major.$((minor + 1)).0"
        ;;
    3)
        new_version="v$((major + 1)).0.0"
        ;;
    4)
        read -p "Enter version (e.g., v1.2.3): " new_version
        if [[ ! $new_version =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo -e "\033[31mInvalid version format. Use: v1.2.3\033[0m"
            exit 1
        fi
        if [[ ! $new_version =~ ^v ]]; then
            new_version="v$new_version"
        fi
        ;;
    *)
        echo -e "\033[31mInvalid choice.\033[0m"
        exit 1
        ;;
esac

# Get commit/release message
echo ""
if [ "$skip_commit" = false ]; then
    read -p "Commit message: " message
    if [ -z "$message" ]; then
        echo -e "\033[31mAborted: No commit message provided.\033[0m"
        exit 1
    fi
fi

# Confirm
echo ""
echo -e "\033[33mCreating release: $new_version\033[0m"
if [ "$skip_commit" = false ]; then
    echo -e "  \033[90mCommit message: $message\033[0m"
fi
echo ""
read -p "Continue? (y/N): " confirm
if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo -e "\033[33mAborted.\033[0m"
    exit 0
fi

# Commit if there are changes
if [ "$skip_commit" = false ]; then
    echo ""
    echo -e "\033[36mCommitting changes...\033[0m"
    git add .
    git commit -m "$message"

    if [ $? -ne 0 ]; then
        echo -e "\033[31mCommit failed.\033[0m"
        exit 1
    fi
fi

# Create tag
echo -e "\033[36mCreating tag $new_version...\033[0m"
if [ "$skip_commit" = true ]; then
    tag_message="Release $new_version"
else
    tag_message="Release $new_version: $message"
fi
git tag -a "$new_version" -m "$tag_message"

if [ $? -ne 0 ]; then
    echo -e "\033[31mTag creation failed.\033[0m"
    exit 1
fi

# Push
echo -e "\033[36mPushing to $branch...\033[0m"
git push origin "$branch"

if [ $? -ne 0 ]; then
    echo -e "\033[31mPush failed.\033[0m"
    exit 1
fi

echo -e "\033[36mPushing tag $new_version...\033[0m"
git push origin "$new_version"

if [ $? -eq 0 ]; then
    echo ""
    echo -e "\033[32mRelease $new_version created and pushed successfully!\033[0m"
else
    echo -e "\033[31mTag push failed.\033[0m"
    exit 1
fi

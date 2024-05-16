#!/bin/bash

# Function to compare versions
function version_gt() {
    test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"
}

# Check if there are any uncommitted changes
if [[ $(git status --porcelain) ]]; then
    echo "Error: There are uncommitted changes. Please commit or stash them before updating the version."
    git status --short
    exit 1
fi

# Get the current version from the version file
current_version=$(cat version)

# Prompt the user for a new version
read -p "Current version is $current_version. Enter new version: " new_version

# Ensure the new version is larger than the current version
while ! version_gt "$new_version" "$current_version"; do
    echo "New version must be greater than $current_version"
    read -p "Enter new version: " new_version
done

# Update the tests/version file with the new version
echo "$new_version" > version

# Stage the modified tests/version file
git add version

# Commit the changes with a message including the new version
git commit -m "Bump version to $new_version"

# Create a tag with the new version
git tag "v$new_version"

# Push the changes to the origin/head
git push origin HEAD

# Push the new tag to the origin
git push origin "v$new_version"

echo "Version updated to $new_version, committed, tagged, and pushed to origin/head."
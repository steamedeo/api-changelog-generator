# Trunk-Based Development Release Tagger
# This script tags the current commit on main/master as a release

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Error { Write-Host $args -ForegroundColor Red }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }

# Check if we're on main/master
$currentBranch = git rev-parse --abbrev-ref HEAD
if ($currentBranch -ne "main" -and $currentBranch -ne "master") {
    Write-Warning "Warning: You are not on main/master branch (current: $currentBranch)"
    $continue = Read-Host "Do you want to continue? (y/N)"
    if ($continue -ne "y" -and $continue -ne "Y") {
        Write-Info "Aborted."
        exit 0
    }
}

# Check for uncommitted changes
$status = git status --porcelain
if ($status) {
    Write-Error "Error: You have uncommitted changes. Please commit or stash them first."
    Write-Host ""
    Write-Host "Uncommitted changes:"
    git status --short
    exit 1
}

# Check if we're up to date with remote
Write-Info "Checking remote status..."
git fetch origin 2>&1 | Out-Null

$localCommit = git rev-parse HEAD
$remoteCommit = git rev-parse "origin/$currentBranch" 2>$null

if ($remoteCommit -and $localCommit -ne $remoteCommit) {
    Write-Error "Error: Your branch is not in sync with origin/$currentBranch"
    Write-Host ""
    Write-Host "Run 'git pull' to sync or 'git push' to push your changes first."
    exit 1
}

# Get current version from git tags
$currentVersion = git describe --tags --abbrev=0 2>$null
if (-not $currentVersion) {
    Write-Info "No existing tags found. This will be your first release."
    $currentVersion = "v0.0.0"
} else {
    Write-Info "Current version: $currentVersion"
}

# Parse semantic version
if ($currentVersion -match '^v?(\d+)\.(\d+)\.(\d+)') {
    $major = [int]$matches[1]
    $minor = [int]$matches[2]
    $patch = [int]$matches[3]
} else {
    Write-Warning "Could not parse version, starting from v0.1.0"
    $major = 0
    $minor = 1
    $patch = 0
}

# Suggest next versions
Write-Host ""
Write-Info "Suggested versions:"
Write-Host "  1) v$($major).$($minor).$($patch + 1)  (patch - bug fixes)"
Write-Host "  2) v$($major).$($minor + 1).0  (minor - new features)"
Write-Host "  3) v$($major + 1).0.0  (major - breaking changes)"
Write-Host "  4) Custom version"
Write-Host ""

$choice = Read-Host "Select version type (1-4)"

switch ($choice) {
    "1" { $newVersion = "v$($major).$($minor).$($patch + 1)" }
    "2" { $newVersion = "v$($major).$($minor + 1).0" }
    "3" { $newVersion = "v$($major + 1).0.0" }
    "4" {
        $newVersion = Read-Host "Enter custom version (e.g., v1.2.3)"
        if ($newVersion -notmatch '^v?\d+\.\d+\.\d+') {
            Write-Error "Invalid version format. Use format: v1.2.3"
            exit 1
        }
        if (-not $newVersion.StartsWith("v")) {
            $newVersion = "v$newVersion"
        }
    }
    default {
        Write-Error "Invalid choice."
        exit 1
    }
}

# Confirm
Write-Host ""
Write-Warning "About to create release tag: $newVersion"
Write-Host "  Branch: $currentBranch"
Write-Host "  Commit: $(git rev-parse --short HEAD)"
Write-Host ""

$confirm = Read-Host "Continue? (y/N)"
if ($confirm -ne "y" -and $confirm -ne "Y") {
    Write-Info "Aborted."
    exit 0
}

# Get release notes
Write-Host ""
Write-Info "Enter release notes (optional, press Enter twice when done):"
$releaseNotes = @()
do {
    $line = Read-Host
    if ($line) {
        $releaseNotes += $line
    }
} while ($line)

$releaseMessage = if ($releaseNotes.Count -gt 0) {
    "Release $newVersion`n`n" + ($releaseNotes -join "`n")
} else {
    "Release $newVersion"
}

# Create the tag
Write-Info "Creating tag $newVersion..."
git tag -a $newVersion -m $releaseMessage

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to create tag."
    exit 1
}

Write-Success "Tag $newVersion created successfully!"

# Ask to push
Write-Host ""
$pushTag = Read-Host "Push tag to remote? (Y/n)"
if ($pushTag -ne "n" -and $pushTag -ne "N") {
    Write-Info "Pushing tag to remote..."
    git push origin $newVersion

    if ($LASTEXITCODE -eq 0) {
        Write-Success "Tag pushed to remote successfully!"
        Write-Host ""
        Write-Success "Release $newVersion complete!"
        Write-Host ""
        Write-Info "Next steps:"
        Write-Host "  - CI/CD should automatically build and deploy this release"
        Write-Host "  - Check your CI/CD pipeline for build status"
        Write-Host "  - Create release notes on GitHub/GitLab if needed"
    } else {
        Write-Error "Failed to push tag. You can push manually with: git push origin $newVersion"
        exit 1
    }
} else {
    Write-Info "Tag created locally. Push it later with: git push origin $newVersion"
}

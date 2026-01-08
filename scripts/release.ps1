# Create and push a release tag

# Get current branch
$branch = git rev-parse --abbrev-ref HEAD

# Check for uncommitted changes
$status = git status --porcelain
if (-not $status) {
    Write-Host "No changes to commit." -ForegroundColor Yellow
    $tagOnly = Read-Host "Tag current commit as release? (y/N)"
    if ($tagOnly -ne "y" -and $tagOnly -ne "Y") {
        exit 0
    }
    $skipCommit = $true
} else {
    # Show status
    Write-Host "Changes to commit:" -ForegroundColor Cyan
    git status --short
    Write-Host ""
    $skipCommit = $false
}

# Get current version (suppress error output)
$ErrorActionPreference = "SilentlyContinue"
$currentVersion = git describe --tags --abbrev=0 2>$null
$ErrorActionPreference = "Stop"

if (-not $currentVersion -or $LASTEXITCODE -ne 0) {
    Write-Host "No existing tags found." -ForegroundColor Yellow
    $currentVersion = "v0.0.0"
} else {
    Write-Host "Current version: $currentVersion" -ForegroundColor Cyan
}

# Parse semantic version
if ($currentVersion -match '^v?(\d+)\.(\d+)\.(\d+)') {
    $major = [int]$matches[1]
    $minor = [int]$matches[2]
    $patch = [int]$matches[3]
} else {
    $major = 0
    $minor = 1
    $patch = 0
}

# Suggest next versions
Write-Host ""
Write-Host "Suggested versions:" -ForegroundColor Cyan
Write-Host "  1) v$($major).$($minor).$($patch + 1)  (patch - bug fixes)"
Write-Host "  2) v$($major).$($minor + 1).0  (minor - new features)"
Write-Host "  3) v$($major + 1).0.0  (major - breaking changes)"
Write-Host "  4) Custom version"
Write-Host ""

$choice = Read-Host "Select version (1-4)"

switch ($choice) {
    "1" { $newVersion = "v$($major).$($minor).$($patch + 1)" }
    "2" { $newVersion = "v$($major).$($minor + 1).0" }
    "3" { $newVersion = "v$($major + 1).0.0" }
    "4" {
        $newVersion = Read-Host "Enter version (e.g., v1.2.3)"
        if ($newVersion -notmatch '^v?\d+\.\d+\.\d+') {
            Write-Host "Invalid version format. Use: v1.2.3" -ForegroundColor Red
            exit 1
        }
        if (-not $newVersion.StartsWith("v")) {
            $newVersion = "v$newVersion"
        }
    }
    default {
        Write-Host "Invalid choice." -ForegroundColor Red
        exit 1
    }
}

# Get commit/release message
Write-Host ""
if (-not $skipCommit) {
    $message = Read-Host "Commit message"
    if (-not $message) {
        Write-Host "Aborted: No commit message provided." -ForegroundColor Red
        exit 1
    }
}

# Confirm
Write-Host ""
Write-Host "Creating release: $newVersion" -ForegroundColor Yellow
if (-not $skipCommit) {
    Write-Host "  Commit message: $message" -ForegroundColor Gray
}
Write-Host ""
$confirm = Read-Host "Continue? (y/N)"
if ($confirm -ne "y" -and $confirm -ne "Y") {
    Write-Host "Aborted." -ForegroundColor Yellow
    exit 0
}

# Commit if there are changes
if (-not $skipCommit) {
    Write-Host ""
    Write-Host "Committing changes..." -ForegroundColor Cyan
    git add .
    git commit -m $message

    if ($LASTEXITCODE -ne 0) {
        Write-Host "Commit failed." -ForegroundColor Red
        exit 1
    }
}

# Create tag
Write-Host "Creating tag $newVersion..." -ForegroundColor Cyan
$tagMessage = if ($skipCommit) { "Release $newVersion" } else { "Release $newVersion`: $message" }
git tag -a $newVersion -m $tagMessage

if ($LASTEXITCODE -ne 0) {
    Write-Host "Tag creation failed." -ForegroundColor Red
    exit 1
}

# Push
Write-Host "Pushing to $branch..." -ForegroundColor Cyan
git push origin $branch

if ($LASTEXITCODE -ne 0) {
    Write-Host "Push failed." -ForegroundColor Red
    exit 1
}

Write-Host "Pushing tag $newVersion..." -ForegroundColor Cyan
git push origin $newVersion

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Release $newVersion created and pushed successfully!" -ForegroundColor Green
} else {
    Write-Host "Tag push failed." -ForegroundColor Red
    exit 1
}

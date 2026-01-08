# Simple push to current branch
$ErrorActionPreference = "Stop"

# Get current branch
$branch = git rev-parse --abbrev-ref HEAD

# Check for uncommitted changes
$status = git status --porcelain
if (-not $status) {
    Write-Host "No changes to commit." -ForegroundColor Yellow
    exit 0
}

# Show status
Write-Host "Changes to commit:" -ForegroundColor Cyan
git status --short

Write-Host ""
$message = Read-Host "Commit message"

if (-not $message) {
    Write-Host "Aborted: No commit message provided." -ForegroundColor Red
    exit 1
}

# Commit and push
Write-Host ""
Write-Host "Committing and pushing to $branch..." -ForegroundColor Cyan

git add .
git commit -m $message

if ($LASTEXITCODE -ne 0) {
    Write-Host "Commit failed." -ForegroundColor Red
    exit 1
}

git push origin $branch

if ($LASTEXITCODE -eq 0) {
    Write-Host "Pushed to $branch successfully!" -ForegroundColor Green
} else {
    Write-Host "Push failed." -ForegroundColor Red
    exit 1
}

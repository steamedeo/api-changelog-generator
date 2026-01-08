# Quick Git Push Helper
Write-Host ""

# Get commit message
$commitMsg = Read-Host "Commit message"
if ([string]::IsNullOrWhiteSpace($commitMsg)) {
    Write-Host "❌ Commit message cannot be empty!" -ForegroundColor Red
    exit 1
}

# Stage, commit, and push
git add .
git commit -m $commitMsg
git push

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Pushed!" -ForegroundColor Green
} else {
    Write-Host "❌ Push failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""

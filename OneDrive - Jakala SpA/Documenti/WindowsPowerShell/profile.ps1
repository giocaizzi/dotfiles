# Load secrets (untracked, user-managed)
$secretsFile = Join-Path $HOME ".secrets.ps1"
if (Test-Path $secretsFile) {
    . $secretsFile
}

# -----------------------
# oh-my-posh settings
oh-my-posh init pwsh --config "~\.mytheme.omp.json" | Invoke-Expression

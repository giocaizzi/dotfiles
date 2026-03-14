$ErrorActionPreference = "Stop"

# Sync AI assistant repositories into ~/.ai-assistants/<repo>.
# This script runs after chezmoi apply and is idempotent:
# - pull when a repo already exists
# - clone when a repo is missing
# - create ~/.claude/skills junction if it doesn't exist

$baseDir = Join-Path $HOME ".ai-assistants"
$legacyBaseDir = Join-Path $HOME "ai-assistants"

$skillsUrl = "git@github.com:giocaizzi/skills.git"
$skillsDir = Join-Path $baseDir "skills"

$ralphUrl = "git@github.com:giocaizzi/ralph-copilot.git"
$ralphDir = Join-Path $baseDir "ralph-copilot"

$copilotAgentsUrl = "git@github.com:giocaizzi/copilot-agents.git"
$copilotAgentsDir = Join-Path $baseDir "copilot-agents"

# Keep each repository up to date without rewriting non-git directories.
function Sync-Repo {
    param(
        [Parameter(Mandatory = $true)]
        [string]$RepoUrl,
        [Parameter(Mandatory = $true)]
        [string]$RepoDir
    )

    if (Test-Path (Join-Path $RepoDir ".git")) {
        git -C $RepoDir pull --ff-only --no-tags --no-rebase
    } elseif (-not (Test-Path $RepoDir)) {
        git clone $RepoUrl $RepoDir
    } else {
        Write-Warning "Skipping clone: $RepoDir exists but is not a git repository"
    }
}

# TODO: remove this migration block after all Windows machines use ~/.ai-assistants.
# Migrate from the legacy visible folder name once.
# If legacy folder is absent, proceed as a fresh start.
if (-not (Test-Path $baseDir) -and (Test-Path $legacyBaseDir -PathType Container)) {
    Move-Item -Path $legacyBaseDir -Destination $baseDir
}

# Ensure the base folder exists, then sync all managed repositories.
if (-not (Test-Path $baseDir)) {
    New-Item -ItemType Directory -Path $baseDir | Out-Null
}

Sync-Repo -RepoUrl $skillsUrl -RepoDir $skillsDir
Sync-Repo -RepoUrl $ralphUrl -RepoDir $ralphDir
Sync-Repo -RepoUrl $copilotAgentsUrl -RepoDir $copilotAgentsDir

$claudeDir = Join-Path $HOME ".claude"
$skillsLink = Join-Path $claudeDir "skills"
$skillsTarget = Join-Path $skillsDir "skills"

# Expose skills to Claude at ~/.claude/skills when available.
if (-not (Test-Path $claudeDir)) {
    New-Item -ItemType Directory -Path $claudeDir | Out-Null
}

if (Test-Path $skillsTarget) {
    $existingLink = Get-Item -LiteralPath $skillsLink -Force -ErrorAction SilentlyContinue

    if ($null -eq $existingLink) {
        New-Item -ItemType Junction -Path $skillsLink -Target $skillsTarget | Out-Null
    } else {
        $isReparsePoint = ($existingLink.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0

        if ($isReparsePoint) {
            $currentTarget = $existingLink.Target
            if ($currentTarget -is [System.Array]) {
                $currentTarget = $currentTarget[0]
            }

            if ($currentTarget -isnot [string] -or ($currentTarget -ine $skillsTarget)) {
                Remove-Item -LiteralPath $skillsLink -Force
                New-Item -ItemType Junction -Path $skillsLink -Target $skillsTarget | Out-Null
            }
        } else {
            Write-Warning "Not updating $skillsLink because it exists and is not a link"
        }
    }
}

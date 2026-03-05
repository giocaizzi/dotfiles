$ErrorActionPreference = "Stop"

$skillsUrl = "git@github.com:giocaizzi/skills.git"
$skillsDir = Join-Path $HOME "skills"

$ralphUrl = "git@github.com:giocaizzi/ralph-copilot.git"
$ralphDir = Join-Path $HOME "ralph-copilot"

$copilotAgentsUrl = "git@github.com:giocaizzi/copilot-agents.git"
$copilotAgentsDir = Join-Path $HOME "copilot-agents"

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

Sync-Repo -RepoUrl $skillsUrl -RepoDir $skillsDir
Sync-Repo -RepoUrl $ralphUrl -RepoDir $ralphDir
Sync-Repo -RepoUrl $copilotAgentsUrl -RepoDir $copilotAgentsDir

$claudeDir = Join-Path $HOME ".claude"
$skillsLink = Join-Path $claudeDir "skills"
$skillsTarget = Join-Path $skillsDir "skills"

if (-not (Test-Path $claudeDir)) {
    New-Item -ItemType Directory -Path $claudeDir | Out-Null
}

if ((Test-Path $skillsTarget) -and (-not (Test-Path $skillsLink))) {
    New-Item -ItemType Junction -Path $skillsLink -Target $skillsTarget | Out-Null
}
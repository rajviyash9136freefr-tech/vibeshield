# VibeShield one-command installer for Windows (PowerShell)
# Usage:
#   irm https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.ps1 | iex

param(
    # Kept current by scripts/bump-version.mjs. A stale default here means the
    # PowerShell one-liner installs an old release without saying so.
    [string]$Version = "v3.0.0",
    [string]$InstallDir = "$HOME\.local\bin",
    [switch]$NoSkill = $false,
    [string]$Agent = "auto"
)

$ErrorActionPreference = "Stop"
$Repo = "rajviyash9136freefr-tech/vibeshield"

Write-Host "🛡️  VibeShield Installer for Windows" -ForegroundColor Cyan
Write-Host "Installing VibeShield $Version to $InstallDir..." -ForegroundColor DarkGray

# Detect Architecture
$Arch = "x86_64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $Arch = "aarch64"
}

$ZipName = "vibeshield-windows-$Arch.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$ZipName"
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

try {
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $ZipPath = Join-Path $TempDir $ZipName
    Write-Host "Downloading $DownloadUrl..." -ForegroundColor DarkGray
    
    $Downloaded = $false
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing
        $Downloaded = $true
    } catch {
        Write-Host "Release binary not found at $DownloadUrl. Checking npm or Go fallback..." -ForegroundColor Yellow
    }

    if ($Downloaded -and (Test-Path $ZipPath)) {
        Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force
        $BinSource = Join-Path $TempDir "vibeshield.exe"
        if (Test-Path $BinSource) {
            Copy-Item -Path $BinSource -Destination (Join-Path $InstallDir "vibeshield.exe") -Force
            Write-Host "✓ Installed vibeshield.exe to $InstallDir" -ForegroundColor Green
        }
    } else {
        # Fallback to npm global or go install. Both are pinned to $Version:
        # an unpinned fallback resolves to whatever is published at run time,
        # which is the window VS-DEP-004 exists to flag.
        $NpmVersion = $Version.TrimStart('v')
        if (Get-Command npm -ErrorAction SilentlyContinue) {
            Write-Host "Installing via npm..." -ForegroundColor Cyan
            npm install -g "vibeshield@$NpmVersion"
        } elseif (Get-Command go -ErrorAction SilentlyContinue) {
            Write-Host "Building from source via Go..." -ForegroundColor Cyan
            go install "github.com/$Repo/scanner/cmd/vibeshield@$Version"
        } else {
            Write-Error "Could not download binary and neither npm nor go were found."
        }
    }

    # Ensure PATH contains InstallDir
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "✓ Added $InstallDir to User PATH" -ForegroundColor Green
    }

    # Coding Agent Skill Setup
    if (-not $NoSkill) {
        Write-Host "Configuring AI coding agent security skills..." -ForegroundColor DarkGray
        
        # Claude Code
        $ClaudeDir = Join-Path $HOME ".claude"
        if (Test-Path $ClaudeDir) {
            $SkillDir = Join-Path $ClaudeDir "skills\vibeshield"
            New-Item -ItemType Directory -Path $SkillDir -Force | Out-Null
            Write-Host "✓ Added VibeShield skill to Claude Code" -ForegroundColor Green
        }

        # Cursor (.cursor)
        $CursorDir = Join-Path $HOME ".cursor"
        if (Test-Path $CursorDir) {
            $CursorSkill = Join-Path $CursorDir "skills\vibeshield"
            New-Item -ItemType Directory -Path $CursorSkill -Force | Out-Null
            Write-Host "✓ Added VibeShield skill to Cursor" -ForegroundColor Green
        }
    }

    Write-Host ""
    Write-Host "🎉 VibeShield is ready! Try running:" -ForegroundColor Green
    Write-Host "   vibeshield scan ." -ForegroundColor White
    Write-Host "   vibeshield --help" -ForegroundColor DarkGray

} finally {
    if (Test-Path $TempDir) {
        Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

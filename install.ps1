# install.ps1 — PowerShell installer for ccc (Copilot Config CLI)
# Usage:
#   .\install.ps1
#   .\install.ps1 -Version v1.2.3
#   irm https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.ps1 | iex
#   $env:CCC_VERSION = "v1.2.3"; irm https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.ps1 | iex
#   $env:INSTALL_DIR = "C:\MyTools"; irm https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.ps1 | iex

param(
    [string]$Version = ""
)

$ErrorActionPreference = 'Stop'

# ── Constants ────────────────────────────────────────────────────────────────

$Repo = "jsburckhardt/co-config"
$ProjectName = "co-config"
$BinaryName = "ccc.exe"
$GitHubApi = "https://api.github.com/repos/$Repo/releases"
$GitHubDownload = "https://github.com/$Repo/releases/download"

# ── Helpers ──────────────────────────────────────────────────────────────────

function Write-Info {
    param([string]$Message)
    Write-Host "[info] $Message"
}

function Write-ErrorAndExit {
    param([string]$Message)
    Write-Host "[error] $Message" -ForegroundColor Red
    exit 1
}

# ── Resolve Version ─────────────────────────────────────────────────────────

function Resolve-Version {
    # Prefer the -Version parameter, then $env:CCC_VERSION
    if ($Version -ne "") {
        Write-Info "Using requested version: $Version"
        return $Version
    }

    if ($env:CCC_VERSION -ne $null -and $env:CCC_VERSION -ne "") {
        Write-Info "Using version from CCC_VERSION env var: $env:CCC_VERSION"
        return $env:CCC_VERSION
    }

    Write-Info "Querying GitHub for latest release..."

    $headers = @{ "Accept" = "application/vnd.github.v3+json" }
    if ($env:GITHUB_TOKEN -ne $null -and $env:GITHUB_TOKEN -ne "") {
        $headers["Authorization"] = "token $env:GITHUB_TOKEN"
    }

    try {
        $release = Invoke-RestMethod -Uri "$GitHubApi/latest" -Headers $headers
    }
    catch {
        Write-ErrorAndExit "Failed to query GitHub API for latest release: $_"
    }

    $tagName = $release.tag_name
    if ($tagName -eq $null -or $tagName -eq "") {
        Write-ErrorAndExit "Failed to determine latest version from GitHub API"
    }

    Write-Info "Latest version: $tagName"
    return $tagName
}

# ── Download Release Assets ─────────────────────────────────────────────────

function Get-ReleaseAssets {
    param(
        [string]$ResolvedVersion,
        [string]$TempDir
    )

    $archiveName = "${ProjectName}_windows_amd64.zip"
    $checksumsName = "checksums.txt"
    $downloadBase = "$GitHubDownload/$ResolvedVersion"

    $archivePath = Join-Path $TempDir $archiveName
    $checksumsPath = Join-Path $TempDir $checksumsName

    Write-Info "Downloading $archiveName..."
    try {
        Invoke-WebRequest -Uri "$downloadBase/$archiveName" -OutFile $archivePath -UseBasicParsing
    }
    catch {
        Write-ErrorAndExit "Failed to download ${archiveName}: $_"
    }

    Write-Info "Downloading $checksumsName..."
    try {
        Invoke-WebRequest -Uri "$downloadBase/$checksumsName" -OutFile $checksumsPath -UseBasicParsing
    }
    catch {
        Write-ErrorAndExit "Failed to download ${checksumsName}: $_"
    }

    return @{
        ArchiveName   = $archiveName
        ArchivePath   = $archivePath
        ChecksumsPath = $checksumsPath
    }
}

# ── Checksum Verification ───────────────────────────────────────────────────

function Confirm-Checksum {
    param(
        [string]$ArchivePath,
        [string]$ArchiveName,
        [string]$ChecksumsPath
    )

    Write-Info "Verifying SHA256 checksum..."

    $checksumLines = Get-Content -Path $ChecksumsPath
    $expectedHash = $null

    foreach ($line in $checksumLines) {
        # Each line is: {hash}  {filename}
        if ($line -match "^([0-9a-fA-F]{64})\s+(.+)$") {
            $hash = $Matches[1]
            $fileName = $Matches[2]
            if ($fileName -eq $ArchiveName) {
                $expectedHash = $hash
                break
            }
        }
    }

    if ($expectedHash -eq $null) {
        Write-ErrorAndExit "Archive $ArchiveName not found in checksums.txt"
    }

    $actualHashObj = Get-FileHash -Path $ArchivePath -Algorithm SHA256
    $actualHash = $actualHashObj.Hash

    if ($expectedHash.ToLower() -ne $actualHash.ToLower()) {
        Write-ErrorAndExit "Checksum mismatch for ${ArchiveName}! Expected: $expectedHash, Got: $actualHash"
    }

    Write-Info "Checksum verified successfully"
}

# ── Install Binary ──────────────────────────────────────────────────────────

function Install-Binary {
    param(
        [string]$ArchivePath,
        [string]$TempDir,
        [string]$ResolvedVersion
    )

    Write-Info "Extracting $BinaryName..."
    $extractDir = Join-Path $TempDir "extracted"
    Expand-Archive -Path $ArchivePath -DestinationPath $extractDir -Force

    # Determine install directory
    if ($env:INSTALL_DIR -ne $null -and $env:INSTALL_DIR -ne "") {
        $targetDir = $env:INSTALL_DIR
    }
    else {
        $targetDir = Join-Path $env:LOCALAPPDATA "Programs\ccc"
    }

    if (-not (Test-Path $targetDir)) {
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }

    $sourceBinary = Join-Path $extractDir $BinaryName
    if (-not (Test-Path $sourceBinary)) {
        Write-ErrorAndExit "Binary $BinaryName not found in archive"
    }

    $destBinary = Join-Path $targetDir $BinaryName
    Write-Info "Installing $BinaryName to $targetDir..."
    Copy-Item -Path $sourceBinary -Destination $destBinary -Force

    Write-Info "Successfully installed $BinaryName $ResolvedVersion to $destBinary"

    return $targetDir
}

# ── PATH Management ─────────────────────────────────────────────────────────

function Add-ToUserPath {
    param(
        [string]$Directory
    )

    $currentUserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentUserPath -eq $null) {
        $currentUserPath = ""
    }

    # Check if directory is already in user PATH
    $pathEntries = $currentUserPath.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries)
    $alreadyPresent = $false
    foreach ($entry in $pathEntries) {
        if ($entry.TrimEnd('\') -eq $Directory.TrimEnd('\')) {
            $alreadyPresent = $true
            break
        }
    }

    if (-not $alreadyPresent) {
        Write-Info "Adding $Directory to user PATH..."
        if ($currentUserPath -ne "" -and -not $currentUserPath.EndsWith(";")) {
            $newUserPath = "$currentUserPath;$Directory"
        }
        else {
            $newUserPath = "$currentUserPath$Directory"
        }
        [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
    }
    else {
        Write-Info "$Directory is already in user PATH"
    }

    # Update current session PATH
    $sessionPath = $env:Path
    $sessionEntries = $sessionPath.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries)
    $inSession = $false
    foreach ($entry in $sessionEntries) {
        if ($entry.TrimEnd('\') -eq $Directory.TrimEnd('\')) {
            $inSession = $true
            break
        }
    }

    if (-not $inSession) {
        $env:Path = "$sessionPath;$Directory"
    }
}

# ── Main ─────────────────────────────────────────────────────────────────────

$tempDir = $null

try {
    $resolvedVersion = Resolve-Version

    # Create temp directory
    $tempBase = [System.IO.Path]::GetTempPath()
    $tempDir = Join-Path $tempBase "co-config-install-$([System.Guid]::NewGuid().ToString('N').Substring(0,8))"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    $assets = Get-ReleaseAssets -ResolvedVersion $resolvedVersion -TempDir $tempDir

    Confirm-Checksum `
        -ArchivePath $assets.ArchivePath `
        -ArchiveName $assets.ArchiveName `
        -ChecksumsPath $assets.ChecksumsPath

    $installDir = Install-Binary `
        -ArchivePath $assets.ArchivePath `
        -TempDir $tempDir `
        -ResolvedVersion $resolvedVersion

    Add-ToUserPath -Directory $installDir
}
finally {
    # Cleanup temp directory
    if ($tempDir -ne $null -and (Test-Path $tempDir)) {
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

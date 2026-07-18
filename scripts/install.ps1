param(
    [string]$Version = $env:QC2_VERSION,
    [string]$Repo = $env:QC2_REPO,
    [string]$InstallDir = $env:QC2_INSTALL_DIR,
    [string[]]$Binaries = @()
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($Repo)) {
    $Repo = "imjlk/qc2"
}
if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = "latest"
}
if ([string]::IsNullOrWhiteSpace($InstallDir)) {
    $InstallDir = Join-Path $env:USERPROFILE "AppData\Local\Programs\qc2\bin"
}
if ($Binaries.Count -eq 0) {
    if ($env:QC2_BINARIES) {
        $Binaries = $env:QC2_BINARIES -split "\s+"
    } else {
        $Binaries = @("qc2")
    }
}

function Resolve-ReleaseTag {
    param([string]$SelectedVersion, [string]$SelectedRepo)

    if ($SelectedVersion -ne "latest") {
        if ($SelectedVersion.StartsWith("v")) {
            return $SelectedVersion
        }
        return "v$SelectedVersion"
    }

    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$SelectedRepo/releases/latest"
    return $release.tag_name
}

function Resolve-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { throw "unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
    }
}

function Assert-ArchiveChecksum {
    param(
        [string]$ArchivePath,
        [string]$AssetName,
        [string]$ChecksumsPath
    )

    $escapedAssetName = [Regex]::Escape($AssetName)
    $pattern = "^[a-fA-F0-9]{64}\s+$escapedAssetName$"
    $match = Select-String -Path $ChecksumsPath -Pattern $pattern | Select-Object -First 1
    if (-not $match) {
        throw "checksum not found for $AssetName"
    }

    $expected = ($match.Line -split "\s+")[0]
    $actual = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash
    if (-not $actual.Equals($expected, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "checksum mismatch for $AssetName"
    }
}

function Install-Binary {
    param(
        [string]$Name,
        [string]$Tag,
        [string]$SelectedRepo,
        [string]$Arch,
        [string]$Destination,
        [string]$TempDir,
        [string]$ChecksumsPath
    )

    $versionValue = $Tag.TrimStart("v")
    $assetName = "{0}_{1}_windows_{2}.zip" -f $Name, $versionValue, $Arch
    $url = "https://github.com/$SelectedRepo/releases/download/$Tag/$assetName"
    $archivePath = Join-Path $TempDir $assetName

    Write-Host "installing $Name from $url"
    Invoke-WebRequest -Uri $url -OutFile $archivePath
    Assert-ArchiveChecksum -ArchivePath $archivePath -AssetName $assetName -ChecksumsPath $ChecksumsPath
    Expand-Archive -Path $archivePath -DestinationPath $TempDir -Force

    $source = Join-Path $TempDir ("{0}_{1}_windows_{2}\{0}.exe" -f $Name, $versionValue, $Arch)
    Copy-Item -Path $source -Destination (Join-Path $Destination ("{0}.exe" -f $Name)) -Force
}

$tag = Resolve-ReleaseTag -SelectedVersion $Version -SelectedRepo $Repo
if ([string]::IsNullOrWhiteSpace($tag)) {
    throw "could not resolve a release tag from GitHub"
}

$arch = Resolve-Arch
$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("qc2-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
New-Item -ItemType Directory -Path $tempDir | Out-Null

try {
    $checksumsPath = Join-Path $tempDir "SHA256SUMS"
    $checksumsUrl = "https://github.com/$Repo/releases/download/$tag/SHA256SUMS"
    Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath

    foreach ($binary in $Binaries) {
        if (-not [string]::IsNullOrWhiteSpace($binary)) {
            Install-Binary -Name $binary -Tag $tag -SelectedRepo $Repo -Arch $arch -Destination $InstallDir -TempDir $tempDir -ChecksumsPath $checksumsPath
        }
    }
}
finally {
    Remove-Item -Path $tempDir -Recurse -Force
}

Write-Host "installed to $InstallDir"
if (-not (($env:Path -split ";") -contains $InstallDir)) {
    Write-Host "add $InstallDir to PATH to run the installed commands"
}

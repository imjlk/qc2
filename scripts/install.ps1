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
        return $SelectedVersion
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

function Install-Binary {
    param(
        [string]$Name,
        [string]$Tag,
        [string]$SelectedRepo,
        [string]$Arch,
        [string]$Destination
    )

    $versionValue = $Tag.TrimStart("v")
    $assetName = "{0}_{1}_windows_{2}.zip" -f $Name, $versionValue, $Arch
    $url = "https://github.com/$SelectedRepo/releases/download/$Tag/$assetName"
    $tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("qc2-" + [System.Guid]::NewGuid().ToString("N"))

    New-Item -ItemType Directory -Path $tempDir | Out-Null
    try {
        $archivePath = Join-Path $tempDir $assetName
        Write-Host "installing $Name from $url"
        Invoke-WebRequest -Uri $url -OutFile $archivePath
        Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force

        $source = Join-Path $tempDir ("{0}_{1}_windows_{2}\{0}.exe" -f $Name, $versionValue, $Arch)
        Copy-Item -Path $source -Destination (Join-Path $Destination ("{0}.exe" -f $Name)) -Force
    }
    finally {
        Remove-Item -Path $tempDir -Recurse -Force
    }
}

$tag = Resolve-ReleaseTag -SelectedVersion $Version -SelectedRepo $Repo
if ([string]::IsNullOrWhiteSpace($tag)) {
    throw "could not resolve a release tag from GitHub"
}

$arch = Resolve-Arch
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

foreach ($binary in $Binaries) {
    if (-not [string]::IsNullOrWhiteSpace($binary)) {
        Install-Binary -Name $binary -Tag $tag -SelectedRepo $Repo -Arch $arch -Destination $InstallDir
    }
}

Write-Host "installed to $InstallDir"


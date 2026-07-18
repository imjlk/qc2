param(
    [Parameter(Mandatory = $true)]
    [string]$FixtureBinary
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Assert-Equal {
    param(
        [string]$Actual,
        [string]$Expected,
        [string]$Message
    )

    if ($Actual -ne $Expected) {
        throw "${Message}: got '$Actual', want '$Expected'"
    }
}

function Assert-ThrowsLike {
    param(
        [scriptblock]$Action,
        [string]$Pattern,
        [string]$Message
    )

    try {
        & $Action
    }
    catch {
        if ($_.Exception.Message -notlike $Pattern) {
            throw "${Message}: unexpected error '$($_.Exception.Message)'"
        }
        return
    }

    throw "${Message}: expected an error"
}

$installerPath = Join-Path $PSScriptRoot "install.ps1"
$tokens = $null
$parseErrors = $null
$installerAst = [System.Management.Automation.Language.Parser]::ParseFile(
    $installerPath,
    [ref]$tokens,
    [ref]$parseErrors
)

if ($parseErrors.Count -gt 0) {
    $messages = ($parseErrors | ForEach-Object { $_.Message }) -join "; "
    throw "installer syntax errors: $messages"
}

$functionDefinitions = $installerAst.FindAll({
    param($node)
    $node -is [System.Management.Automation.Language.FunctionDefinitionAst]
}, $true)

foreach ($definition in $functionDefinitions) {
    . ([scriptblock]::Create($definition.Extent.Text))
}

Assert-Equal (Resolve-ReleaseTag "0.1.0" "unused") "v0.1.0" "version normalization failed"
Assert-Equal (Resolve-ReleaseTag "v0.1.0" "unused") "v0.1.0" "version preservation failed"

$originalArchitecture = $env:PROCESSOR_ARCHITECTURE
try {
    $env:PROCESSOR_ARCHITECTURE = "AMD64"
    Assert-Equal (Resolve-Arch) "amd64" "AMD64 detection failed"
    $env:PROCESSOR_ARCHITECTURE = "ARM64"
    Assert-Equal (Resolve-Arch) "arm64" "ARM64 detection failed"
}
finally {
    $env:PROCESSOR_ARCHITECTURE = $originalArchitecture
}

$testRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("qc2-installer-test-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $testRoot | Out-Null

try {
    $version = "0.0.0"
    $archiveBase = "qc2_${version}_windows_amd64"
    $assetName = "${archiveBase}.zip"
    $payloadParent = Join-Path $testRoot "payload"
    $payloadDir = Join-Path $payloadParent $archiveBase
    $fixtureArchive = Join-Path $testRoot $assetName
    $checksumsPath = Join-Path $testRoot "SHA256SUMS"

    New-Item -ItemType Directory -Path $payloadDir -Force | Out-Null
    Copy-Item -LiteralPath $FixtureBinary -Destination (Join-Path $payloadDir "qc2.exe")
    Compress-Archive -Path $payloadDir -DestinationPath $fixtureArchive

    $archiveHash = (Get-FileHash -Path $fixtureArchive -Algorithm SHA256).Hash.ToLowerInvariant()
    [System.IO.File]::WriteAllText($checksumsPath, "${archiveHash}  ${assetName}`n")
    Assert-ArchiveChecksum $fixtureArchive $assetName $checksumsPath

    $tamperedArchive = Join-Path $testRoot "tampered.zip"
    Copy-Item -LiteralPath $fixtureArchive -Destination $tamperedArchive
    [System.IO.File]::AppendAllText($tamperedArchive, "tampered")
    Assert-ThrowsLike {
        Assert-ArchiveChecksum $tamperedArchive $assetName $checksumsPath
    } "checksum mismatch*" "tampered archive was accepted"

    $script:FixtureArchive = $fixtureArchive
    function Invoke-WebRequest {
        param(
            [string]$Uri,
            [string]$OutFile,
            [switch]$UseBasicParsing
        )

        Copy-Item -LiteralPath $script:FixtureArchive -Destination $OutFile -Force
    }

    $downloadDir = Join-Path $testRoot "downloads"
    $extractDir = Join-Path $testRoot "extracted"
    $installDir = Join-Path $testRoot "installed"
    New-Item -ItemType Directory -Path $downloadDir, $extractDir, $installDir -Force | Out-Null

    Install-Binary "qc2" "v${version}" "unused" "amd64" $installDir $downloadDir $extractDir $checksumsPath

    $installedBinary = Join-Path $installDir "qc2.exe"
    if (-not (Test-Path -LiteralPath $installedBinary -PathType Leaf)) {
        throw "installer did not create qc2.exe"
    }

    $fixtureHash = (Get-FileHash -Path $FixtureBinary -Algorithm SHA256).Hash
    $installedHash = (Get-FileHash -Path $installedBinary -Algorithm SHA256).Hash
    Assert-Equal $installedHash $fixtureHash "installed binary differs from fixture"

    $versionOutput = & $installedBinary version
    if ($LASTEXITCODE -ne 0 -or $versionOutput -notlike "qc2 *") {
        throw "installed qc2.exe did not run successfully"
    }
}
finally {
    Remove-Item -LiteralPath $testRoot -Recurse -Force
}

Write-Host "PowerShell installer tests passed"

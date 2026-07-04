[CmdletBinding()]
param(
    [ValidateSet('report', 'enforce')]
    [string]$Mode = 'report',

    [string]$OutputDirectory
)

$ErrorActionPreference = 'Stop'

if (-not $OutputDirectory) {
    $OutputDirectory = Join-Path $PSScriptRoot '..\.artifacts\security'
}

$outputPath = [System.IO.Path]::GetFullPath($OutputDirectory)
$sarifPath = Join-Path $outputPath 'gosec.sarif'
$repositoryRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
New-Item -ItemType Directory -Force -Path $outputPath | Out-Null
if (Test-Path -LiteralPath $sarifPath) {
    Remove-Item -LiteralPath $sarifPath -Force
}

function Convert-ToRepositoryRelativeSarifPaths {
    param(
        [Parameter(Mandatory)]
        [string]$Path,

        [Parameter(Mandatory)]
        [string]$Root
    )

    $rootPath = [System.IO.Path]::GetFullPath($Root).TrimEnd(
        [System.IO.Path]::DirectorySeparatorChar,
        [System.IO.Path]::AltDirectorySeparatorChar
    )
    $rootPrefix = $rootPath + [System.IO.Path]::DirectorySeparatorChar
    $sarif = Get-Content -LiteralPath $Path -Raw | ConvertFrom-Json

    foreach ($run in @($sarif.runs)) {
        foreach ($result in @($run.results)) {
            foreach ($location in @($result.locations)) {
                $artifact = $location.physicalLocation.artifactLocation
                if (-not $artifact -or [string]::IsNullOrWhiteSpace($artifact.uri)) {
                    continue
                }
                if (-not [System.IO.Path]::IsPathRooted($artifact.uri)) {
                    continue
                }

                $absolutePath = [System.IO.Path]::GetFullPath($artifact.uri)
                if (-not $absolutePath.StartsWith(
                    $rootPrefix,
                    [System.StringComparison]::OrdinalIgnoreCase
                )) {
                    throw "SARIF location is outside the repository: $absolutePath"
                }
                $artifact.uri = $absolutePath.Substring($rootPrefix.Length).Replace('\', '/')
            }
        }
    }

    $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    $json = $sarif | ConvertTo-Json -Depth 100
    [System.IO.File]::WriteAllText($Path, $json, $utf8NoBom)
}

foreach ($tool in @('gosec', 'govulncheck')) {
    if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) {
        throw "$tool is required. Install the pinned version documented in README.md."
    }
}

$packages = @(
    go list -f '{{.Dir}}' ./cmd/goverter ./internal/...
)
if ($LASTEXITCODE -ne 0 -or $packages.Count -eq 0) {
    throw 'Could not enumerate Goverter production packages.'
}

$gosecArguments = @(
    '-exclude-generated',
    '-track-suppressions',
    '-fmt=sarif',
    "-out=$sarifPath"
)
if ($Mode -eq 'report') {
    $gosecArguments += '-no-fail'
}
else {
    $gosecArguments += @('-severity=medium', '-confidence=medium')
}
$gosecArguments += $packages

Write-Host "Running gosec in $Mode mode..."
& gosec @gosecArguments
$gosecExitCode = $LASTEXITCODE
if (Test-Path -LiteralPath $sarifPath -PathType Leaf) {
    Convert-ToRepositoryRelativeSarifPaths -Path $sarifPath -Root $repositoryRoot
}

Write-Host 'Running govulncheck...'
& govulncheck ./...
$govulncheckExitCode = $LASTEXITCODE

if (-not (Test-Path -LiteralPath $sarifPath -PathType Leaf)) {
    throw "gosec did not create the SARIF report: $sarifPath"
}

Write-Host "Security report written to $sarifPath"
if ($gosecExitCode -ne 0 -or $govulncheckExitCode -ne 0) {
    exit 1
}

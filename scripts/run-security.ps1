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
New-Item -ItemType Directory -Force -Path $outputPath | Out-Null
if (Test-Path -LiteralPath $sarifPath) {
    Remove-Item -LiteralPath $sarifPath -Force
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

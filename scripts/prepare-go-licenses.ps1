[CmdletBinding()]
param(
    [string]$OutputDirectory
)

$ErrorActionPreference = 'Stop'

if (-not $OutputDirectory) {
    $OutputDirectory = Join-Path $PSScriptRoot '..\dist'
}

$outputPath = [System.IO.Path]::GetFullPath($OutputDirectory)
$licenseRoot = Join-Path $outputPath 'third_party\go'
New-Item -ItemType Directory -Force -Path $licenseRoot | Out-Null

$licenses = @(
    @{
        Module = 'github.com/spf13/cobra@v1.10.2'
        Source = 'LICENSE.txt'
        Name = 'cobra'
    },
    @{
        Module = 'github.com/spf13/pflag@v1.0.9'
        Source = 'LICENSE'
        Name = 'pflag'
    },
    @{
        Module = 'github.com/inconshreveable/mousetrap@v1.1.0'
        Source = 'LICENSE'
        Name = 'mousetrap'
    }
)

foreach ($license in $licenses) {
    $moduleInfo = go mod download -json $license.Module | ConvertFrom-Json
    if ($LASTEXITCODE -ne 0 -or -not $moduleInfo.Dir) {
        throw "Could not locate Go module $($license.Module)."
    }

    $source = Join-Path $moduleInfo.Dir $license.Source
    if (-not (Test-Path -LiteralPath $source -PathType Leaf)) {
        throw "License file not found for $($license.Module): $source"
    }

    $destination = Join-Path $licenseRoot $license.Name
    New-Item -ItemType Directory -Force -Path $destination | Out-Null
    Copy-Item -LiteralPath $source -Destination (Join-Path $destination 'LICENSE') -Force
}

Write-Host "Go dependency licenses staged in $licenseRoot"

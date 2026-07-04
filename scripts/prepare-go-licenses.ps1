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

$moduleLines = @(
    go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}}|{{.Version}}|{{.Dir}}{{end}}{{end}}' `
        ./cmd/goverter |
        Where-Object { $_ } |
        Sort-Object -Unique
)
if ($LASTEXITCODE -ne 0) {
    throw 'Could not enumerate Go dependencies used by Goverter.'
}

$manifest = @(
    'Go modules included in Goverter',
    'Generated from the dependencies compiled into ./cmd/goverter.',
    ''
)

foreach ($line in $moduleLines) {
    $parts = $line -split '\|', 3
    if ($parts.Count -ne 3) {
        throw "Unexpected Go module information: $line"
    }

    $modulePath, $version, $moduleDirectory = $parts
    $licenseFiles = @(
        Get-ChildItem -LiteralPath $moduleDirectory -File |
            Where-Object { $_.Name -match '^(LICENSE|COPYING|NOTICE)(\..*)?$' } |
            Sort-Object Name
    )
    if ($licenseFiles.Count -eq 0) {
        throw "No license or notice file found for $modulePath@$version."
    }

    $directoryName = (($modulePath + '@' + $version) -replace '[^A-Za-z0-9._@-]', '_')
    $destination = Join-Path $licenseRoot $directoryName
    New-Item -ItemType Directory -Force -Path $destination | Out-Null
    foreach ($licenseFile in $licenseFiles) {
        Copy-Item -LiteralPath $licenseFile.FullName -Destination $destination -Force
    }
    $manifest += "$modulePath $version -> $directoryName"
}

$manifest | Set-Content -LiteralPath (Join-Path $licenseRoot 'THIRD_PARTY_MODULES.txt') `
    -Encoding UTF8

Write-Host "Go dependency licenses staged in $licenseRoot"

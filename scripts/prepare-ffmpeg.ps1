[CmdletBinding()]
param(
    [string]$LockFile,
    [string]$OutputDirectory
)

$ErrorActionPreference = 'Stop'

if (-not $LockFile) {
    $LockFile = Join-Path $PSScriptRoot '..\tools.lock.json'
}
if (-not $OutputDirectory) {
    $OutputDirectory = Join-Path $PSScriptRoot '..\dist'
}

$lockPath = (Resolve-Path -LiteralPath $LockFile).Path
$lock = Get-Content -LiteralPath $lockPath -Raw | ConvertFrom-Json
$package = $lock.ffmpeg

$outputPath = [System.IO.Path]::GetFullPath($OutputDirectory)
$toolsPath = Join-Path $outputPath 'tools'
$noticePath = Join-Path $outputPath 'third_party\ffmpeg'
New-Item -ItemType Directory -Force -Path $toolsPath, $noticePath | Out-Null

$tempRoot = [System.IO.Path]::GetFullPath([System.IO.Path]::GetTempPath())
$workPath = Join-Path $tempRoot ("goverter-ffmpeg-" + [guid]::NewGuid().ToString('N'))
$archivePath = Join-Path $workPath 'ffmpeg.zip'
$expandedPath = Join-Path $workPath 'expanded'

try {
    New-Item -ItemType Directory -Force -Path $workPath | Out-Null
    Write-Host "Downloading FFmpeg $($package.version)..."
    $curl = Get-Command 'curl.exe' -ErrorAction SilentlyContinue
    if ($curl) {
        & $curl.Source --fail --location --retry 3 --output $archivePath $package.url
        if ($LASTEXITCODE -ne 0) {
            throw "curl failed with exit code $LASTEXITCODE."
        }
    }
    else {
        Invoke-WebRequest -UseBasicParsing -Uri $package.url -OutFile $archivePath
    }

    $actualHash = (Get-FileHash -LiteralPath $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actualHash -ne $package.sha256.ToLowerInvariant()) {
        throw "FFmpeg checksum mismatch. Expected $($package.sha256), got $actualHash."
    }

    Expand-Archive -LiteralPath $archivePath -DestinationPath $expandedPath
    $ffmpeg = Get-ChildItem -LiteralPath $expandedPath -Filter 'ffmpeg.exe' -File -Recurse |
        Select-Object -First 1
    $ffprobe = Get-ChildItem -LiteralPath $expandedPath -Filter 'ffprobe.exe' -File -Recurse |
        Select-Object -First 1
    $license = Get-ChildItem -LiteralPath $expandedPath -File -Recurse |
        Where-Object { $_.Name -in @('LICENSE', 'LICENSE.txt', 'COPYING.GPLv3') } |
        Select-Object -First 1
    $readme = Get-ChildItem -LiteralPath $expandedPath -Filter 'README.txt' -File -Recurse |
        Select-Object -First 1

    if (-not $ffmpeg -or -not $ffprobe -or -not $license -or -not $readme) {
        throw 'The FFmpeg archive does not contain the expected tools and notice files.'
    }

    Copy-Item -LiteralPath $ffmpeg.FullName -Destination (Join-Path $toolsPath 'ffmpeg.exe') -Force
    Copy-Item -LiteralPath $ffprobe.FullName -Destination (Join-Path $toolsPath 'ffprobe.exe') -Force
    Copy-Item -LiteralPath $license.FullName -Destination (Join-Path $noticePath 'LICENSE') -Force
    Copy-Item -LiteralPath $readme.FullName -Destination (Join-Path $noticePath 'README.txt') -Force
    Copy-Item -LiteralPath $lockPath -Destination (Join-Path $noticePath 'tools.lock.json') -Force

    Write-Host "FFmpeg staged in $toolsPath"
}
finally {
    $resolvedWorkPath = [System.IO.Path]::GetFullPath($workPath)
    if ($resolvedWorkPath.StartsWith($tempRoot, [System.StringComparison]::OrdinalIgnoreCase) -and
        (Test-Path -LiteralPath $resolvedWorkPath)) {
        Remove-Item -LiteralPath $resolvedWorkPath -Recurse -Force
    }
}

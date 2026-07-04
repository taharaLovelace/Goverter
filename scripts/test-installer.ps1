[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$InstallerPath
)

$ErrorActionPreference = 'Stop'

$installer = (Resolve-Path -LiteralPath $InstallerPath).Path
$installPath = Join-Path ([System.IO.Path]::GetTempPath()) (
    'goverter-installer-test-' + [guid]::NewGuid().ToString('N')
)
$originalUserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$installed = $false

function Test-PathEntry {
    param([string]$PathValue, [string]$Expected)

    $expectedPath = $Expected.TrimEnd('\')
    foreach ($entry in ($PathValue -split ';')) {
        if ($entry.Trim().TrimEnd('\').Equals(
            $expectedPath,
            [System.StringComparison]::OrdinalIgnoreCase
        )) {
            return $true
        }
    }
    return $false
}

try {
    $arguments = @(
        '/VERYSILENT',
        '/SUPPRESSMSGBOXES',
        '/NORESTART',
        '/TASKS="addtopath"',
        ('/DIR="' + $installPath + '"')
    )
    $process = Start-Process -FilePath $installer -ArgumentList $arguments `
        -WindowStyle Hidden -Wait -PassThru
    if ($process.ExitCode -ne 0) {
        throw "Installer exited with code $($process.ExitCode)."
    }
    $installed = $true

    $goverter = Join-Path $installPath 'goverter.exe'
    foreach ($file in @(
        $goverter,
        (Join-Path $installPath 'tools\ffmpeg.exe'),
        (Join-Path $installPath 'tools\ffprobe.exe'),
        (Join-Path $installPath 'licenses\ffmpeg\LICENSE'),
        (Join-Path $installPath 'licenses\go\THIRD_PARTY_MODULES.txt'),
        (Join-Path $installPath 'licenses\go\github.com_pdfcpu_pdfcpu@v0.13.0\LICENSE.txt'),
        (Join-Path $installPath 'licenses\go\github.com_spf13_cobra@v1.10.2\LICENSE.txt'),
        (Join-Path $installPath 'licenses\go\golang.org_x_sys@v0.46.0\LICENSE')
    )) {
        if (-not (Test-Path -LiteralPath $file -PathType Leaf)) {
            throw "Installer did not create expected file: $file"
        }
    }

    & $goverter formats --json | ConvertFrom-Json | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Installed Goverter exited with code $LASTEXITCODE."
    }

    $installedUserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (-not (Test-PathEntry -PathValue $installedUserPath -Expected $installPath)) {
        throw 'Installer did not add Goverter to the current user PATH.'
    }

    $uninstaller = Join-Path $installPath 'unins000.exe'
    $process = Start-Process -FilePath $uninstaller `
        -ArgumentList @('/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART') `
        -WindowStyle Hidden -Wait -PassThru
    if ($process.ExitCode -ne 0) {
        throw "Uninstaller exited with code $($process.ExitCode)."
    }
    $installed = $false

    $uninstalledUserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (Test-PathEntry -PathValue $uninstalledUserPath -Expected $installPath) {
        throw 'Uninstaller left Goverter in the current user PATH.'
    }
    if (Test-Path -LiteralPath $installPath) {
        throw "Uninstaller left the installation directory behind: $installPath"
    }

    Write-Host 'Installer round-trip test passed.'
}
finally {
    if ($installed) {
        $uninstaller = Join-Path $installPath 'unins000.exe'
        if (Test-Path -LiteralPath $uninstaller -PathType Leaf) {
            Start-Process -FilePath $uninstaller `
                -ArgumentList @('/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART') `
                -WindowStyle Hidden -Wait | Out-Null
        }
    }
    [Environment]::SetEnvironmentVariable('Path', $originalUserPath, 'User')
}

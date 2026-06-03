param(
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"

$OutputDir = Join-Path -Path $PSScriptRoot -ChildPath $OutputDir

function Copy-RuntimeFiles {
    Write-Host ">>> Copy runtime files to $OutputDir ..." -ForegroundColor Cyan
    Copy-Item -LiteralPath (Join-Path -Path $PSScriptRoot -ChildPath "templates") -Destination $OutputDir -Recurse
    Copy-Item -LiteralPath (Join-Path -Path $PSScriptRoot -ChildPath "static") -Destination $OutputDir -Recurse
    Copy-Item -LiteralPath (Join-Path -Path $PSScriptRoot -ChildPath "config.ini") -Destination $OutputDir
    Copy-Item -LiteralPath (Join-Path -Path $PSScriptRoot -ChildPath "to_dist.md") -Destination (Join-Path -Path $OutputDir -ChildPath "README.md")
}

function Build-Target {
    param(
        [string]$Goos,
        [string]$Goarch,
        [string]$TargetDir,
        [string]$BinaryName
    )

    Write-Host ">>> Build $Goos/$Goarch binary ..." -ForegroundColor Cyan
    $env:CGO_ENABLED = "0"
    $env:GOOS = $Goos
    $env:GOARCH = $Goarch
    go build -o (Join-Path -Path $TargetDir -ChildPath $BinaryName) -ldflags="-s -w -X main.version=$(Get-Date -Format 'yy.MMdd.HHmm')"
}

if (Test-Path -LiteralPath $OutputDir) {
    Remove-Item -LiteralPath $OutputDir -Recurse -Force
}
New-Item -ItemType Directory -Path $OutputDir | Out-Null

$prevCgo = $env:CGO_ENABLED
$prevGoos = $env:GOOS
$prevGoarch = $env:GOARCH

try {
    Build-Target -Goos "linux" -Goarch "amd64" -TargetDir $OutputDir -BinaryName "cms"
    Build-Target -Goos "windows" -Goarch "amd64" -TargetDir $OutputDir -BinaryName "cms.exe"
    Copy-RuntimeFiles
}
finally {
    $env:CGO_ENABLED = $prevCgo
    $env:GOOS = $prevGoos
    $env:GOARCH = $prevGoarch
}

Write-Host "<<< Build complete: $OutputDir" -ForegroundColor Green
Write-Host "    Linux binary:   cms" -ForegroundColor Green
Write-Host "    Windows binary: cms.exe" -ForegroundColor Green

param(
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"

$OutputDir = Join-Path -Path $PSScriptRoot -ChildPath $OutputDir

if (Test-Path -LiteralPath $OutputDir) {
    Remove-Item -LiteralPath $OutputDir -Recurse -Force
}
New-Item -ItemType Directory -Path $OutputDir | Out-Null

Write-Host ">>> 编译 linux/amd64 二进制文件 ..." -ForegroundColor Cyan
$prevCgo = $env:CGO_ENABLED
$prevGoos = $env:GOOS
$prevGoarch = $env:GOARCH
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o (Join-Path -Path $OutputDir -ChildPath "cms") -ldflags="-s -w"
$env:CGO_ENABLED = $prevCgo
$env:GOOS = $prevGoos
$env:GOARCH = $prevGoarch

Write-Host ">>> 复制 templates ..." -ForegroundColor Cyan
Copy-Item -LiteralPath (Join-Path -Path $PSScriptRoot -ChildPath "templates") -Destination $OutputDir -Recurse

Write-Host ">>> 复制 static ..." -ForegroundColor Cyan
Copy-Item -LiteralPath (Join-Path -Path $PSScriptRoot -ChildPath "static") -Destination $OutputDir -Recurse

Write-Host ">>> 复制 config.ini ..." -ForegroundColor Cyan
Copy-Item -LiteralPath (Join-Path -Path $PSScriptRoot -ChildPath "config.ini") -Destination $OutputDir

Write-Host "<<< 打包完成: $OutputDir" -ForegroundColor Green

param([string]$Version = "0.1.0")
$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$out = Join-Path $root "dist\supercode-$Version-windows-amd64"
$zip = "$out.zip"

Remove-Item -Recurse -Force $out, $zip -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $out | Out-Null

Push-Location $root
try {
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -trimpath -ldflags "-s -w -X supercode/internal/cli.Version=$Version" -o (Join-Path $out "supercode.exe") .
    if ($LASTEXITCODE -ne 0) { throw "go build failed" }
} finally {
    Remove-Item Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
    Pop-Location
}

Copy-Item (Join-Path $PSScriptRoot "install.cmd"), (Join-Path $PSScriptRoot "install.ps1"), (Join-Path $PSScriptRoot "uninstall.cmd"), (Join-Path $PSScriptRoot "README.txt") $out
for ($i = 1; $i -le 10; $i++) {
    try {
        Compress-Archive -Force -Path "$out\*" -DestinationPath $zip -ErrorAction Stop
        break
    } catch {
        if ($i -eq 10) { throw }
        Start-Sleep -Seconds 2
    }
}
Write-Host "Built $zip"

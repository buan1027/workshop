param(
    [switch]$SkipOnlineTools
)

$ErrorActionPreference = "Stop"

function Invoke-Checked {
    param(
        [string]$FilePath,
        [string[]]$Arguments
    )

    & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed with exit code $LASTEXITCODE`: $FilePath $($Arguments -join ' ')"
    }
}

$go = "go"
if (Test-Path "C:\Program Files\Go\bin\go.exe") {
    $go = "C:\Program Files\Go\bin\go.exe"
}

$gofmt = "gofmt"
if (Test-Path "C:\Program Files\Go\bin\gofmt.exe") {
    $gofmt = "C:\Program Files\Go\bin\gofmt.exe"
}

if (-not $env:GOCACHE) {
    $env:GOCACHE = Join-Path (Get-Location) ".gocache"
}
if (-not $env:GOFLAGS) {
    $env:GOFLAGS = "-buildvcs=false"
}

Write-Host "==> gofmt pruefen"
$unformatted = & $gofmt -l .
if ($unformatted) {
    Write-Error "Diese Dateien muessen mit gofmt formatiert werden:`n$($unformatted -join "`n")"
}

Write-Host "==> go vet"
Invoke-Checked $go @("vet", "./...")

Write-Host "==> go test"
Invoke-Checked $go @("test", "./...")

if (-not $SkipOnlineTools) {
    Write-Host "==> staticcheck"
    Invoke-Checked $go @("run", "honnef.co/go/tools/cmd/staticcheck@latest", "./...")

    Write-Host "==> govulncheck"
    Invoke-Checked $go @("run", "golang.org/x/vuln/cmd/govulncheck@latest", "./...")
}

Write-Host "Alle Checks erfolgreich."

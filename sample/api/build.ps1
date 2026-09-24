# =============================================================================
# KIZUNA - Build do exemplo de descritores de API (.api): o mesmo programa em
# BASIC (DIGNAC) e em Pascal (WIRTH80), chamando rotinas da MSXLIB.
# =============================================================================
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir   = Split-Path -Parent (Split-Path -Parent $scriptDir)
$libFile   = Join-Path $rootDir "lib\msxlib.hlib"
$apiDir    = Join-Path $rootDir "lib\api"

if (-not (Test-Path $libFile)) {
    Write-Host "Aviso: msxlib.hlib nao encontrada. Construindo biblioteca..." -ForegroundColor Yellow
    pwsh -File (Join-Path $rootDir "lib\build.ps1")
}

$jobs = @(
    @{ Name = "api_demo";  Src = "api_demo.bas"; Tool = "dignac"  },
    @{ Name = "api_demop"; Src = "api_demo.pas"; Tool = "wirth80" }
)

foreach ($j in $jobs) {
    $src = Join-Path $scriptDir $j.Src
    $mob = Join-Path $scriptDir "$($j.Name).mob"
    $com = Join-Path $scriptDir "$($j.Name).com"
    $map = Join-Path $scriptDir "$($j.Name).map"

    Write-Host "`n[$($j.Name)] Compilando $($j.Src) com $($j.Tool)..." -ForegroundColor Green
    go run "$rootDir\cmd\$($j.Tool)" $src -o $mob -api $apiDir
    if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar $src" }

    Write-Host "[$($j.Name)] Linkando com msxlib.hlib..." -ForegroundColor Green
    go run "$rootDir\cmd\musubi" -o $com -m $map $mob $libFile
    if ($LASTEXITCODE -ne 0) { throw "Falha ao linkar $mob" }

    Write-Host "[$($j.Name)] Executavel: $com" -ForegroundColor Cyan
}
Write-Host "`nOK: api_demo.com (BASIC) e api_demop.com (Pascal)." -ForegroundColor Green

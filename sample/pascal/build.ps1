# =============================================================================
# KIZUNA - Build dos Exemplos Pascal (WIRTH80 + MUSUBI + MSXLIB)
# =============================================================================
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir   = Split-Path -Parent (Split-Path -Parent $scriptDir)
$libFile   = Join-Path $rootDir "lib\msxlib.hlib"

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "    KIZUNA - Compilacao dos Exemplos Pascal (WIRTH80)           " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

# Verifica se a biblioteca padrao existe
if (-not (Test-Path $libFile)) {
    Write-Host "Aviso: msxlib.hlib nao encontrada. Construindo biblioteca..." -ForegroundColor Yellow
    pwsh -File (Join-Path $rootDir "lib\build.ps1")
}

$examples = @("hello", "calc")

foreach ($name in $examples) {
    $pasFile = Join-Path $scriptDir "$name.pas"
    $mobFile = Join-Path $scriptDir "$name.mob"
    $comFile = Join-Path $scriptDir "$name.com"
    $mapFile = Join-Path $scriptDir "$name.map"

    Write-Host "`n[$name] Compilando $name.pas com WIRTH80..." -ForegroundColor Green
    go run "$rootDir\cmd\wirth80" $pasFile -o $mobFile -v
    if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar $pasFile com wirth80" }

    Write-Host "[$name] Linkando $name.mob com msxlib.hlib (Smart-Linking)..." -ForegroundColor Green
    go run "$rootDir\cmd\musubi" -o $comFile -m $mapFile $mobFile $libFile
    if ($LASTEXITCODE -ne 0) { throw "Falha ao linkar $mobFile com musubi" }

    Write-Host "[$name] Executavel gerado com sucesso: $comFile" -ForegroundColor Cyan
}

Write-Host "`n=================================================================" -ForegroundColor Green
Write-Host "       TODOS OS EXEMPLOS PASCAL FORAM CONSTRUIDOS!               " -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Green

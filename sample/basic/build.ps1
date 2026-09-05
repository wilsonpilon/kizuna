# =============================================================================
# KIZUNA - Build dos Exemplos MSX-BASIC Dignified (DIGNAC + MUSUBI + MSXLIB)
# =============================================================================
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir   = Split-Path -Parent (Split-Path -Parent $scriptDir)
$libFile   = Join-Path $rootDir "lib\msxlib.hlib"

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "    KIZUNA - Compilacao dos Exemplos BASIC (DIGNAC)              " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

# Verifica se a biblioteca padrao existe
if (-not (Test-Path $libFile)) {
    Write-Host "Aviso: msxlib.hlib nao encontrada. Construindo biblioteca..." -ForegroundColor Yellow
    pwsh -File (Join-Path $rootDir "lib\build.ps1")
}

# 1. Compila os executáveis standalone: hello, calc e chart
$executables = @("hello", "calc", "chart")

foreach ($name in $executables) {
    $basFile = Join-Path $scriptDir "$name.bas"
    $mobFile = Join-Path $scriptDir "$name.mob"
    $comFile = Join-Path $scriptDir "$name.com"
    $mapFile = Join-Path $scriptDir "$name.map"

    Write-Host "`n[$name] Compilando $name.bas com DIGNAC..." -ForegroundColor Green
    go run "$rootDir\cmd\dignac" $basFile -o $mobFile --log -v
    if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar $basFile com dignac" }

    Write-Host "[$name] Linkando $name.mob com msxlib.hlib (Smart-Linking)..." -ForegroundColor Green
    go run "$rootDir\cmd\musubi" -o $comFile -m $mapFile --log $mobFile $libFile
    if ($LASTEXITCODE -ne 0) { throw "Falha ao linkar $mobFile com musubi" }

    Write-Host "[$name] Executavel gerado com sucesso: $comFile" -ForegroundColor Cyan
}

Write-Host "`n=================================================================" -ForegroundColor Green
Write-Host "       TODOS OS EXEMPLOS BASIC FORAM CONSTRUIDOS!                " -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Green

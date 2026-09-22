# =============================================================================
# KIZUNA - Build do Exemplo OBI (orquestrador declarativo via Obifile)
# =============================================================================
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir   = Split-Path -Parent (Split-Path -Parent $scriptDir)
$libFile   = Join-Path $rootDir "lib\msxlib.hlib"
$obifile   = Join-Path $scriptDir "Obifile"

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "    KIZUNA - Compilacao do Exemplo OBI (Obifile declarativo)     " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

# Verifica se a biblioteca padrao existe
if (-not (Test-Path $libFile)) {
    Write-Host "Aviso: msxlib.hlib nao encontrada. Construindo biblioteca..." -ForegroundColor Yellow
    pwsh -File (Join-Path $rootDir "lib\build.ps1")
}

Write-Host "`n[obi] Construindo $obifile ..." -ForegroundColor Green
go run "$rootDir\cmd\obi" build -v --log $obifile
if ($LASTEXITCODE -ne 0) { throw "Falha ao construir $obifile com obi" }

Write-Host "`n=================================================================" -ForegroundColor Green
Write-Host "       EXEMPLO OBI CONSTRUIDO COM SUCESSO!                       " -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Green

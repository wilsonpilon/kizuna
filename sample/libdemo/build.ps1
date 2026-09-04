# =============================================================================
# KIZUNA - Script de Build do Exemplo MSXLIB com Smart-Linking
# =============================================================================

$ErrorActionPreference = "Stop"

$SampleDir = $PSScriptRoot
if (-not $SampleDir) {
    $SampleDir = (Get-Location).Path
}

$RootDir = (Resolve-Path (Join-Path $SampleDir "..\..")).Path
$LibPath = Join-Path $RootDir "lib\msxlib.hlib"

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "    KIZUNA - Build do Exemplo MSXLIB (Smart-Linking Demo)        " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

# 1. Montar main.asm
$asmFile = Join-Path $SampleDir "main.asm"
$mobFile = Join-Path $SampleDir "main.mob"
Write-Host "[1/2] Montando main.asm -> main.mob..." -ForegroundColor Yellow
& go run "$RootDir/cmd/kaji80" -v -o $mobFile $asmFile
if ($LASTEXITCODE -ne 0) { throw "Falha na montagem de $asmFile" }

# 2. Linkar com a biblioteca msxlib.hlib
$outCom = Join-Path $SampleDir "libdemo.com"
$mapFile = Join-Path $SampleDir "libdemo.map"
Write-Host "[2/2] Linkando main.mob com msxlib.hlib (Smart-Linking)..." -ForegroundColor Yellow
& go run "$RootDir/cmd/musubi" -v -m $mapFile -o $outCom $mobFile $LibPath
if ($LASTEXITCODE -ne 0) { throw "Falha na linkagem" }

Write-Host ""
Write-Host "Build concluido com sucesso!" -ForegroundColor Green
Write-Host "Executavel: $outCom" -ForegroundColor White
Write-Host "Mapa:       $mapFile" -ForegroundColor White

# =============================================================================
# KIZUNA - Build dos Exemplos de recursos ao estilo asMSX do KAJI80
# (avaliador de expressoes, rotulos locais, IF/REPT/MACRO, rotulos
# pre-definidos de BIOS/BDOS, CALLBIOS/CALLDOS, INCBIN)
# =============================================================================

$ErrorActionPreference = "Stop"

$SampleDir = $PSScriptRoot
if (-not $SampleDir) {
    $SampleDir = (Get-Location).Path
}

$RootDir = (Resolve-Path (Join-Path $SampleDir "..\..")).Path
$LibPath = Join-Path $RootDir "lib\msxlib.hlib"

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "    KIZUNA - Build dos Exemplos de recursos ao estilo asMSX      " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

$Programs = @("expr_labels", "predefined", "incbin")

foreach ($name in $Programs) {
    $asmFile = Join-Path $SampleDir "$name.asm"
    $mobFile = Join-Path $SampleDir "$name.mob"
    $comFile = Join-Path $SampleDir "$name.com"
    $mapFile = Join-Path $SampleDir "$name.map"
    $logFile = Join-Path $SampleDir "$name.log"

    Write-Host "`n[$name] Montando $name.asm -> $name.mob..." -ForegroundColor Yellow
    & go run "$RootDir/cmd/kaji80" -v --log-file $logFile -o $mobFile $asmFile
    if ($LASTEXITCODE -ne 0) { throw "Falha na montagem de $asmFile" }

    Write-Host "[$name] Linkando $name.mob com msxlib.hlib (Smart-Linking)..." -ForegroundColor Yellow
    & go run "$RootDir/cmd/musubi" -v --log-file $logFile -m $mapFile -o $comFile $mobFile $LibPath
    if ($LASTEXITCODE -ne 0) { throw "Falha na linkagem de $mobFile" }

    Write-Host "[$name] Executavel gerado: $comFile" -ForegroundColor Green
}

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Green
Write-Host "    TODOS OS EXEMPLOS FORAM CONSTRUIDOS!                         " -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Green
Write-Host ""
Write-Host "expr_labels.com -- avaliador de expressoes, rotulos locais, IF, REPT, MACRO" -ForegroundColor White
Write-Host "predefined.com  -- rotulos pre-definidos de BIOS/BDOS + CALLBIOS/CALLDOS (troca SCREEN 1/0 na tela)" -ForegroundColor White
Write-Host "incbin.com      -- INCBIN com e sem SKIP, a partir de greeting.dat" -ForegroundColor White

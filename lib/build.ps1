# =============================================================================
# KIZUNA - Script de Montagem e Empacotamento da Biblioteca MSXLIB
# =============================================================================

$ErrorActionPreference = "Stop"

$LibDir = $PSScriptRoot
if (-not $LibDir) {
    $LibDir = (Get-Location).Path
}

$RootDir = (Resolve-Path (Join-Path $LibDir "..")).Path
$SrcDir  = Join-Path $LibDir "src"
$OutHlib = Join-Path $LibDir "msxlib.hlib"

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "    KIZUNA (絆) - Compilacao da Biblioteca Padrao MSXLIB         " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

# Lista dos módulos da biblioteca
$Modules = @("bdos", "bios", "vdp", "psg", "string", "math")
$MobFiles = @()

foreach ($m in $Modules) {
    $asmFile = Join-Path $SrcDir "$m.asm"
    $mobFile = Join-Path $SrcDir "$m.mob"
    Write-Host "Montando $m.asm -> $m.mob..." -ForegroundColor Yellow
    & go run "$RootDir/cmd/kaji80" -v -o $mobFile $asmFile
    if ($LASTEXITCODE -ne 0) { throw "Falha na montagem de $asmFile" }
    $MobFiles += $mobFile
}

Write-Host "Empacotando $($MobFiles.Count) modulos em $OutHlib..." -ForegroundColor Green
& go run "$RootDir/cmd/hako" -c $OutHlib @MobFiles
if ($LASTEXITCODE -ne 0) { throw "Falha no empacotamento da biblioteca $OutHlib" }

Write-Host ""
Write-Host "Conteudo e Dicionario da Biblioteca gerada:" -ForegroundColor Cyan
& go run "$RootDir/cmd/hako" -t $OutHlib

Write-Host ""
Write-Host "MSXLIB construida com sucesso -> $OutHlib" -ForegroundColor Green

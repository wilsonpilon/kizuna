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

# Cada .asm sob lib/src (recursivamente) e UM modulo da biblioteca -- o MUSUBI
# so traz para o programa os modulos cujos simbolos ele realmente usa, entao
# quanto menor o modulo, menor o executavel final. Constantes compartilhadas
# ficam em lib/inc/*.inc (INCLUDE).
$ObjDir = Join-Path $LibDir "obj"
if (Test-Path $ObjDir) { Remove-Item -Recurse -Force $ObjDir }
New-Item -ItemType Directory -Force $ObjDir | Out-Null

$AsmFiles = Get-ChildItem -Path $SrcDir -Filter *.asm -Recurse | Sort-Object FullName
$MobFiles = @()

# Compila o kaji80 e o hako uma vez so (go run por modulo custaria minutos)
$BinDir = Join-Path $ObjDir "bin"
New-Item -ItemType Directory -Force $BinDir | Out-Null
& go build -o (Join-Path $BinDir "kaji80.exe") "$RootDir/cmd/kaji80"
if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar kaji80" }
& go build -o (Join-Path $BinDir "hako.exe") "$RootDir/cmd/hako"
if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar hako" }
$Kaji = Join-Path $BinDir "kaji80.exe"
$Hako = Join-Path $BinDir "hako.exe"

foreach ($f in $AsmFiles) {
    $rel = $f.FullName.Substring($SrcDir.Length + 1)
    $mobName = ($rel -replace '[\\/]', '_') -replace '\.asm$', '.mob'
    $mobFile = Join-Path $ObjDir $mobName
    Write-Host "Montando $rel -> $mobName" -ForegroundColor Yellow
    & $Kaji -o $mobFile $f.FullName
    if ($LASTEXITCODE -ne 0) { throw "Falha na montagem de $($f.FullName)" }
    $MobFiles += $mobFile
}

Write-Host "Empacotando $($MobFiles.Count) modulos em $OutHlib..." -ForegroundColor Green
& $Hako -c $OutHlib @MobFiles
if ($LASTEXITCODE -ne 0) { throw "Falha no empacotamento da biblioteca $OutHlib" }

Write-Host ""
Write-Host "Conteudo e Dicionario da Biblioteca gerada:" -ForegroundColor Cyan
& $Hako -t $OutHlib

Write-Host ""
Write-Host "MSXLIB construida com sucesso -> $OutHlib" -ForegroundColor Green

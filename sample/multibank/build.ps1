# ==============================================================================
# KIZUNA (絆) - Script de Build do Exemplo Multi-Banco
# ==============================================================================
# Este script pode ser executado diretamente pelo PowerShell:
#   .\sample\multibank\build.ps1
# ou entrando no diretório:
#   cd sample\multibank; .\build.ps1
# ==============================================================================

$ErrorActionPreference = "Stop"

# Determina o diretório onde o script está localizado
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
# O diretório raiz do KIZUNA é 2 níveis acima (e:\kizuna)
$RootDir = Split-Path -Parent (Split-Path -Parent $ScriptDir)

# Obtém a versão atual do Kizuna
$KizunaVer = & go run "$RootDir/cmd/musubi" --version 2>&1

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "       $KizunaVer" -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "Diretorio do Projeto: $RootDir" -ForegroundColor Gray
Write-Host "Diretorio do Exemplo: $ScriptDir" -ForegroundColor Gray
Write-Host ""

# Arquivos Fonte Assembly
$MainAsm  = Join-Path $ScriptDir "main.asm"
$Bank1Asm = Join-Path $ScriptDir "bank1.asm"
$Bank2Asm = Join-Path $ScriptDir "bank2.asm"

# Arquivos Objeto .MOB intermediários
$MainMob  = Join-Path $ScriptDir "main.mob"
$Bank1Mob = Join-Path $ScriptDir "bank1.mob"
$Bank2Mob = Join-Path $ScriptDir "bank2.mob"

# Saídas finais do Linker
$OutputCom = Join-Path $ScriptDir "multibank.com"
$OutputMap = Join-Path $ScriptDir "multibank.map"

# ------------------------------------------------------------------------------
# ETAPA 1: Montagem dos Módulos Z80 para o formato .MOB via KAJI80
# ------------------------------------------------------------------------------
Write-Host "[1/4] Montando modulo Banco 0 (Area Comum): main.asm -> main.mob..." -ForegroundColor Yellow
go run "$RootDir/cmd/kaji80" -v -o $MainMob $MainAsm
if ($LASTEXITCODE -ne 0) { throw "Falha na montagem de main.asm" }

Write-Host "[2/4] Montando modulo Banco 1 (Pagina 2): bank1.asm -> bank1.mob..." -ForegroundColor Yellow
go run "$RootDir/cmd/kaji80" -v -o $Bank1Mob $Bank1Asm
if ($LASTEXITCODE -ne 0) { throw "Falha na montagem de bank1.asm" }

Write-Host "[3/4] Montando modulo Banco 2 (Pagina 2): bank2.asm -> bank2.mob..." -ForegroundColor Yellow
go run "$RootDir/cmd/kaji80" -v -o $Bank2Mob $Bank2Asm
if ($LASTEXITCODE -ne 0) { throw "Falha na montagem de bank2.asm" }

Write-Host ""

# ------------------------------------------------------------------------------
# ETAPA 2: Linkagem Multi-Banco via MUSUBI
# ------------------------------------------------------------------------------
Write-Host "[4/4] Linkando modulos e gerando executavel MSX-DOS 2 com MUSUBI..." -ForegroundColor Yellow
Write-Host "      Entrada: main.mob (B0) + bank1.mob (B1) + bank2.mob (B2)" -ForegroundColor Gray
Write-Host "      Saida:   multibank.com + multibank.map" -ForegroundColor Gray

go run "$RootDir/cmd/musubi" -v -m $OutputMap -o $OutputCom $MainMob $Bank1Mob $Bank2Mob
if ($LASTEXITCODE -ne 0) { throw "Falha na linkagem com musubi" }

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Green
Write-Host "       BUILD CONCLUIDO COM SUCESSO!                             " -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Green

$FileInfo = Get-Item $OutputCom
Write-Host "Executavel Gerado: $($FileInfo.FullName)" -ForegroundColor White
Write-Host "Tamanho do .COM:   $($FileInfo.Length) bytes" -ForegroundColor White
Write-Host "Mapa de Memoria:   $OutputMap" -ForegroundColor White
Write-Host ""
Write-Host "Para executar no MSX-DOS 2:" -ForegroundColor Cyan
Write-Host "  Copie o arquivo 'multibank.com' para o seu cartucho/disquete e rode: MULTIBANK" -ForegroundColor Gray
Write-Host ""

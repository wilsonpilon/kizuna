# ==============================================================================
# KIZUNA (絆) - Script Mestre de Build e Empacotamento de Distribuição
# ==============================================================================
# Este script:
# 1. Compila os binários da toolchain (kaji80, musubi, mobdump) e o instalador TUI.
# 2. Prepara a pasta 'distribute/' com binários, amostras limpas e documentação.
# 3. Compacta o pacote em 'kizuna-vX.Y.Z-dist.zip' pronto para distribuição no GitHub.
# ==============================================================================

$ErrorActionPreference = "Stop"

$RootDir = $PSScriptRoot
if (-not $RootDir) {
    $RootDir = (Get-Location).Path
}

# 1. Obter a versão e codinome dinamicamente via Musubi
$VersionOutput = & go run "$RootDir/cmd/musubi" --version 2>&1
$VersionMatch = [regex]::Match($VersionOutput, "v(\d+\.\d+\.\d+)")
if ($VersionMatch.Success) {
    $KizunaVersion = $VersionMatch.Groups[1].Value
} else {
    $KizunaVersion = "4.1.0"
}

$DistDir = Join-Path $RootDir "distribute"
$ZipFileName = "kizuna-v$KizunaVersion-dist.zip"
$ZipFilePath = Join-Path $RootDir $ZipFileName

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "    KIZUNA (絆) - Build & Empacotamento para Distribuicao        " -ForegroundColor Cyan
Write-Host "    Versao: v$KizunaVersion [Release Akatsuki (暁)]              " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "Diretorio Raiz: $RootDir" -ForegroundColor Gray
Write-Host "Destino:        $DistDir" -ForegroundColor Gray
Write-Host "Pacote Zip:     $ZipFileName" -ForegroundColor Gray
Write-Host ""

# 2. Limpar diretório de distribuição anterior
if (Test-Path $DistDir) {
    Write-Host "[1/6] Limpando pasta distribute/ antiga..." -ForegroundColor Yellow
    Remove-Item -Path $DistDir -Recurse -Force
}

$BinDir  = Join-Path $DistDir "bin"
$DocsDir = Join-Path $DistDir "docs"
$SampDir = Join-Path $DistDir "sample"

New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
New-Item -ItemType Directory -Path $DocsDir -Force | Out-Null
New-Item -ItemType Directory -Path $SampDir -Force | Out-Null

# 3. Compilar executáveis da toolchain
Write-Host "[2/6] Compilando executaveis da toolchain (Go -> .exe)..." -ForegroundColor Yellow

$Tools = @("kaji80", "musubi", "mobdump")
foreach ($t in $Tools) {
    $outExe = Join-Path $BinDir "$t.exe"
    Write-Host "      - Compilando $t -> $outExe..." -ForegroundColor Gray
    & go build -ldflags="-s -w" -o $outExe "$RootDir/cmd/$t"
    if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar $t" }
}

# 4. Compilar instalador interativo TUI
Write-Host "[3/6] Compilando instalador TUI (install.exe)..." -ForegroundColor Yellow
$InstallExe = Join-Path $DistDir "install.exe"
& go build -ldflags="-s -w" -o $InstallExe "$RootDir/cmd/installer"
if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar o instalador TUI" }

# Cria um atalho .cmd conveniente para abrir com 2 cliques
$InstallCmd = Join-Path $DistDir "install.cmd"
Set-Content -Path $InstallCmd -Value "@echo off`r`ncd /d `"%~dp0`"`r`ninstall.exe`r`npause`r`n" -Encoding ASCII

# 5. Copiar documentação essencial e licença
Write-Host "[4/6] Copiando documentacao de usuario e licenca..." -ForegroundColor Yellow
Copy-Item -Path "$RootDir/README.md" -Destination "$DocsDir/README.md"
Copy-Item -Path "$RootDir/HELP.md" -Destination "$DocsDir/HELP.md"
Copy-Item -Path "$RootDir/CHANGELOG.md" -Destination "$DocsDir/CHANGELOG.md"
Copy-Item -Path "$RootDir/LICENSE" -Destination "$DistDir/LICENSE"

# 6. Copiar exemplos (limpos e organizados)
Write-Host "[5/6] Copiando exemplos de codigo (sample/)..." -ForegroundColor Yellow
Copy-Item -Path "$RootDir/sample/hello.asm" -Destination "$SampDir/hello.asm"

$MultiSampDir = Join-Path $SampDir "multibank"
New-Item -ItemType Directory -Path $MultiSampDir -Force | Out-Null
Copy-Item -Path "$RootDir/sample/multibank/main.asm" -Destination "$MultiSampDir/main.asm"
Copy-Item -Path "$RootDir/sample/multibank/bank1.asm" -Destination "$MultiSampDir/bank1.asm"
Copy-Item -Path "$RootDir/sample/multibank/bank2.asm" -Destination "$MultiSampDir/bank2.asm"
Copy-Item -Path "$RootDir/sample/multibank/build.ps1" -Destination "$MultiSampDir/build.ps1"

# 7. Gerar pacote compactado .ZIP
Write-Host "[6/6] Criando arquivo compactado $ZipFileName..." -ForegroundColor Yellow
if (Test-Path $ZipFilePath) {
    Remove-Item -Path $ZipFilePath -Force
}

# Compacta os arquivos contidos dentro de distribute/ para a raiz do .zip
Compress-Archive -Path "$DistDir\*" -DestinationPath $ZipFilePath -CompressionLevel Optimal

$ZipInfo = Get-Item $ZipFilePath

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Green
Write-Host "       EMPACOTAMENTO CONCLUIDO COM SUCESSO!                     " -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Green
Write-Host "Diretorio de Distribuicao: $DistDir" -ForegroundColor White
Write-Host "Arquivo Zip Gerado:        $ZipFilePath" -ForegroundColor White
Write-Host "Tamanho do Pacote:         $([math]::Round($ZipInfo.Length / 1KB, 2)) KB" -ForegroundColor White
Write-Host ""
Write-Host "Conteudo do pacote distribute/:" -ForegroundColor Cyan
Get-ChildItem -Path $DistDir -Recurse | Select-Object FullName | ForEach-Object {
    $rel = $_.FullName.Replace($DistDir, "")
    Write-Host "  .$rel" -ForegroundColor Gray
}
Write-Host ""

# ==============================================================================
# KIZUNA (絆) - Script Mestre de Build e Empacotamento de Distribuição
# ==============================================================================
# Este script:
# 1. Compila os binários da toolchain (kaji80, musubi, mobdump, hako, wirth80, dignac)
#    e o instalador TUI.
# 2. Constrói a biblioteca padrão MSXLIB (msxlib.hlib).
# 3. Compila todos os programas de exemplo em sample/ (Assembly, Pascal e BASIC).
# 4. Prepara a pasta 'distribute/' com binários, documentação e amostras já compiladas (.COM).
# 5. Compacta o pacote em 'kizuna-vX.Y.Z-dist.zip' pronto para distribuição.
# 6. Exibe relatório detalhado com a localização de cada executável compilado.
# ==============================================================================

$ErrorActionPreference = "Stop"

$RootDir = $PSScriptRoot
if (-not $RootDir) {
    $RootDir = (Get-Location).Path
}

# 1. Obter a versão dinamicamente via Musubi
$VersionOutput = & go run "$RootDir/cmd/musubi" --version 2>&1
$VersionMatch = [regex]::Match($VersionOutput, "v(\d+\.\d+\.\d+)")
if ($VersionMatch.Success) {
    $KizunaVersion = $VersionMatch.Groups[1].Value
} else {
    $KizunaVersion = "4.5.0"
}

$DistDir = Join-Path $RootDir "distribute"
$ZipFileName = "kizuna-v$KizunaVersion-dist.zip"
$ZipFilePath = Join-Path $RootDir $ZipFileName

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "    KIZUNA (絆) - Build & Empacotamento para Distribuicao        " -ForegroundColor Cyan
Write-Host "    Versao: v$KizunaVersion [Release Hinode (日の出)]                " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "Diretorio Raiz: $RootDir" -ForegroundColor Gray
Write-Host "Destino:        $DistDir" -ForegroundColor Gray
Write-Host "Pacote Zip:     $ZipFileName" -ForegroundColor Gray
Write-Host ""

# 2. Limpar diretório de distribuição anterior
if (Test-Path $DistDir) {
    Write-Host "[1/8] Limpando pasta distribute/ antiga..." -ForegroundColor Yellow
    Remove-Item -Path $DistDir -Recurse -Force
}

$BinDir  = Join-Path $DistDir "bin"
$DocsDir = Join-Path $DistDir "docs"
$SampDir = Join-Path $DistDir "sample"
$LibDistDir = Join-Path $DistDir "lib"

New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
New-Item -ItemType Directory -Path $DocsDir -Force | Out-Null
New-Item -ItemType Directory -Path $SampDir -Force | Out-Null
New-Item -ItemType Directory -Path $LibDistDir -Force | Out-Null

# 3. Compilar executáveis da toolchain (Go -> .exe)
Write-Host "[2/8] Compilando executaveis da toolchain (Go -> .exe)..." -ForegroundColor Yellow

$Tools = @("kaji80", "musubi", "mobdump", "hako", "wirth80", "dignac")
foreach ($t in $Tools) {
    $outExe = Join-Path $BinDir "$t.exe"
    Write-Host "      - Compilando $t -> $outExe..." -ForegroundColor Gray
    & go build -ldflags="-s -w" -o $outExe "$RootDir/cmd/$t"
    if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar $t" }
}

# 4. Compilar instalador interativo TUI
Write-Host "[3/8] Compilando instalador TUI (install.exe)..." -ForegroundColor Yellow
$InstallExe = Join-Path $DistDir "install.exe"
& go build -ldflags="-s -w" -o $InstallExe "$RootDir/cmd/installer"
if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar o instalador TUI" }

# Cria atalho .cmd conveniente
$InstallCmd = Join-Path $DistDir "install.cmd"
Set-Content -Path $InstallCmd -Value "@echo off`r`ncd /d `"%~dp0`"`r`ninstall.exe`r`npause`r`n" -Encoding ASCII

# 5. Construir a Biblioteca Padrão MSXLIB
Write-Host "[4/8] Construindo Biblioteca Padrao MSXLIB (msxlib.hlib)..." -ForegroundColor Yellow
& pwsh -ExecutionPolicy Bypass -File "$RootDir/lib/build.ps1"
if ($LASTEXITCODE -ne 0) { throw "Falha ao construir a biblioteca MSXLIB" }
Copy-Item -Path "$RootDir/lib/msxlib.hlib" -Destination "$LibDistDir/msxlib.hlib"

# 6. Compilar todos os programas de exemplo em sample/
Write-Host "[5/8] Compilando todos os programas de exemplo em sample/..." -ForegroundColor Yellow

# 6.1. Exemplo Hello (Assembly puro)
Write-Host "      - Compilando sample/hello.asm -> sample/hello.com..." -ForegroundColor Gray
& go run "$RootDir/cmd/kaji80" -o "$RootDir/sample/hello.mob" "$RootDir/sample/hello.asm"
& go run "$RootDir/cmd/musubi" -o "$RootDir/sample/hello.com" -m "$RootDir/sample/hello.map" "$RootDir/sample/hello.mob"
if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar sample/hello.asm" }

# 6.2. Exemplo Libdemo (Assembly + MSXLIB)
Write-Host "      - Compilando sample/libdemo/main.asm -> sample/libdemo/libdemo.com..." -ForegroundColor Gray
& pwsh -ExecutionPolicy Bypass -File "$RootDir/sample/libdemo/build.ps1"
if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar sample/libdemo" }

# 6.3. Exemplo Multibank (Assembly Multi-Banco)
Write-Host "      - Compilando sample/multibank -> sample/multibank/multibank.com..." -ForegroundColor Gray
& pwsh -ExecutionPolicy Bypass -File "$RootDir/sample/multibank/build.ps1"
if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar sample/multibank" }

# 6.4. Exemplos Pascal WIRTH80 (hello.pas, calc.pas)
Write-Host "      - Compilando sample/pascal (hello.pas, calc.pas)..." -ForegroundColor Gray
& pwsh -ExecutionPolicy Bypass -File "$RootDir/sample/pascal/build.ps1"
if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar sample/pascal" }

# 6.5. Exemplos BASIC Dignified DIGNAC (hello.bas, calc.bas, chart.bas)
Write-Host "      - Compilando sample/basic (hello.bas, calc.bas, chart.bas)..." -ForegroundColor Gray
& pwsh -ExecutionPolicy Bypass -File "$RootDir/sample/basic/build.ps1"
if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar sample/basic" }

# 7. Copiar documentação essencial e licença
Write-Host "[6/8] Copiando documentacao de usuario e licenca..." -ForegroundColor Yellow
Copy-Item -Path "$RootDir/README.md" -Destination "$DocsDir/README.md"
Copy-Item -Path "$RootDir/HELP.md" -Destination "$DocsDir/HELP.md"
Copy-Item -Path "$RootDir/CHANGELOG.md" -Destination "$DocsDir/CHANGELOG.md"
Copy-Item -Path "$RootDir/LICENSE" -Destination "$DistDir/LICENSE"

# 8. Copiar exemplos para distribute/sample/ com executáveis .COM e fontes
Write-Host "[7/8] Copiando exemplos (fontes e .COM compilados) para distribute/sample/..." -ForegroundColor Yellow

# Hello (Assembly)
Copy-Item -Path "$RootDir/sample/hello.asm" -Destination "$SampDir/hello.asm"
Copy-Item -Path "$RootDir/sample/hello.com" -Destination "$SampDir/hello.com"

# Multibank
$MultiSampDir = Join-Path $SampDir "multibank"
New-Item -ItemType Directory -Path $MultiSampDir -Force | Out-Null
Copy-Item -Path "$RootDir/sample/multibank/main.asm" -Destination "$MultiSampDir/main.asm"
Copy-Item -Path "$RootDir/sample/multibank/bank1.asm" -Destination "$MultiSampDir/bank1.asm"
Copy-Item -Path "$RootDir/sample/multibank/bank2.asm" -Destination "$MultiSampDir/bank2.asm"
Copy-Item -Path "$RootDir/sample/multibank/build.ps1" -Destination "$MultiSampDir/build.ps1"
Copy-Item -Path "$RootDir/sample/multibank/multibank.com" -Destination "$MultiSampDir/multibank.com"

# Libdemo
$LibDemoDir = Join-Path $SampDir "libdemo"
New-Item -ItemType Directory -Path $LibDemoDir -Force | Out-Null
Copy-Item -Path "$RootDir/sample/libdemo/main.asm" -Destination "$LibDemoDir/main.asm"
Copy-Item -Path "$RootDir/sample/libdemo/build.ps1" -Destination "$LibDemoDir/build.ps1"
Copy-Item -Path "$RootDir/sample/libdemo/libdemo.com" -Destination "$LibDemoDir/libdemo.com"

# Pascal
$PascalDir = Join-Path $SampDir "pascal"
New-Item -ItemType Directory -Path $PascalDir -Force | Out-Null
Copy-Item -Path "$RootDir/sample/pascal/hello.pas" -Destination "$PascalDir/hello.pas"
Copy-Item -Path "$RootDir/sample/pascal/hello.com" -Destination "$PascalDir/hello.com"
Copy-Item -Path "$RootDir/sample/pascal/calc.pas" -Destination "$PascalDir/calc.pas"
Copy-Item -Path "$RootDir/sample/pascal/calc.com" -Destination "$PascalDir/calc.com"
Copy-Item -Path "$RootDir/sample/pascal/build.ps1" -Destination "$PascalDir/build.ps1"

# BASIC
$BasicDir = Join-Path $SampDir "basic"
New-Item -ItemType Directory -Path $BasicDir -Force | Out-Null
Copy-Item -Path "$RootDir/sample/basic/hello.bas" -Destination "$BasicDir/hello.bas"
Copy-Item -Path "$RootDir/sample/basic/hello.com" -Destination "$BasicDir/hello.com"
Copy-Item -Path "$RootDir/sample/basic/calc.bas" -Destination "$BasicDir/calc.bas"
Copy-Item -Path "$RootDir/sample/basic/calc.com" -Destination "$BasicDir/calc.com"
Copy-Item -Path "$RootDir/sample/basic/chart.bas" -Destination "$BasicDir/chart.bas"
Copy-Item -Path "$RootDir/sample/basic/chart.com" -Destination "$BasicDir/chart.com"
Copy-Item -Path "$RootDir/sample/basic/build.ps1" -Destination "$BasicDir/build.ps1"

# 9. Gerar pacote compactado .ZIP
Write-Host "[8/8] Criando arquivo compactado $ZipFileName..." -ForegroundColor Yellow
if (Test-Path $ZipFilePath) {
    Remove-Item -Path $ZipFilePath -Force
}

Compress-Archive -Path "$DistDir\*" -DestinationPath $ZipFilePath -CompressionLevel Optimal

$ZipInfo = Get-Item $ZipFilePath

# Remove a pasta temporaria distribute/ para que exista apenas UMA pasta sample/ no projeto (sample/)
if (Test-Path $DistDir) {
    Remove-Item -Path $DistDir -Recurse -Force
}

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Green
Write-Host "       EMPACOTAMENTO CONCLUIDO COM SUCESSO!                     " -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Green
Write-Host "Arquivo Zip Gerado: $ZipFilePath" -ForegroundColor White
Write-Host "Tamanho do Pacote:  $([math]::Round($ZipInfo.Length / 1KB, 2)) KB" -ForegroundColor White
Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "   EXECUTAVEIS COMPILADOS (.COM) OFICIAIS (EM sample/):          " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

$kizunaComs = @(
    "sample/hello.com",
    "sample/basic/hello.com",
    "sample/basic/calc.com",
    "sample/basic/chart.com",
    "sample/pascal/hello.com",
    "sample/pascal/calc.com",
    "sample/libdemo/libdemo.com",
    "sample/multibank/multibank.com"
)
foreach ($rel in $kizunaComs) {
    $full = Join-Path $RootDir $rel
    if (Test-Path $full) {
        $f = Get-Item $full
        Write-Host "  -> $rel" -ForegroundColor Green
        Write-Host "     Tamanho: $($f.Length) bytes | Modificado: $($f.LastWriteTime.ToString('yyyy-MM-dd HH:mm:ss'))" -ForegroundColor Gray
    }
}
Write-Host ""

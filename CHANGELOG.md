# Changelog

Todas as mudanças notáveis deste projeto são documentadas aqui.
Formato baseado em [Keep a Changelog](https://keepachangelog.com/).

### Política de Versionamento (`MAJOR.MINOR.COMPILAÇÃO`)
- **MAJOR**: Incrementado a cada encerramento de fase da toolchain (ex.: Fase 1 = MOB, Fase 2 = KAJI80, Fase 3 = MUSUBI monobanco, Fase 4 = Multi-banco & Memory Mapper).
- **MINOR**: Incrementado a cada feature ou subsistema novo adicionado.
- **COMPILAÇÃO (BUILD)**: Incrementado a cada compilação / build realizado no projeto.

## [4.5.1] - 2026-09-04 - Release Kuyashii (悔しい)

### Modificado & Corrigido
- **MSX-BASIC Dignified DIGNAC (`pkg/dignac`) e MSXLIB (`lib/src/`)**:
  - **Exportação de Símbolos**: Exportação das rotinas `VDP_PSet`, `VDP_Line`, `VDP_BoxFill` e `VDP_InitScreen2_Tables` na diretiva `PUBLIC` de `lib/src/vdp.asm`, permitindo resolução completa e correta pelo smart-linker `musubi`.
  - **Suporte de Vídeo e BIOS ([lib/src/bios.asm](lib/src/bios.asm))**:
    - Ajuste em `BIOS_CHGMOD`: integração com `INIGRP (0072h)` da Main-ROM BIOS para SCREEN 2 e rotinas de tabela VRAM (`VDP_InitScreen2_Tables`).
    - Integração com `INITXT (006Ch)` e `CHGCLR (0062h)` na saída para SCREEN 0, garantindo restauração da fonte de caracteres ROM e das cores de tela ao retornar ao MSX-DOS.
    - Temporizador em `BIOS_CHGET` com dreno de caracteres residuais e saída antecipada via tecla `ESC`.
- **Estado dos Testes Gráficos**:
  - O pipeline de montagem e linkagem do exemplo gráfico (`sample/basic/chart.bas` -> `chart.com`) compila e empacota perfeitamente de ponta a ponta.
  - A renderização em tela real/emulador sob o MSX-DOS 2 em hardware MSX2+ permanece em investigação para ajustes de inicialização de VRAM/VDP (motivo do codinome *Kuyashii (悔しい)* — expressando o sentimento de frustração respeitosa e determinação para a próxima sessão).

---

## [4.5.0] - 2026-09-04 - Compilador MSX-BASIC Dignified DIGNAC

### Adicionado
- **Compilador MSX-BASIC Dignified DIGNAC (`pkg/dignac` e `cmd/dignac`) - Fase 5.3**:
  - Nova ferramenta da toolchain para compilação de código BASIC estruturado para MSX2+ / MSX-DOS 2, emitindo objetos relocáveis `.MOB`.
  - **Modularidade e Paginação**: Suporte a diretivas `MODULE`, `BANK`, `PUBLIC` e `EXTERN`.
  - **Sub-rotinas e ABI Kizuna**: Suporte a `PROCEDURE`, `SUB`, `FUNCTION`, passagem de parâmetros em pilha, alocação de quadro de pilha com frame pointer `IX` e variáveis locais (`LOCAL x%, y%`).
  - **Ponto de Entrada Automático**: Geração automática do ponto de entrada `Start` chamando `PROCEDURE Main` e finalizando com `BDOS_Exit`, permitindo linkagem direta de `.COM` standalone.
  - **Controle de Fluxo Estruturado**: Laços `FOR ... TO ... STEP ... NEXT`, `WHILE ... WEND`, `DO ... LOOP`, e condicionais `IF ... THEN ... ELSE ... END IF` (em linha única ou blocos).
  - **Primitivas Gráficas**: Comandos `LINE (x1,y1)-(x2,y2)[, color][, B | BF]` e `PSET (x, y)[, color]`.
  - **Expressões e Aritmética de 16 bits**: Operações `+`, `-`, `*` (via `Mul16`), `/` e operador `MOD` (via `Div16`), além de operadores lógicos `AND`, `OR`, `XOR`, `NOT` e comparações relacionais (`=`, `<>`, `<`, `<=`, `>`, `>=`).
  - **CLI `dignac`**: Suporte às flags `-o <saida.mob>`, `-S` (código assembly intermediário formatado), `-v` (modo detalhado) e `--version`.
- **Endereçamento Indexado Z80 no Assembler KAJI80 ([pkg/kaji80/assembler.go](pkg/kaji80/assembler.go))**:
  - Suporte completo a instruções `LD r, (IX+d)` e `LD (IX+d), r` para registradores de 8 bits.
  - Suporte a `LD (IX+d), n` (imediato).
  - Suporte análogo para o registrador de índice `IY` (`0xFD`).
  - Suporte a `LD SP, IX` (`DD F9`) e `LD SP, IY` (`FD F9`).
- **Primitivas Gráficas na Biblioteca Padrão MSXLIB ([lib/src/vdp.asm](lib/src/vdp.asm))**:
  - `VDP_PSet`: Cálculo de endereços nas tabelas de padrão e cor em SCREEN 2 (TMS9918/V9938) e plotagem de pixel.
  - `VDP_BoxFill`: Preenchimento de retângulos e aceleração para limpeza de tela cheia via `VDP_FillVRAM`.
  - `VDP_Line`: Traçado de segmentos de reta horizontais e diagonais.
- **Exemplos em BASIC Dignified (`sample/basic/`)**:
  - `hello.bas`: Hello World em BASIC compilado (apenas 186 bytes).
  - `calc.bas`: Demonstração de variáveis locais, aritmética de 16 bits, divisão, módulo e exibição decimal.
  - `chart.bas`: Módulo paginável no banco 2 para traçado de curvas.
  - `build.ps1`: Automação de compilação e smart-linking dos exemplos BASIC.
- **Empacotamento Global**:
  - `dignac.exe` integrado ao script mestre [build.ps1](build.ps1) e empacotado em `distribute/bin/`.
  - Pasta `sample/basic/` incluída na distribuição oficial.

---

## [4.4.0] - 2026-09-04 - Compilador Pascal WIRTH80

### Adicionado
- **Compilador Pascal WIRTH80 (`pkg/wirth80` e `cmd/wirth80`)**:
  - Nova ferramenta da toolchain para compilação de código-fonte Pascal nativo para MSX2+ / MSX-DOS 2, gerando módulos objeto no formato `.MOB`.
  - **Análise Léxica e Sintática**: Suporte a identificadores, números (decimais e `$hex`), strings literais (`'...'`), comentários (`{ ... }`, `(* ... *)`, `// ...`), `program`, `var`, `begin ... end.`.
  - **Tipos de Dados**: Suporte a `Integer` (16 bits sinalizado) e `Char` / `Boolean` (8 bits).
  - **Comandos e Controle de Fluxo**: Atribuição (`:=`), `Write`, `WriteLn`, condicionais `if ... then ... else` e laços `while ... do`.
  - **Expressões**: Aritmética de 16 bits (`+`, `-`, `*` via `Mul16`, `div` via `Div16`), unários e comparações relacionais (`=`, `<>`, `<`, `<=`, `>`, `>=`).
  - **Integração com MSXLIB**: Geração automática de chamadas externas para `BDOS_PrintString`, `BDOS_PrintChar`, `PrintDec16`, `Mul16`, `Div16` e `BDOS_Exit`.
  - **CLI `wirth80`**: Suporte às flags `-o <saida.mob>`, `-S` (emissão de código assembly Z80 legível), `-v` (modo detalhado) e `--version`.
- **Exemplos em Pascal (`sample/pascal/`)**:
  - `hello.pas`: Exemplo clássico Hello World em Pascal para MSX.
  - `calc.pas`: Demonstração de declaração de variáveis, multiplicação de 16 bits e exibição de números decimais.
  - `build.ps1`: Script de automação para compilação e smart-linking dos exemplos Pascal.
- **Empacotamento Global**:
  - `wirth80.exe` integrado ao script mestre [build.ps1](build.ps1) e empacotado em `distribute/bin/`.
  - Pasta `sample/pascal/` incluída na distribuição oficial.

---

## [4.3.3] - 2026-09-04

### Corrigido
- **Suporte Aritmético de 16 bits no Assembler KAJI80 ([pkg/kaji80/assembler.go](pkg/kaji80/assembler.go))**:
  - Implementada a codificação e estimativa precisa de `SBC HL, ss` (`ED 42/52/62/72`) e `ADC HL, ss` (`ED 4A/5A/6A/7A`) de 16 bits.
  - Anteriormente, `SBC HL, DE` caía no caso ALU de 8 bits e montava incorretamente como `0xDE 0x00` (`SBC A, 00h`). Isso fazia o loop de cálculo de dígitos em `PrintDecDigit` entrar em um loop infinito no Z80 sem subtrair HL, travando a máquina antes de imprimir qualquer caractere e impedindo a execução de prosseguir para o som e retorno ao DOS.
  - Implementado suporte a `LD HL, (nn)` (`0x2A`) e `LD (nn), HL` (`0x22`) no KAJI80.
- **Ajuste de Duração do Loop de Pausa de Áudio ([sample/libdemo/main.asm](sample/libdemo/main.asm))**:
  - Reduzido o contador do loop externo `LD B, 18h` (que durava ~11.5 segundos) para `LD B, 02h` (~1 segundo audível), tornando a execução e retorno ao DOS fluídos e imediatos.
- **Validação Completa via Emulação de CPU**:
  - Execução validada instrução-a-instrução em simulador Z80, comprovando o cálculo de `123 x 45 = 5535`, ativação correta do PSG e encerramento com sucesso via BDOS 00h.

---

## [4.3.2] - 2026-09-04

### Corrigido
- **Sincronização de Offsets no Assembler KAJI80 ([pkg/kaji80/assembler.go](pkg/kaji80/assembler.go))**:
  - Corrigida a estimativa de tamanho no Pass 1 para instruções ALU imediatas de 1 operando (ex: `CP n`, `SUB n`, `AND n`), que retornava 1 byte em vez de 2 bytes.
  - Corrigida a estimativa de tamanho para `LD (DE), A`, `LD (BC), A`, `LD A, (DE)` e `LD A, (BC)` que retornava 3 bytes em vez de 1 byte.
  - A discrepância acumulada causava desvio de endereço nos símbolos de `string.mob`, fazendo o linker `musubi` calcular chamadas como `PrintDec16` com desvio de 2 a 3 bytes no meio de outra instrução.
- **Canal de Som do PSG ([lib/src/psg.asm](lib/src/psg.asm))**:
  - Inserido `LD A, C` após `INC C` no registrador grosso de período da frequência.
  - Configuração exata do misturador do Registrador 7: ativa Canal A (`0xBE`), Canal B (`0xBD`) e Canal C (`0xBB`) sem afetar as portas de joystick e teclado.
- **Pausa Audível e Frequência no Demo ([sample/libdemo/main.asm](sample/libdemo/main.asm))**:
  - Ajustada a frequência para A440 (`0x00FE`), volume máximo 15 (`0x0F`) e loop de pausa aninhado audível (~0.5s).

---

## [4.3.1] - 2026-09-03

### Corrigido
- **Conversão Decimal em `PrintDec16` ([lib/src/string.asm](lib/src/string.asm))**:
  - Reescrita da rotina de conversão para utilizar subtração sucessiva direta de potências de 10 sem endereçamento indexado `(IX+d)`, eliminando falha que impedia a exibição do resultado numérico.
- **Preservação de Registradores em Chamadas BDOS ([lib/src/bdos.asm](lib/src/bdos.asm))**:
  - `BDOS_PrintChar` e `BDOS_PrintString` agora salvam e restauram explicitamente todos os registradores (`AF`, `BC`, `DE`, `HL`, `IX`, `IY`), prevenindo corrupção de registradores durante chamadas de console do MSX-DOS.
- **Controle Seguro do Misturador PSG ([lib/src/psg.asm](lib/src/psg.asm))**:
  - `PSG_PlayTone` e `PSG_MuteAll` agora gravam valores fixos e seguros no Registrador 7 (`0xB8` / `0xBF`), eliminando leituras na porta `0xA2` e preservando intactas as direções das portas de I/O (joystick e teclado) do MSX.
- **Lexer do KAJI80 ([pkg/kaji80/lexer.go](pkg/kaji80/lexer.go))**:
  - Ajustada a precedência na análise de literais para verificar o sufixo `H` antes do prefixo `0b`, corrigindo erro de montagem em constantes hexadecimais iniciadas em `0B` (ex: `0BFh`, `0B8h`).

---

## [4.3.0] - 2026-09-03 - Biblioteca Padrão MSXLIB

### Adicionado
- **Biblioteca Padrão MSXLIB (`lib/src/` e `lib/msxlib.hlib`)**:
  - Pacote de 6 módulos modulares escritos em Z80 Assembly e empacotados em arquivo de biblioteca `.HLIB`:
    - `bdos`: Rotinas de chamada ao kernel MSX-DOS (`BDOS_Call`, `BDOS_PrintChar`, `BDOS_PrintString`, `BDOS_ReadChar`, `BDOS_Exit`).
    - `bios`: Chamadas inter-slot seguras à Main-ROM via `CALSLT` (`BIOS_Call`, `BIOS_CHPUT`, `BIOS_CHGET`, `BIOS_CLS`, `BIOS_POSIT`, `BIOS_BEEP`, `BIOS_INIT32`).
    - `vdp`: Controle de portas e registradores do processador de vídeo TMS9918/V9938/V9958 (`VDP_WriteReg`, `VDP_SetWriteAddr`, `VDP_SetReadAddr`, `VDP_FillVRAM`, `VDP_WriteVRAM`, `VDP_ReadVRAM`, `VDP_SetColor`).
    - `psg`: Manipulação do gerador de som programável AY-3-8910 (`PSG_Write`, `PSG_Read`, `PSG_MuteAll`, `PSG_PlayTone`).
    - `string`: Funções de texto e conversão numérica rápida com supressão de zeros (`StrLen`, `StrCopy`, `StrToUpper`, `PrintHex8`, `PrintHex16`, `PrintDec16`).
    - `math`: Multiplicação e divisão inteira não sinalizada de 16 bits (`Mul16`, `Div16`).
- **Suporte a Novas Instruções Z80 no Assembler KAJI80 (`pkg/kaji80`)**:
  - Instruções de rotação e deslocamento: `RLCA`, `RRCA`, `RLA`, `RRA`, `RLC`, `RRC`, `RL`, `RR`, `SLA`, `SRA`, `SRL`.
  - Operações bit-a-bit: `BIT`, `RES`, `SET`.
  - Flags e complemento aritmético: `CPL`, `SCF`, `CCF`, `NEG`.
- **Exemplo Prático com Smart-Linking ([sample/libdemo](sample/libdemo))**:
  - Exemplo demonstrativo em Assembly importando e executando rotinas da `msxlib.hlib`, comprovando eliminação de código morto no mapa de memória.
- **Integração no Script de Distribuição ([build.ps1](build.ps1))**:
  - Geração automática de `msxlib.hlib` e inclusão em `distribute/lib/` e no pacote ZIP de distribuição.

---

## [4.2.0] - 2026-09-03 - HAKO & Smart-Linking

### Adicionado
- **Bibliotecário de Objetos HAKO (`pkg/hako` e `cmd/hako`)**:
  - Implementação do utilitário `hako` (箱) para gerenciamento de arquivos de biblioteca de objetos relocáveis.
  - Especificação e codificação do formato binário de arquivo de biblioteca **`.HLIB`**:
    - Cabeçalho de 14 bytes com Magic `"HLIB"`, versão do formato (`1`), contagem de módulos e apontador para o dicionário global de símbolos.
    - Tabela de módulos com nomes, offsets e tamanhos dos arquivos `.MOB` brutos embutidos.
    - Dicionário global de símbolos públicos (`PUBLIC`) mapeando nomes para módulos de origem.
  - Validação na criação de bibliotecas: rejeita símbolos públicos duplicados entre módulos distintos com mensagem de erro clara.
  - Interface de linha de comando com opções `-c` (criar/empacotar), `-t` (listar conteúdo e símbolos), `-x` (extrair módulo), `-v` (detalhado) e `--version`.
- **Smart-Linking e Dead-Code Elimination no Linker MUSUBI**:
  - Suporte a inclusão de bibliotecas estáticas `.hlib` diretamente na linha de comando do `musubi` (`musubi main.mob math.hlib`).
  - Resolução transitiva inteligente de símbolos: apenas os módulos do `.hlib` contendo símbolos realmente referenciados por `main` ou por outros módulos ativos são extraídos e incluídos no binário final.
  - Eliminação de código morto (*dead-code elimination*): módulos não referenciados no `.hlib` são descartados, mantendo o executável `.com` no menor tamanho possível.
- **Atualização do Script Mestre de Build ([build.ps1](build.ps1))**:
  - Adicionado `hako` à lista de ferramentas compiladas automaticamente para `distribute/bin/hako.exe` e incluído no pacote `.zip`.
- **Documentação de Referência ([HELP.md](HELP.md))**:
  - Adicionada Seção 10 documentando o utilitário `hako`, o formato binário `.HLIB` e os comandos de empacotamento e listagem.

---

## [4.1.0] - 2026-09-03 - Release "Akatsuki" (暁 - Alvorecer)

### Adicionado
- **Script Mestre de Build e Empacotamento ([build.ps1](build.ps1))**: Automação completa para compilação da toolchain, preparação do diretório `distribute/` e geração do arquivo compactado `kizuna-v4.1.0-dist.zip`.
- **Instalador Interativo TUI ([cmd/installer/main.go](cmd/installer/main.go))**: Utilitário em modo texto com identidade visual japonesa, menu interativo, instalação automática dos binários, configuração automática da variável `PATH` do usuário no Windows e rotina de teste/validação da toolchain.
- **Estrutura de Distribuição Organizada (`distribute/`)**:
  - `bin/`: Executáveis pré-compilados (`kaji80.exe`, `musubi.exe`, `mobdump.exe`).
  - `docs/`: Documentação essencial para o usuário final (`README.md`, `HELP.md`, `CHANGELOG.md`).
  - `sample/`: Exemplos limpos e prontos para teste (`sample/hello.asm` e `sample/multibank/`).
  - `install.exe` e atalho `install.cmd` para instalação com 2 cliques.
  - `LICENSE`: Licença do projeto.
- **Codinome Oficial da Release**: Batizada como **"Akatsuki" (暁 - Alvorecer)**, simbolizando o início de uma nova era para desenvolvimento no MSX2+.

---

## [4.0.1] - 2026-09-03

### Adicionado
- **Suporte Multi-Banco no Linker MUSUBI**: Alocação de módulos em bancos pagináveis (1..N) na Página 2 (`0x8000..0xBFFF`).
- **Bootstrap Loader Automatizado**: Inserção automática de código de carga em `0x0100` que transfere os bancos embutidos no `.COM` para os segmentos da RAM expandida com indicador de progresso `[L]`.
- **Alinhamento de Slot via Porta 0xA8**: Configuração automática no bootstrap para que a Página 2 assuma o mesmo slot primário da Página 1 (RAM do TPA), assegurando compatibilidade com cartuchos externos de MSX-DOS 2.
- **Descoberta Dinâmica de Tabela via EXTBIO**: Conexão oficial ao kernel do MSX-DOS 2 chamando `EXTBIO` (`0xFFCA`, Device ID 4) para obter os pontos de entrada oficiais de `PUT_P2` (+24h) e `GET_P2` (+27h).
- **Despachante Central & Fallback**: Rastreamento de banco em variável de memória e fallback para manipulação direta de porta I/O com suporte a sistemas sem EXTBIO.
- **Trampolins Cross-Bank Transparentes**: Geração automática de trampolins na Área Comum (Banco 0) com salvamento de contexto, troca para banco de destino, execução e restauração do banco anterior.
- **Script de Pipeline [sample/multibank/build.ps1](sample/multibank/build.ps1)**: Script PowerShell orquestrando montagem de múltiplos módulos (`main.asm`, `bank1.asm`, `bank2.asm`), linkagem com mapa `.map` e geração do executável `.com`.
- **Pacote Centralizado de Versão (`pkg/version`)**: Controle semântico unificado com exibição de versão e banner em `kaji80`, `musubi` e `mobdump`.

### Corrigido
- **Falha/Aborto após `[L]`**: Removida chamada a vetor estático `0xF3A4` (que causava salto para memória não inicializada e aborto ao DOS em cartuchos externos) e substituída pela detecção oficial via `EXTBIO`.
- **Sobrescrita do Banco 2 durante chamadas BDOS**: Resolvido problema em que rotinas de impressão BDOS (funções 02h e 09h) executadas dentro do Banco 2 faziam o kernel do MSX-DOS 2 restaurar a Página 2 para o Banco 1. Com a integração via `PUT_P2`, o kernel do DOS mantém o banco ativo correto.
- **Leitura em Porta Somente-Escrita (0xFE)**: Substituída a instrução `IN A, (0xFE)` (que retorna `0xFF` por ser porta write-only na maioria dos mappers) por rastreamento em software na Área Comum.

### Validado
- Validação completa com execução bem-sucedida em hardware real: MSX 2+ com 2048KB/4096KB de RAM mapeada e cartucho MSX-DOS 2 externo.

---

## [3.0.0] - 2026-09-03

### Adicionado
- **Linker Monobanco MUSUBI (`pkg/musubi` e `cmd/musubi`)**:
  - Resolução de símbolos globais `PUBLIC` e `EXTERN`.
  - Posicionamento de seções de código e dados a partir de `0x0100` (TPA do MSX-DOS).
  - Aplicação de relocações `ABS16` de 16 bits.
  - Emissão de binário executável `.COM` pronto para execução no MSX-DOS 2.
  - Geração de relatório de mapa de memória e tabela de símbolos (`.map`).
  - Suporte a ponto de entrada configurável (padrão: `Start`).
- **Programa de Exemplo [sample/hello.asm](sample/hello.asm)**:
  - Exemplo funcional chamando BDOS função 09h para imprimir texto.
  - Validado de ponta a ponta gerando `hello.mob` e `hello.com`.

---

## [2.0.0] - 2026-09-03

### Adicionado
- **Assembler Z80 KAJI80 (`pkg/kaji80` e `cmd/kaji80`)**:
  - Lexer modular com suporte a comentários Z80 (`;`), strings com caracteres de escape e múltiplos formatos de literais numéricos (hexadecimal, binário e decimal).
  - Montador de dois passos (*two-pass assembler*) emitindo arquivos de objeto `.MOB`.
  - Diretivas suportadas: `MODULE`, `BANK`, `PUBLIC`, `EXTERN`, `EQU`, `ORG`, `DB` / `DEFB` / `BYTE`, `DW` / `DEFW` / `WORD`, `DS` / `DEFS` / `BLKB`, `ENDMOD` / `END`.
  - Conjunto de instruções Z80:
    - Controle: `NOP`, `HALT`, `DI`, `EI`, `EXX`, `EX DE, HL`, `EX AF, AF'`.
    - Fluxo: `RET`, `RET cc`, `CALL nn`, `CALL cc, nn`, `JP nn`, `JP cc, nn`, `JP (HL|IX|IY)`, `JR e`, `JR cc, e`, `DJNZ e`.
    - Pilha: `PUSH` / `POP` (BC, DE, HL, AF, IX, IY).
    - I/O: `IN A, (n)`, `OUT (n), A`.
    - Aritmética/Lógica: `INC`, `DEC`, `ADD`, `ADC`, `SUB`, `SBC`, `AND`, `XOR`, `OR`, `CP`.
    - Movimentação: `LD` em todas as combinações fundamentais (8-bit, 16-bit, indiretos, registradores indexados IX/IY, SP).
  - Emissão automática de relocações `ABS16` e `REL8`.
- **Manual de Referência [HELP.md](HELP.md)**: Documentação completa da sintaxe do assembly, instruções e formato `.MOB`.

---

## [1.0.0] - 2026-09-03

### Adicionado
- **Pacote do Formato de Objeto Relocável .MOB (`pkg/mob`)**:
  - Estrutura binária em Little-Endian com cabeçalho `MOB1` (versão 1).
  - Segmentos tipados (`CODE`, `DATA`, `BSS`) com anotação de banco de memória (`0` = comum, `1..N` = paginável).
  - Tabela de símbolos com classes `PUBLIC` e `EXTERN`, categorias `PROC` e `DATA`.
  - Tabela de relocações (`ABS16`, `REL8`, `BANKNUM`).
  - Serializador (`writer.go`) com deduplicação inteligente no pool de strings.
  - Deserializador (`reader.go`) com checagem rigorosa de integridade e bounds.
  - Bateria de testes unitários com 100% de sucesso.
- **Utilitário de Inspeção MOBDUMP (`cmd/mobdump`)**:
  - Ferramenta de linha de comando para inspecionar cabeçalho, segmentos, símbolos e relocações de qualquer arquivo `.mob`.

---

## [0.0.0] - 2026-09-03

### Adicionado
- Primeira versão conceitual do projeto: nome, identidade das ferramentas (`KAJI80`, `WIRTH80`, `DIGNAC`, `MUSUBI`, `HAKO`, `OBI`) e especificação técnica inicial (`SPEC.md`).
- Definição do escopo da linguagem Pascal alvo: fiel ao Turbo Pascal 4, com suporte a units.
- Definição do modelo de memória: MSX2+, 256Kb ou mais com memory mapper, 4 páginas de 16Kb, janela comutável na página 2.
- Desenho da ABI própria: passagem via pilha, frame pointer em `IX`, limpeza pelo chamador.

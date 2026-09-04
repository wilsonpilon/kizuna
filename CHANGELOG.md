# Changelog

Todas as mudanças notáveis deste projeto são documentadas aqui.
Formato baseado em [Keep a Changelog](https://keepachangelog.com/).

### Política de Versionamento (`MAJOR.MINOR.COMPILAÇÃO`)
- **MAJOR**: Incrementado a cada encerramento de fase da toolchain (ex.: Fase 1 = MOB, Fase 2 = KAJI80, Fase 3 = MUSUBI monobanco, Fase 4 = Multi-banco & Memory Mapper).
- **MINOR**: Incrementado a cada feature ou subsistema novo adicionado.
- **COMPILAÇÃO (BUILD)**: Incrementado a cada compilação / build realizado no projeto.

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

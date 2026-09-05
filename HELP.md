# KIZUNA (絆) — Guia de Referência & Ajuda

Este documento detalha o funcionamento da toolchain **KIZUNA**, o formato de objetos **`.MOB`**, a sintaxe do assembler **`KAJI80`**, suas diretivas, instruções Z80 suportadas e parâmetros de compilação.

---

## 1. O Formato de Objeto `.MOB` (MSX Object)

O formato `.MOB` é o arquivo binário intermediário relocável gerado pelos compiladores do KIZUNA (`KAJI80`, `WIRTH80`, `DIGNAC`) e consumido pelo linker (`MUSUBI`).

### 1.1. Layout Binário em Disco

Todos os inteiros de 16 bits são codificados em **Little-Endian** (padrão Z80).

```
+-------------------------------------------------------------+
| Header (13 bytes)                                           |
|   Offset 0..3:  Magic "MOB1" (4 bytes)                      |
|   Offset 4:     Versão do formato (uint8, atual = 1)        |
|   Offset 5..6:  Nº de Segmentos (uint16)                    |
|   Offset 7..8:  Nº de Símbolos (uint16)                     |
|   Offset 9..10: Nº de Relocações (uint16)                   |
|   Offset 11..12:Offset inicial da Tabela de Strings (uint16)|
+-------------------------------------------------------------+
| Tabela de Segmentos                                         |
|   Cada segmento possui:                                     |
|     - Tipo (uint8): 1 = CODE, 2 = DATA, 3 = BSS             |
|     - Banco (uint8): 0 = Área comum, 1..N = Banco paginável |
|     - Tamanho (uint16): bytes em disco ou reserva de BSS    |
|     - Dados brutos: [Tamanho] bytes (ausente se BSS)        |
+-------------------------------------------------------------+
| Tabela de Símbolos (8 bytes por entrada)                    |
|   - Offset do Nome na String Table (uint16)                 |
|   - Classe (uint8): 1 = PUBLIC (exportado), 2 = EXTERN      |
|   - Tipo (uint8): 1 = PROC (código), 2 = DATA (dados)       |
|   - Índice do Segmento (uint16, se PUBLIC)                  |
|   - Offset dentro do Segmento (uint16, se PUBLIC)           |
+-------------------------------------------------------------+
| Tabela de Relocações (7 bytes por entrada)                  |
|   - Índice do Segmento onde aplicar (uint16)                |
|   - Offset dentro do Segmento onde aplicar (uint16)         |
|   - Índice do Símbolo alvo na Tabela de Símbolos (uint16)   |
|   - Tipo de Reloc (uint8):                                  |
|       1 = ABS16   (endereço absoluto de 16 bits)            |
|       2 = REL8    (deslocamento relativo de 8 bits)         |
|       3 = BANKNUM (número do banco de memória, 1 byte)      |
+-------------------------------------------------------------+
| Tabela de Strings                                           |
|   Pool de strings terminadas com byte nulo (\0).            |
|   Offset 0 sempre contém \0 para representar string vazia.  |
+-------------------------------------------------------------+
```

---

## 2. O Assembler `KAJI80`

O `KAJI80` (*kaji* = ferreiro/forjador) é o montador Z80 nativo do projeto Kizuna, escrito em Go. Ele monta arquivos `.asm` e emite objetos `.mob` compatíveis com paginação de memória.

### 2.1. Parâmetros de Linha de Comando

```bash
kaji80 [opções] <arquivo.asm>
```

| Opção | Descrição |
|---|---|
| `-o <saida.mob>` | Define o nome e caminho do arquivo `.mob` de saída. Por padrão, substitui a extensão `.asm` por `.mob`. |
| `-v` | Modo detalhado (*verbose*): exibe contagem de segmentos, símbolos, relocações e bancos. |
| `-h`, `--help` | Exibe a tela de ajuda completa no terminal. |

#### Exemplo de uso:
```bash
# Montagem padrão
kaji80 sample/hello.asm

# Montagem com saída explícita e modo detalhado
kaji80 -v -o build/hello.mob sample/hello.asm
```

---

## 3. Diretivas Suportadas pelo `KAJI80`

| Diretiva | Sintaxe | Descrição |
|---|---|---|
| `MODULE` | `MODULE <nome>` | Define o identificador do módulo. |
| `BANK` | `BANK <n>` | Atribui o segmento ao banco de memória: `0` para a área comum fixa (`4000h..7FFFh`), ou `1..N` para bancos trocáveis na janela de paginação (`8000h..BFFFh`). |
| `PUBLIC` | `PUBLIC sym1 [, sym2...]` | Exporta labels para outros módulos e para o linker. |
| `EXTERN` | `EXTERN sym1 [, sym2...]` | Declara símbolos externos (importados de outros módulos ou rotinas de BIOS/BDOS). |
| `EQU` | `<nome> EQU <valor>` | Define uma constante simbólica. Não gera relocation nem ocupa espaço em disco. |
| `ORG` | `ORG <endereço>` | Ajusta a origem/offset base do segmento atual. |
| `DB` / `DEFB` / `BYTE` | `DB item1, item2...` | Emite bytes ou strings literais de texto. |
| `DW` / `DEFW` / `WORD` | `DW val1, val2...` | Emite palavras de 16 bits em Little-Endian. Suporta labels que geram relocações `ABS16`. |
| `DS` / `DEFS` / `BLKB` | `DS <tamanho>` | Reserva espaço de `N` bytes preenchidos com zeros. |
| `ENDMOD` / `END` | `ENDMOD` | Finaliza a declaração do módulo (opcional). |

---

## 4. Sintaxe e Literais Aceitos

### Comentários
Iniciados pelo caractere `;` até o fim da linha:
```asm
; Isto é um comentário
ld a, 2 ; comentário inline
```

### Formatos Numéricos
- **Hexadecimal:** `0x100`, `100h`, `100H`, `$100`, `#100`
- **Binário:** `%10100111`, `0b10100111`, `10100111b`
- **Decimal:** `42`, `255`, `0`

### Strings
Delimitadas por aspas duplas `"..."` ou simples `'...'`, com suporte a caracteres de escape:
```asm
db "Hello\r\n", 0
db 'MSX', 0x0D, 0x0A, '$'
```

---

## 5. Conjunto de Instruções Z80 Suportadas

O `KAJI80` cobre todas as instruções essenciais para controle, chamadas, fluxo, aritmética e movimentação:

### Controle e Estado
- `NOP` (`0x00`)
- `HALT` (`0x76`)
- `DI` (`0xF3`), `EI` (`0xFB`)
- `EXX` (`0xD9`)
- `EX DE, HL` (`0xEB`), `EX AF, AF'` (`0x08`)

### Fluxo e Chamadas
- `RET` (`0xC9`) e `RET cc` (`NZ`, `Z`, `NC`, `C`, `PO`, `PE`, `P`, `M`)
- `CALL nn` / `CALL sym` (`0xCD nnL nnH`) e `CALL cc, nn` (gera relocation `ABS16`)
- `JP nn` / `JP sym` (`0xC3 nnL nnH`) e `JP cc, nn`
- `JP (HL)` (`0xE9`), `JP (IX)` (`0xDD 0xE9`), `JP (IY)` (`0xFD 0xE9`)
- `JR e` (`0x18 e`), `JR cc, e` (`NZ`, `Z`, `NC`, `C`) (gera relocation `REL8` se跨-label)
- `DJNZ e` (`0x10 e`)

### Pilha (Stack)
- `PUSH rr` e `POP rr` para `BC`, `DE`, `HL`, `AF`
- `PUSH IX` (`0xDD 0xE5`), `POP IX` (`0xDD 0xE1`)
- `PUSH IY` (`0xFD 0xE5`), `POP IY` (`0xFD 0xE1`)

### Entrada e Saída (I/O)
- `IN A, (n)` (`0xDB n`)
- `OUT (n), A` (`0xD3 n`)

### Aritmética e Lógica
- `INC r`, `DEC r` (8-bit: `A`, `B`, `C`, `D`, `E`, `H`, `L`)
- `INC rr`, `DEC rr` (16-bit: `BC`, `DE`, `HL`, `SP`, `IX`, `IY`)
- `ADD A, ...`, `ADC A, ...`, `SUB ...`, `SBC A, ...`, `AND ...`, `XOR ...`, `OR ...`, `CP ...`:
  - Operando registrador: `r` ou `(HL)`
  - Operando imediato: `n` ou constante `EQU`
- `ADD HL, rr` (`BC`, `DE`, `HL`, `SP`)
- `ADD IX, rr` (`BC`, `DE`, `IX`, `SP`)

### Movimentação de Dados (`LD`)
- `LD r, r'` (cópia entre registradores de 8 bits)
- `LD r, n` (carregamento imediato de 8 bits)
- `LD r, (HL)` e `LD (HL), r`
- `LD (HL), n`
- `LD A, (BC)` e `LD A, (DE)`
- `LD (BC), A` e `LD (DE), A`
- `LD A, (nn)` e `LD (nn), A`
- `LD HL, (nn)` e `LD (nn), HL`
- `LD rr, nn` (`BC`, `DE`, `HL`, `SP`) — gera relocation `ABS16` se `nn` for um label
- `LD IX, nn` (`0xDD 0x21`) e `LD IY, nn` (`0xFD 0x21`)
- `LD SP, HL` (`0xF9`)

---

## 6. Ferramenta de Inspeção `mobdump`

O KIZUNA inclui o utilitário `mobdump` para inspecionar visualmente o conteúdo de qualquer arquivo `.mob`:

```bash
go run ./cmd/mobdump sample/hello.mob
```

---

## 7. O Linker `MUSUBI`

O `MUSUBI` (*musubi* = atar, amarrar) é o linker do KIZUNA. Ele é responsável por unir múltiplos objetos `.mob`, posicionar os segmentos em memória a partir do endereço base (padrão `0x0100` no MSX-DOS 2), resolver referências cruzadas entre símbolos e aplicar as relocações (`ABS16`, `REL8` e `BANKNUM`).

### 7.1. Parâmetros de Linha de Comando

```bash
musubi [opções] <objeto.mob...> [biblioteca.hlib...]
```

| Opção | Descrição |
|---|---|
| `-o <saida.com>` | Nome do executável de saída (padrão: nome do primeiro arquivo com extensão `.com`). |
| `-m <mapa.map>` | Gera relatório de mapa de memória e tabela global de símbolos. |
| `-b <endereço>` | Endereço base de carregamento (padrão: `0x0100` para a TPA do MSX-DOS 2). |
| `-e <símbolo>` | Símbolo do ponto de entrada do programa (padrão: `Start`). |
| `-v` | Modo detalhado (*verbose*): exibe relatório na tela com endereço base, ponto de entrada e tamanho final. |
| `-h`, `--help` | Exibe a ajuda detalhada do linker. |

---

## 8. Ciclo Completo de Desenvolvimento (Exemplo Prático)

Para compilar um programa Assembly Z80 e gerar o binário `.COM` executável para MSX-DOS 2:

```bash
# 1. Montar o arquivo fonte gerando o objeto relocável .mob
go run ./cmd/kaji80 -v -o sample/hello.mob sample/hello.asm

# 2. (Opcional) Inspecionar os segmentos, símbolos e relocações do .mob
go run ./cmd/mobdump sample/hello.mob

# 3. Linkar o objeto gerando o executável .com e o relatório de mapa de memória .map
go run ./cmd/musubi -v -m sample/hello.map -o sample/hello.com sample/hello.mob
```

---

## 9. Suporte Multi-Banco & Trampolins de Bank Switching

O grande diferencial do KIZUNA é quebrar a barreira dos 64KB através da Memory Mapper do MSX2+:

- **Área Comum (Banco 0):** Reside em `0x0100..0x7FFF` (sempre presente na memória).
- **Janela Comutável (Bancos 1..N):** Mapeados na Página 2 (`0x8000..0xBFFF`), cada banco podendo ter até 16KB.
- **Trampolins Automáticos:** Quando um módulo na Área Comum ou no Banco A chama uma rotina no Banco B, o `MUSUBI` intercepta a chamada e gera automaticamente um trampolim na Área Comum que:
  1. Salva o banco atual da Página 2.
  2. Escreve o banco de destino na porta I/O da Memory Mapper (`0xFE`).
  3. Executa a função via `CALL`.
  4. Restaura o banco anterior na Página 2 preservando os registradores e flags de retorno.
  5. Retorna ao chamador original.
- **Otimização Intra-banco:** Se chamador e chamado estiverem no mesmo banco, o linker emite `CALL` direto, sem overhead.
- **Bootstrap Loader:** Programas multi-banco geram um único arquivo `.COM` autocontido que copia os payloads de cada banco para as páginas estendidas de RAM na inicialização.

### Exemplo de Compilação Multi-Banco:
```bash
# Montar cada módulo indicando seu banco no código-fonte (BANK 0, BANK 1, BANK 2...)
kaji80 sample/multibank/main.asm  -o sample/multibank/main.mob
kaji80 sample/multibank/bank1.asm -o sample/multibank/bank1.mob
kaji80 sample/multibank/bank2.asm -o sample/multibank/bank2.mob

# Linkar os 3 bancos num único .COM
musubi -v -m sample/multibank/multibank.map -o sample/multibank/multibank.com \
  sample/multibank/main.mob sample/multibank/bank1.mob sample/multibank/bank2.mob
```

---

## 10. O Bibliotecário de Objetos `HAKO` e o Formato `.HLIB`

O **`HAKO`** (*hako* = caixa/baú) gerencia bibliotecas de arquivos objeto `.MOB` empacotadas no formato **`.HLIB`**.

### 10.1. O Formato `.HLIB`
- **Cabeçalho (14 bytes):** Magic `"HLIB"`, versão do formato (`1`), contagem de módulos, offset e contagem do dicionário de símbolos públicos.
- **Tabela de Módulos:** Nome do módulo, offset e tamanho dos dados brutos do `.MOB`.
- **Dicionário Global de Símbolos:** Mapeia todos os símbolos `PUBLIC` de todos os módulos para seus respectivos módulos de origem. Rejeita símbolos públicos duplicados na criação.
- **Smart-Linking (Dead-Code Elimination):** Ao linkar com `musubi main.mob lib.hlib`, o linker consulta o dicionário da biblioteca e puxa **apenas os módulos requisitados** direta ou transitivamente pelo programa. Módulos não utilizados são completamente descartados, gerando binários enxutos.

### 10.2. Comandos do `HAKO`

```bash
# Criar ou atualizar biblioteca a partir de múltiplos arquivos .mob
hako -c math.hlib math_add.mob math_sub.mob math_trig.mob

# Listar módulos e símbolos contidos na biblioteca
hako -t math.hlib

# Extrair módulo específico de uma biblioteca
hako -x math.hlib math_add.mob
```

---

## 11. A Biblioteca Padrão `MSXLIB` (`msxlib.hlib`)

O KIZUNA inclui uma biblioteca de rotinas padrão prontas para uso em `lib/msxlib.hlib`:

| Módulo | Símbolos Exportados | Descrição |
|---|---|---|
| **`bdos`** | `BDOS_Call`, `BDOS_PrintChar`, `BDOS_PrintString`, `BDOS_ReadChar`, `BDOS_Exit` | Chamadas diretas ao kernel MSX-DOS (BDOS 0x0005). |
| **`bios`** | `BIOS_Call`, `BIOS_CHPUT`, `BIOS_CHGET`, `BIOS_CLS`, `BIOS_POSIT`, `BIOS_BEEP`, `BIOS_INIT32` | Chamadas inter-slot seguras à Main-ROM BIOS via `CALSLT (0x0024)` preservando o estado do MSX-DOS. |
| **`vdp`** | `VDP_WriteReg`, `VDP_SetWriteAddr`, `VDP_SetReadAddr`, `VDP_FillVRAM`, `VDP_WriteVRAM`, `VDP_ReadVRAM`, `VDP_SetColor` | Manipulação das portas I/O do processador de vídeo TMS9918 / V9938 / V9958 e acesso direto à VRAM. |
| **`psg`** | `PSG_Write`, `PSG_Read`, `PSG_MuteAll`, `PSG_PlayTone` | Controle dos registradores de som do AY-3-8910 / YM2149, reprodução de notas e silenciamento. |
| **`string`** | `StrLen`, `StrCopy`, `StrToUpper`, `PrintHex8`, `PrintHex16`, `PrintDec16` | Manipulação de textos terminados em zero (`\0`) e conversão de números para hexadecimal e decimal formatado. |
| **`math`** | `Mul16`, `Div16` | Multiplicação e divisão inteira não sinalizada de 16 bits rápida por deslocamento e soma. |

### Exemplo de Uso com Smart-Linking:
```bash
# Linkando seu programa com a biblioteca MSXLIB
# O MUSUBI inclui no .COM apenas os módulos utilizados pelo seu código
musubi -v -m app.map -o app.com main.mob msxlib.hlib
```

---

## 12. O Compilador Pascal `WIRTH80`

O **`WIRTH80`** (em homenagem a Niklaus Wirth) é o compilador da linguagem Pascal no estilo Turbo Pascal 4 da toolchain KIZUNA, projetado para produzir módulos objeto `.MOB` nativos para MSX2+ / MSX-DOS 2.

### 12.1. Parâmetros de Linha de Comando

```bash
wirth80 [opções] <arquivo.pas>
```

| Opção | Descrição |
|---|---|
| `-o <caminho>` | Define o arquivo de saída `.mob` (padrão: mesmo nome com extensão `.mob`). |
| `-S` | Emite o código intermediário Assembly Z80 formatado (`.asm`) em vez de gerar o `.mob`. Excelente para inspeção, aprendizado e depuração. |
| `-v` | Modo detalhado (*verbose*): exibe a contagem de segmentos, símbolos, relocações e o assembly gerado internamente. |
| `--version` | Exibe a versão atual do compilador. |
| `-h`, `--help` | Exibe a ajuda detalhada de uso. |

### 12.2. Recursos e Sintaxe Suportada (v4.4.0)

- **Estrutura de Programa**: `program <Nome>;` seguido de declarações e bloco `begin ... end.`.
- **Declarações de Variáveis**: Bloco `var` suportando `Integer` (16 bits sinalizado em Little-Endian) e `Char` / `Boolean` (8 bits).
- **Controle de Fluxo e Comandos**:
  - Atribuição: `identificador := expressao;`
  - Impressão: `Write(...)` e `WriteLn(...)` suportando literais de string e expressões inteiras, além de `WriteLn;` (quebra de linha).
  - Condicionais: `if cond then comando [else comando]`
  - Laços: `while cond do comando`
  - Blocos aninhados: `begin ... end;`
- **Expressões e Operadores**:
  - Aritmética de 16 bits: Adição (`+`), subtração (`-`), multiplicação (`*` via chamada a `Mul16` da MSXLIB) e divisão (`div` via `Div16`).
  - Comparações relacionais: Igual (`=`), diferente (`<>`), menor (`<`), menor ou igual (`<=`), maior (`>`) e maior ou igual (`>=`).
  - Constantes: Inteiros decimais (`123`), hexadecimais com prefixo `$` (`$FF`), e literais de string (`'texto'`).
  - Comentários: Suporte a `{ bloco }`, `(* bloco *)` e `// linha`.

### 12.3. Ciclo Completo de Compilação Pascal:

```bash
# 1. Compilar o código-fonte Pascal gerando o módulo objeto .mob
wirth80 -v sample/pascal/hello.pas -o sample/pascal/hello.mob

# 2. Linkar com a biblioteca padrão msxlib.hlib para gerar o executável MSX-DOS
musubi -v -m sample/pascal/hello.map -o sample/pascal/hello.com \
  sample/pascal/hello.mob lib/msxlib.hlib

# 3. Executar o binário gerado no MSX2+ ou emulador!
```

---

## 13. O Compilador MSX-BASIC Dignified `DIGNAC`

O **`DIGNAC`** (*Dignified-ac*) é o compilador da linguagem **MSX-BASIC Dignified** da toolchain KIZUNA. Ele traduz código BASIC estruturado diretamente para módulos relocáveis `.MOB` compatíveis com o linker `MUSUBI`, com suporte a bancos pagináveis e convenção de chamadas em pilha (ABI).

### 13.1. Parâmetros de Linha de Comando

```bash
dignac [opções] <arquivo.bas>
```

| Opção | Descrição |
|---|---|
| `-o <caminho>` | Define o arquivo de saída `.mob` (padrão: mesmo nome com extensão `.mob`). |
| `-S` | Emite o código intermediário Assembly Z80 formatado (`.asm`) em vez de gerar o `.mob`. |
| `-v` | Modo detalhado (*verbose*): exibe a contagem de segmentos, símbolos, relocações e o assembly gerado internamente. |
| `--version` | Exibe a versão atual do compilador. |
| `-h`, `--help` | Exibe a ajuda detalhada de uso. |

### 13.2. Recursos e Sintaxe Suportada (v4.5.0)

- **Módulos & Paginação**:
  - `MODULE <Nome>` e `END MODULE`: Define a unidade de compilação.
  - `BANK <n>`: Define o banco de alocação (0 para área comum fixa, 1..N para bancos pagináveis na página 2).
  - `PUBLIC sym1 [, sym2...]`: Exporta símbolos e sub-rotinas para o linker.
  - `EXTERN sym1 [, sym2...]`: Declara símbolos importados de outros módulos.
- **Sub-rotinas & ABI**:
  - `PROCEDURE Nome([par1%, par2$ ...])` ... `END PROCEDURE` (ou `SUB ... END SUB`).
  - `FUNCTION Nome([params]) [AS Tipo]` ... `END FUNCTION`.
  - Passagem de parâmetros em pilha (esquerda para a direita) e quadro de ativação com frame pointer `IX`.
  - Se houver `PROCEDURE Main`, o compilador gera automaticamente o ponto de entrada `Start` para criação direta de executáveis `.COM`.
- **Variáveis**:
  - `LOCAL var1%, var2%`: Variáveis locais alocadas no quadro de pilha (`IX - offset`).
  - `DIM var1%, var2%`: Variáveis globais com escopo de módulo.
  - Tipos primitivos: Inteiro de 16 bits (sufixo `%`), strings (sufixo `$`) e booleanos/bytes (sufixo `!`).
- **Controle de Fluxo**:
  - `FOR var% = inicio TO fim [STEP passo] ... NEXT [var%]`
  - `IF condicao THEN ... [ELSE ...] [END IF]` (linha única ou multi-linha).
  - `WHILE condicao ... WEND` e `DO [WHILE condicao] ... LOOP`.
- **Primitivas Gráficas & MSX**:
  - `LINE (x1, y1)-(x2, y2)[, color][, B | BF]`: Traçado de linhas, retângulos vazados e caixas preenchidas (com preenchimento acelerado de tela cheia).
  - `PSET (x, y)[, color]`: Plotagem de pixel de alta precisão em SCREEN 2.
  - `PRINT [expressões...]`: Exibição de texto e números via BDOS e MSXLIB.
  - `CLS`, `BEEP`, `SCREEN modo`, `COLOR fg, bg, bd`.
- **Expressões & Aritmética**:
  - Aritmética de 16 bits: `+`, `-`, `*` (via `Mul16`), `/` e operador `MOD` (via `Div16`).
  - Lógica: `AND`, `OR`, `XOR`, `NOT`.
  - Operadores relacionais: `=`, `<>`, `<`, `<=`, `>`, `>=`.

### 13.3. Ciclo Completo de Compilação BASIC:

```bash
# 1. Compilar o código-fonte BASIC gerando o módulo objeto .mob
dignac -v sample/basic/hello.bas -o sample/basic/hello.mob

# 2. Linkar com a biblioteca padrão msxlib.hlib para gerar o executável .COM
musubi -v -m sample/basic/hello.map -o sample/basic/hello.com \
  sample/basic/hello.mob lib/msxlib.hlib
```






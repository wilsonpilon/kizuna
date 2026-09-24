# KIZUNA — Manual das Ferramentas (Formatos, Linker, Bibliotecário, Build)

> Este manual cobre tudo que não é específico de uma linguagem de entrada:
> o formato de objeto `.MOB`, o linker `MUSUBI` (e o formato `.MAP`), o
> bank switching multi-banco, o bibliotecário `HAKO` e o formato `.HLIB`,
> a biblioteca padrão `MSXLIB`, o inspetor `MOBDUMP`, o orquestrador de
> build `OBI` (o "make" do KIZUNA), e um exemplo prático combinando as três
> linguagens de entrada num mesmo projeto. Para a sintaxe de cada
> linguagem, veja `docs/manual-assembly.md`, `docs/manual-basic-dignified.md`
> e `docs/manual-pascal.md`.

## 1. Visão geral do pipeline

```
fonte.asm  ──KAJI80──►   fonte.mob  ─┐
fonte.bas  ──DIGNAC──►   fonte.mob  ─┼──MUSUBI (+ .hlib opcionais)──►  programa.com
fonte.pas  ──WIRTH80──►  fonte.mob  ─┘

     lib/src/**/*.asm ──KAJI80──► *.mob ──HAKO──► msxlib.hlib
```

Os três compiladores de frontend (`KAJI80`, `DIGNAC`, `WIRTH80`) emitem o
mesmo formato de objeto relocável `.MOB` (§2). `MUSUBI` (§3) resolve
símbolos entre módulos e gera o `.COM` final, distribuindo código por
bancos de memória quando necessário (§4). `HAKO` (§5) empacota objetos
`.mob` reutilizáveis num arquivo `.HLIB`, como a biblioteca padrão
`MSXLIB` (§6). `OBI` (§7) automatiza tudo isso a partir de uma receita
declarativa — o "make" do projeto.

## 2. O formato de objeto `.MOB`

Todos os inteiros de 16 bits são little-endian (padrão Z80).

```
+-------------------------------------------------------------+
| Header (13 bytes)                                           |
|   0..3   Magic "MOB1"                                       |
|   4      Versão do formato (uint8, atual = 1)                |
|   5..6   Nº de Segmentos (uint16)                            |
|   7..8   Nº de Símbolos (uint16)                             |
|   9..10  Nº de Relocações (uint16)                           |
|   11..12 Offset inicial da Tabela de Strings (uint16)        |
+-------------------------------------------------------------+
| Tabela de Segmentos                                          |
|   Tipo (uint8): 1=CODE, 2=DATA, 3=BSS                        |
|   Banco (uint8): 0=área comum, 1..N=banco paginável           |
|   Tamanho (uint16)                                            |
|   Dados brutos: [Tamanho] bytes (ausente se BSS)              |
+-------------------------------------------------------------+
| Tabela de Símbolos (8 bytes por entrada)                     |
|   Offset do Nome na String Table (uint16)                    |
|   Classe (uint8): 1=PUBLIC, 2=EXTERN                          |
|   Kind (uint8): 1=PROC, 2=DATA                                |
|   Índice do Segmento (uint16, se PUBLIC)                      |
|   Offset dentro do Segmento (uint16, se PUBLIC)               |
+-------------------------------------------------------------+
| Tabela de Relocações (7 bytes por entrada)                   |
|   Índice do Segmento onde aplicar (uint16)                   |
|   Offset dentro do Segmento onde aplicar (uint16)             |
|   Índice do Símbolo alvo na Tabela de Símbolos (uint16)       |
|   Tipo (uint8): 1=ABS16, 2=REL8, 3=BANKNUM                    |
+-------------------------------------------------------------+
| Tabela de Strings                                             |
|   Pool de strings terminadas em \0; offset 0 = string vazia   |
+-------------------------------------------------------------+
```

Regras:
- `REL8` (usado por `JR`/`DJNZ`) só é permitido dentro do mesmo
  segmento/banco — um `JR` cruzando bancos não faz sentido fisicamente
  (a janela paginável muda de conteúdo).
- Chamadas cross-bank nunca usam salto relativo; sempre viram `CALL`/`JP`
  absoluto por trampolim (§4).
- `Kind` (`PROC`/`DATA`) é o que o linker usa pra decidir se um símbolo
  chamado de outro banco precisa de trampolim (`PROC`, código) ou é só um
  endereço de dado (`DATA`, sem trampolim — dados não "saltam").

Inspecione qualquer `.mob` com `mobdump <arquivo.mob>` (§8).

## 3. O linker `MUSUBI`

```bash
musubi [opções] <objeto.mob...> [biblioteca.hlib...]
```

| Opção            | Descrição                                                                  |
| ----------------- | ------------------------------------------------------------------------------ |
| `-o <saida.com>`  | Nome do executável de saída (padrão: nome do primeiro arquivo, extensão `.com`). |
| `-m <mapa.map>`   | Gera o relatório de mapa de memória (§3.1).                                    |
| `-b <endereço>`   | Endereço base de carregamento (padrão `0x0100`, TPA do MSX-DOS 2).             |
| `-e <símbolo>`    | Ponto de entrada (padrão `Start`).                                             |
| `-v`               | Modo detalhado.                                                                |
| `-h`, `--help`     | Ajuda completa.                                                                |

```bash
musubi -v -m app.map -o app.com main.mob lib/msxlib.hlib
```

`MUSUBI` carrega todos os `.mob`/`.hlib` passados, resolve cada `EXTERN`
contra um `PUBLIC` em outro módulo (via **Smart-Linking** quando a origem é
uma `.hlib` — ver §5), posiciona os segmentos em memória a partir do
endereço base, aplica as relocações (`ABS16`/`REL8`/`BANKNUM`) e — se
houver módulos em bancos diferentes — gera os trampolins de bank switching
automaticamente (§4).

### 3.1. O formato `.MAP`

Texto simples, gerado por `-m`, com quatro seções:

```
Modo de Memória: Monobanco | Multi-Banco (Memory Mapper)
Bancos Utilizados: [...]
Endereço Base / Ponto de Entrada / Tamanho do .COM

--- SEGMENTOS ALOCADOS ---
[ 0] 0x0100 - 0x0142 | Banco:  0 | Tipo: CODE | Mod:  0 | Tam:    66 bytes
...

--- TRAMPOLINS DE BANK SWITCHING (ÁREA COMUM -> PÁGINA 2) ---
0xNNNN -> Alvo: 'Simbolo' no Banco  2 (0xNNNN) | Tamanho: N bytes
...

--- TABELA GLOBAL DE SÍMBOLOS ---
0xNNNN [Banco  0] PROC Main
...
```

Use o `.map` pra depurar "por que meu `.com` ficou desse tamanho" ou "que
endereço esse símbolo caiu" sem precisar abrir o binário num debugger.

## 4. Bank switching multi-banco

O grande diferencial do KIZUNA sobre um `.COM` de 64KB comum é usar a
Memory Mapper do MSX2+ para quebrar essa barreira.

### 4.1. Modelo de memória

```
0000h-3FFFh  Página 0  Fixa — reservada para MSX-DOS 2 / BIOS
4000h-7FFFh  Página 1  Fixa — runtime comum + código sempre presente (Banco 0)
8000h-BFFFh  Página 2  Janela comutável — bancos pagináveis 1..N entram aqui
C000h-FFFFh  Página 3  Fixa — pilha, heap, buffers
```

Um módulo grande (unit Pascal, módulo BASIC Dignified, bloco de Assembly)
pode ser alocado inteiro num banco paginável mapeado na Página 2. O
`MUSUBI` decide a alocação de cada `BANK <n>` declarado no fonte e resolve
chamadas entre bancos automaticamente.

### 4.2. Trampolins automáticos

Gerado para todo símbolo `PUBLIC` do tipo `PROC` cujo chamador está num
banco diferente do chamado. Fica alocado na área comum (Página 1):

```asm
CALL_simbolo:
    push af
    in   a,(MAPPER_PAGE2)   ; salva banco atual da página 2
    push af
    ld   a, N               ; banco onde o símbolo mora
    out  (MAPPER_PAGE2), a
    call real_simbolo       ; endereço dentro de 8000h-BFFFh
    pop  af
    out  (MAPPER_PAGE2), a  ; restaura banco anterior
    pop  af
    ret
```

Se chamador e chamado estão no **mesmo** banco, `MUSUBI` emite `CALL`
direto — sem overhead nenhum de trampolim (otimização automática, não
precisa pedir). A mesma otimização vale sempre que o **alvo** é o banco
comum (banco 0), não importa em qual banco o chamador está — a área
comum está sempre presente na memória, independente do que estiver
mapeado na Página 2 no momento, então uma rotina num banco paginável
chamando de volta uma rotina `MSXLIB` no banco comum (ex: um
`PROCEDURE` `DIGNAC` no banco 2 chamando `VDP_PSet`) também vira `CALL`
direto. **Bug real corrigido em 2026-09-23**: antes disso, esse caminho
específico gerava um trampolim mesmo assim, que lia uma entrada da
tabela de bancos nunca inicializada pelo bootstrap (só preenche
1..N, nunca a entrada 0) — `sample/obi/main.com` carregava e voltava
limpo pro MSX-DOS, mas o desenho em SCREEN 2 nunca aparecia (tela
preta). Corrigido e confirmado em hardware — ver `CHANGELOG.md`.

### 4.3. Bootstrap multi-banco

Um `.COM` que usa mais de um banco é ainda um único arquivo autocontido: o
bootstrap gerado copia o payload de cada banco pra sua página de RAM
estendida na inicialização, usando mapeamento por identidade (segmento
físico da Memory Mapper = número lógico do banco no `MUSUBI`) — a
alocação dinâmica via `ALL_SEG` do EXTBIO foi tentada em uma versão
anterior, mas nunca funcionou de verdade em hardware real e foi
revertida; o mecanismo de detecção de EXTBIO que resta só decide *como*
trocar de página (via rotinas oficiais quando disponíveis, com
fallback pra porta de I/O direta), não *qual* segmento físico usar.

### 4.4. Compilando um programa multi-banco

```bash
# Cada módulo com seu BANK declarado no próprio fonte (BANK 0, BANK 1, BANK 2...)
kaji80 sample/multibank/main.asm  -o sample/multibank/main.mob
kaji80 sample/multibank/bank1.asm -o sample/multibank/bank1.mob
kaji80 sample/multibank/bank2.asm -o sample/multibank/bank2.mob

musubi -v -m sample/multibank/multibank.map -o sample/multibank/multibank.com \
  sample/multibank/main.mob sample/multibank/bank1.mob sample/multibank/bank2.mob
```

O mesmo vale misturando compiladores — um `.mob` gerado por `dignac` num
`BANK 2` linka junto com um `.mob` gerado por `kaji80` num `BANK 0` sem
diferença nenhuma pro `MUSUBI` (ele só enxerga símbolos e bancos, nunca a
linguagem de origem). Ver §10 para um exemplo real disso.

## 5. O bibliotecário `HAKO` e o formato `.HLIB`

`HAKO` (箱, "caixa") empacota vários `.mob` reutilizáveis num único arquivo
`.HLIB`.

### 5.1. Formato `.HLIB`

- **Cabeçalho (14 bytes)**: magic `"HLIB"`, versão (`1`), contagem de
  módulos, offset e contagem do dicionário de símbolos públicos.
- **Tabela de Módulos**: nome, offset e tamanho dos dados brutos de cada
  `.mob` embutido.
- **Dicionário Global de Símbolos**: mapeia cada símbolo `PUBLIC` de cada
  módulo pro módulo de origem. Rejeita símbolos públicos duplicados na
  criação.
- **Smart-Linking (eliminação de código morto)**: ao linkar
  `musubi main.mob lib.hlib`, o `MUSUBI` consulta o dicionário e puxa
  **só os módulos requisitados** direta ou transitivamente pelo programa
  — o resto fica de fora do `.com` final.

### 5.2. Comandos

```bash
# Criar/atualizar biblioteca a partir de vários .mob
hako -c math.hlib math_add.mob math_sub.mob math_trig.mob

# Listar módulos e símbolos
hako -t math.hlib

# Extrair um módulo específico (ou todos, se omitido)
hako -x math.hlib math_add.mob
```

## 6. A biblioteca padrão `MSXLIB` (`lib/msxlib.hlib`)

Construída a partir de `lib/src/**/*.asm` via `lib/build.ps1`. **Cada
arquivo `.asm` é um módulo do `.hlib`** — uma rotina (ou uma família de
rotinas inseparáveis) por arquivo, organizados em subdiretórios por área
(`lib/src/vdp/`, `lib/src/psg/`, `lib/src/bdos/`, ...), no mesmo espírito da
organização por tópico do MSXgl. Como o `MUSUBI` só traz para o `.com` os
módulos cujos símbolos o programa realmente usa (§6.1), módulos pequenos
significam executáveis menores: `sample/macroasm/predefined.com` (só usa
`CALLBIOS`) caiu de 1774 para 382 bytes com essa divisão, porque
`BIOS_Call` deixou de arrastar a biblioteca de VDP inteira.

A tabela abaixo agrupa por **área** (subdiretório); os nomes públicos das
rotinas não mudaram em relação à versão monolítica, então nenhum programa
existente precisa ser alterado:

| Módulo | Símbolos | Descrição |
| ------- | --------- | ----------- |
| **`bdos`** | `BDOS_Call`, `BDOS_PrintChar`, `BDOS_PrintString`, `BDOS_ReadChar`, `BDOS_Exit`, `BDOS_FileOpen`, `BDOS_FileCreate`, `BDOS_FileClose`, `BDOS_FileRead`, `BDOS_FileWrite`, `BDOS_FileSeek` | Kernel MSX-DOS (BDOS 0x0005), I/O de arquivo por handle (MSX-DOS 2, funções 43h-4Ah). |
| **`bios`** | `BIOS_Call`, `BIOS_CHPUT`, `BIOS_CHGET`, `BIOS_CLS`, `BIOS_POSIT`, `BIOS_BEEP`, `BIOS_INIT32`, `BIOS_CHGMOD` | Chamadas inter-slot seguras à Main-ROM BIOS via `CALSLT`, preservando o estado do MSX-DOS. |
| **`vdp`** | `VDP_WriteReg`, `VDP_WriteReg_Raw`, `VDP_SetWriteAddr`, `VDP_SetReadAddr`, `VDP_FillVRAM`, `VDP_WriteVRAM`, `VDP_ReadVRAM`, `VDP_CopyVRAM`, `VDP_SetColor`, `VDP_SetScreen`, `VDP_InitScreen0/1/2`, `VDP_InitScreen2_Tables`, `VDP_PSet`, `VDP_PSet_Raw`, `VDP_PSet_HW`, `VDP_CommandWait_Raw`, `VDP_Line`, `VDP_BoxFill`, `VDP_SpriteDefine`, `VDP_SpriteSet`, `VDP_SpriteHide`, `VDP_SpriteHideAll`, `VDP_SpriteSetSize`, e da Fase 2a: `VDP_SetReg/GetReg/UpdateReg` (cópia sombra), `VDP_ReadStatus`, `VDP_WaitVBlank/WaitFrames`, `VDP_DisplayOn/Off`, `VDP_SetMode`, `VDP_VramSetWrite/SetRead/Put/Get/WriteStream/ReadStream/FillStream` (17 bits), `VDP_VPoke/VPeek`, `VDP_ClearVRAM`, `VDP_SetPaletteEntry/Block`, `VDP_SetDefaultPalette/SetMSX1Palette`, e da Fase 2b: `VDP_Set/Get{Name,Pattern,Color,SpriteAttr,SpritePattern}Table`, `VDP_SpriteSetPos/SetAll/SetColor/SetLineColors/DisableFrom/PatternLoad` (modos 1 e 2), `VDP_BlinkFill/Line/Cell`, `VDP_SetVerticalOffset`, `VDP_GetVersion`, e da Fase 2c (motor de comandos, SCREEN 5-8): `VDP_CmdRun/Wait/Busy/Stop`, `VDP_HwPlot/HwPoint`, `VDP_HwFillRect/FillRectFast/BoxFill/Box/Line`, `VDP_HwCopyRect/MoveRect/CopyLines`, `VDP_HwSearch`, `VDP_HwLoadRect/LoadFast/ReadRect` e mais | V9938/TMS9918: VRAM, paleta, SCREEN 0/1/2, linhas/caixas, sprites (Sprite Mode 1, 8x8/16x16). |
| **`psg`** | `PSG_Write`, `PSG_Read`, `PSG_MuteAll`, `PSG_PlayTone`, `PSG_PlayNoteIndexed`, `PSG_PlaySequence` | AY-3-8910/YM2149: registradores, notas por nome/oitava, sequências de melodia. |
| **`string`** | `StrLen`, `StrCopy`, `StrToUpper`, `PrintHex8`, `PrintHex16`, `PrintDec16`, `PrintDec16ToBuffer`, `StrCopyLen`, `BDOS_PrintLenStr` | Texto terminado em `\0` (rotinas clássicas) e no formato "short string" 1-byte-tamanho+dados (`StrCopyLen`/`BDOS_PrintLenStr`, usadas pelas STRING de verdade do DIGNAC — ver `docs/manual-basic-dignified.md` §3); conversão pra hex/decimal. | **As rotinas novas de texto estão nas áreas `char`, `cstr`, `str`, `num` e `console` (abaixo); estes nomes continuam como estão.**
| **`char`** | `CHAR_IsDigit/Alpha/AlNum/Upper/Lower/HexDigit/Space/Control/Ascii/Print/Graph/Punct`, `CHAR_ToUpper`, `CHAR_ToLower`, `CHAR_DigitValue`, `CHAR_HexChar` | Classificação e conversão de caracteres (ASCII de 7 bits). Predicados devolvem A = 1/0 e o flag Z. Testados nos 256 valores. |
| **`cstr`** | `CSTR_Len`, `Copy`, `CopyN`, `Cat`, `CatN`, `Compare`, `CompareN`, `CompareNoCase`, `FindChar`, `FindLastChar`, `FindStr`, `ToUpper`, `ToLower`, `Reverse`, `TrimLeft`, `TrimRight`, `ReplaceChar` | Strings terminadas em zero, como as de C e do Fusion-C. |
| **`str`** | `STR_Len`, `Copy`, `Cat`, `Compare`, `Left`, `Right`, `Mid`, `InStr`, `Chr`, `Repeat`, `ToUpper`, `ToLower`, `FromC`, `ToC`, `FromU16`, `FromI16`, `Hex8`, `Hex16`, `Val`, `HexVal` | Strings no formato **tamanho + dados** (1 byte de tamanho, até 255 caracteres, sem terminador) — o formato das `STRING` do DIGNAC e do WIRTH80. `Left`/`Right`/`Mid`/`InStr`/`Chr`/`Repeat`/`Val` são o `LEFT$`/`RIGHT$`/`MID$`/`INSTR`/`CHR$`/`STRING$`/`VAL` do MS-BASIC. |
| **`num`** | `NUM_U16ToCStr`, `I16ToCStr`, `U16ToDecW`, `U8ToHex`, `U16ToHex`, `U8ToBin`, `U16ToBin`, `CStrToU16`, `CStrToI16`, `CStrToHex16` | Conversões número ⇄ texto. Os leitores pulam espaços, aceitam sinal e prefixos (`0x`, `&H`, `$`) e avisam erro/estouro pelo carry. |
| **`console`** | `CON_PrintCStr`, `PrintLine`, `NewLine`, `PrintI16`, `Cls`, `Locate`, `ReadLine`, `ReadKey`, `KeyPressed` | Console de texto do MSX-DOS: saída de strings e números, posicionamento do cursor, leitura de linha e de tecla. (A pasta é `console`, não `con`: `CON` é nome de dispositivo reservado no Windows e não pode ser nome de arquivo nem de pasta.) |
| **`mem`** | `MEM_Copy`, `MEM_CopyFast`, `MEM_CopyRev`, `MEM_CopyWords`, `MEM_CopyFastWords`, `MEM_Fill`, `MEM_Fill16`, `MEM_Zero`, `MEM_Swap`, `MEM_Compare`, `MEM_Find`, `MEM_GetSP`, `MEM_TPATop`, `MEM_HeapInit`, `MEM_HeapInitToStack`, `MEM_Alloc`, `MEM_Free`, `MEM_BlockSize`, `MEM_HeapCompact`, `MEM_HeapSize`, `MEM_HeapFree`, `MEM_HeapLargest` | Blocos de memória (a cópia trata sobreposição, como `memmove`) e um heap dinâmico: lista de blocos com cabeçalho de 2 bytes, "primeiro que couber", fusão de livres vizinhos, sentinela no fim; `MEM_Free` recusa ponteiro inválido e liberação dupla. Testado no simulador contra um modelo em Go comparado **byte a byte** depois de cada operação (`pkg/msxlib/mem_test.go`). Descritores em `lib/api/mem.api`. As rotinas de heap não são reentrantes (não chame do tratador de interrupção). |
| **`math`** | `Mul16`, `Div16`, `MATH_Mod16`, `MATH_DivS16`, `MATH_ModS16`, `MATH_Mul8`, `MATH_MulU16x16`, `MATH_MulS16x16`, `MATH_DivU32By16`, `MATH_Neg8/16/32`, `MATH_Abs8/16/32`, `MATH_Sign16`, `MATH_Cmp16U/S`, `MATH_Min16U/S`, `MATH_Max16U/S`, `MATH_Clamp16S`, `MATH_Shl/Shr/Sar` (8 e 16 bits), `MATH_Sqrt16`, `MATH_DivMod10`, `MATH_Mod10`, `MATH_DivS10`, `MATH_Flip8/16`, `MATH_Swap16`, `MATH_FixMul88`, `MATH_FixDiv88`, `MATH_RandSeed`, `MATH_Rand8/16`, `MATH_RandRange8/16`, `MATH_RandBetween8/16` | Inteiros de 8/16/32 bits com e sem sinal, comparação, faixa, deslocamentos, raiz, decimal, ponto fixo 8.8 e gerador pseudoaleatório (xorshift de 16 bits, período 65535). Cada rotina documenta o que preserva; todas são testadas contra referência em Go no simulador Z80 (`pkg/msxlib/math_test.go`). Descritores em `lib/api/math.api`. |

### 6.1. Uso com Smart-Linking

```bash
musubi -v -m app.map -o app.com main.mob lib/msxlib.hlib
```

Só os módulos da `MSXLIB` de fato referenciados (direta ou
transitivamente) entram no `.com` final.

### 6.1.1. Layout de `lib/` e como acrescentar uma rotina

```
lib/
  build.ps1          monta tudo e empacota em msxlib.hlib
  inc/               constantes compartilhadas (INCLUDE): vdp.inc, psg.inc, bdos.inc, bios.inc
  src/<area>/*.asm   um módulo por arquivo (MODULE <area>_<nome>)
  obj/               .mob intermediários (gerado, ignorado pelo git)
```

Para acrescentar uma rotina: crie `lib/src/<area>/<nome>.asm` com `MODULE`,
`BANK 0`, `PUBLIC`/`EXTERN` e `INCLUDE "../../inc/<area>.inc"` se precisar
das constantes de porta/endereço, rode `lib/build.ps1` e pronto — o script
descobre os arquivos sozinho (busca recursiva). Regras que os testes
(`TestMsxlibModulesAssembleConsistently`) impõem: nenhum símbolo `PUBLIC`
duplicado entre módulos, e todo `EXTERN` de um módulo precisa ser `PUBLIC`
de algum outro módulo da biblioteca. Rótulos auxiliares referenciados de
outro módulo são promovidos a `PUBLIC` automaticamente pelo KAJI80 — por
isso rotinas que compartilham dados/rótulos internos (ex.: `VDP_Line` e suas
células de trabalho) ficam juntas num mesmo módulo.

### 6.2. Depuração e garantias do KAJI80

Este projeto já foi mordido mais de uma vez por instruções que o `KAJI80`
aceitava mas codificava **silenciosamente errado** (nenhum erro de
montagem, resultado binário incorreto): operandos indexados `(IX+d)` em
instruções ALU virando um imediato `0` em silêncio (causa raiz de um bug
gráfico real em SCREEN 2), e um literal de caractere `'$'` sendo
desembrulhado incorretamente ao reconstruir o operando de `LD`. Ambos
foram corrigidos e ganharam teste de regressão. A lição registrada no
projeto: o `KAJI80` sempre recusar montagem de uma forma não reconhecida
(erro alto) é preferível a "advinhar" um valor — todo o resto da
documentação segue essa mesma regra (nunca documentar como "funciona" algo
que só falha silenciosamente).

### 6.3. Descritores de API (`lib/api/*.api`): chamar a MSXLIB do BASIC e do Pascal

As rotinas da `MSXLIB` recebem parâmetros em **registradores** (`A`, `BC`, `HL`…),
não na pilha como as `PROCEDURE` de BASIC/Pascal. Um arquivo `.api` descreve, por
rotina, quais parâmetros existem, em que registrador cada um entra e onde o valor
volta; com ele o `DIGNAC` e o `WIRTH80` chamam a rotina **como uma procedure/função
qualquer**, sem comando dedicado no compilador:

```
; lib/api/vdp.api
proc VDP_SetColor(fg: byte in H, bg: byte in L)
proc VDP_FillVRAM(addr: word in HL, count: word in BC, val: byte in A)
func Mul16(a: word in HL, b: word in DE): word out HL
alias basic  Random = MATH_Random8      ; apelido só para BASIC ("pascal" = só Pascal, "all" = ambos)
```

- **Tipos**: `byte` (registrador de 8 bits `A B C D E H L`), `word` e `ptr` (par `BC`,
  `DE` ou `HL`). Retorno de `func`: `A`, `BC`, `DE` ou `HL`. Uma só saída; rotinas com
  várias saídas (ex.: `BDOS_FileOpen` devolve erro em `A` e handle em `B`) ainda não
  são descritas.
- **Uso**: `PSG_Write(7, 62)`, `x% = Mul16(6, 7)` em BASIC; `PSG_Write(7, 62);`,
  `x := Mul16(6, 7);` em Pascal. Funções sem parâmetros levam parênteses vazios em
  expressão (`x% = Random()`). Maiúsculas/minúsculas não importam; o compilador emite
  o nome canônico e declara o `EXTERN` sozinho.
- **Argumentos** são expressões inteiras; um `byte` recebe só o byte baixo do valor.
  Argumento a menos/a mais, ou `proc` usada como expressão, é **erro de compilação**.
- **Precedência**: uma `PROCEDURE`/`EXTERN` do próprio programa com o mesmo nome vence
  a rotina descrita.
- **IX**: as rotinas podem destruir qualquer registrador (menos `SP`); o compilador
  salva/restaura `IX` (frame pointer) em volta de cada chamada.
- **Onde o compilador procura**: `-api <arquivo|diretório>` (repetível) em `dignac` e
  `wirth80`; sem `-api`, usa `../lib/api` ao lado do executável (layout da distribuição:
  `bin/` ao lado de `lib/`). No `OBI`, a chave `api:` do `Obifile`.
- **Exemplo completo**: `sample/api/` (mesmo programa em BASIC e Pascal).

Rotinas de convenção de pilha (`VDP_Line`, `VDP_BoxFill`) já têm comandos próprios
(`LINE`) e não são descritas aqui. As descritas hoje: `lib/api/{bdos,bios,vdp,psg,string,math}.api`.

## 7. O orquestrador de build `OBI`

`OBI` (帯, "faixa que amarra o conjunto") é o "make" do KIZUNA — lê uma
receita declarativa (`Obifile`) e invoca `kaji80`/`wirth80`/`dignac` +
`musubi` (ou `hako`, se o alvo for `.hlib`) na ordem certa, sem precisar
escrever os comandos manualmente.

```bash
obi build [Obifile]   # "Obifile" no diretório atual se omitido
```

| Opção            | Descrição                                                                 |
| ----------------- | ------------------------------------------------------------------------------ |
| `-v`               | Lista cada módulo/resource montado e o resultado da linkagem.                 |
| `--log`            | Gera ou anexa ao log da build (`<Obifile>.log`).                              |
| `--log-file <f>`   | Caminho customizado para o log.                                               |
| `--version`        | Versão atual.                                                                 |
| `-h`, `--help`     | Ajuda completa.                                                               |

### 7.1. Formato do `Obifile`

```yaml
target: <nome>.com | <nome>.hlib   # dita a ferramenta final (MUSUBI ou HAKO)
entry: Start                        # opcional, padrão "Start"
base: 0x0100                        # opcional, padrão 0x0100

resources:
  - file: <caminho>                 # arquivo binário bruto embutido como DATA
    bank: <n>
    symbol: <Nome>                  # opcional, derivado do nome do arquivo se omitido
    size: <NNN|NNNK>                # opcional, checagem de limite (não faz padding)

modules:
  - name: <Nome>                    # opcional, só pra exibição no -v
    source: <caminho>
    compiler: kaji80|wirth80|dignac # opcional, inferido pela extensão (.asm/.pas/.bas)
    bank: <n>                       # opcional, valida contra o BANK declarado no fonte

link:
  map: <caminho.map>                # opcional

library:
  archive: <caminho.hlib>           # forma singular — uma biblioteca
libraries:
  - <caminho.hlib>                  # forma plural — várias

api:
  - <arquivo.api ou diretório>      # descritores de API da MSXLIB p/ DIGNAC/WIRTH80 (§6.3)
```

- `target: *.com` invoca `MUSUBI` no final; `target: *.hlib` invoca `HAKO`
  em vez disso, empacotando os módulos compilados.
- Um `resources:` sem `bank:` explícito, ou com `bank:` divergente do
  `BANK` declarado no fonte do módulo, é pego como erro antes de linkar —
  não silenciosamente ignorado.
- `size:` num resource é só uma **checagem de limite máximo** (o build
  falha se o arquivo for maior) — não faz padding até esse tamanho.

### 7.2. Exemplo real (compila e linka)

`sample/obi/Obifile` — KAJI80 dono do `Start` no banco 0, um módulo DIGNAC
sem `PROCEDURE Main` (só rotinas `PUBLIC`) no banco 2, um resource
(`banner.txt`) embutido no banco 0, e a `MSXLIB` linkada via `.hlib`:

```yaml
target: main.com
entry: Start
base: 0x0100

resources:
  - file: banner.txt
    bank: 0
    symbol: Res_Banner
    size: 256

modules:
  - name: Main
    source: main.asm
    compiler: kaji80
    bank: 0

  - name: ChartLib
    source: chart_lib.bas
    compiler: dignac
    bank: 2

link:
  map: main.map

libraries:
  - ../../lib/msxlib.hlib
```

```bash
obi build sample/obi/Obifile -v --log
```

`MUSUBI` gera o trampolim de banco automaticamente para a chamada cruzando
bancos (`Main` → `ChartLib.Desenhar`).

## 8. Inspeção com `mobdump`

```bash
mobdump <arquivo.mob> | mobdump --version
```

Sem opções — passa o caminho do `.mob` e ele imprime segmentos, símbolos
(`PUBLIC`/`EXTERN`, `PROC`/`DATA`, endereço quando aplicável) e relocações.
Útil pra confirmar o que um compilador realmente gerou antes de linkar, ou
pra depurar um símbolo que não resolveu.

## 9. Como compilar tudo — passo a passo

### 9.1. Via ferramentas individuais

```bash
# 1. Compilar cada fonte pro seu .mob (kaji80 / dignac / wirth80 conforme a linguagem)
kaji80  modulo_asm.asm  -o modulo_asm.mob
dignac  modulo_bas.bas  -o modulo_bas.mob
wirth80 programa.pas    -o programa.mob

# 2. (Opcional) inspecionar antes de linkar
mobdump modulo_asm.mob

# 3. Linkar tudo (+ MSXLIB se precisar) gerando o .com e o mapa de memória
musubi -v -m saida.map -o saida.com modulo_asm.mob modulo_bas.mob lib/msxlib.hlib
```

### 9.2. Via `OBI` (recomendado para projetos com mais de 1-2 módulos)

```bash
obi build Obifile -v --log
```

Um `Obifile` por alvo (`.com` ou `.hlib`) — projetos maiores costumam ter
vários `Obifile`s, um por executável ou biblioteca, e um script externo
(ou outro `Obifile`... por enquanto `OBI` não invoca outro `OBI`
recursivamente) chamando cada um na ordem certa.

### 9.3. Rebuild da MSXLIB

Sempre que algo em `lib/src/` ou `lib/inc/` muda:

```bash
pwsh -File lib/build.ps1
```

## 10. Um único ponto de entrada (Main) por executável

Um `.COM` só pode ter **um** ponto de entrada (`Start` por padrão — ou o
nome passado em `-e`/`entry:`), não importa em qual das três linguagens ele
esteja escrito. As regras, iguais nas três:

- **`KAJI80`**: totalmente manual — só existe `Start` se o programador
  escrever o label e marcar `PUBLIC Start` (ou o nome configurado como
  ponto de entrada) explicitamente. Um módulo Assembly sem isso é sempre
  uma "biblioteca".
- **`DIGNAC`**: automático — `Start` só é gerado se existir
  `PROCEDURE Main` no módulo. Um módulo sem `Main` (como
  `sample/obi/chart_lib.bas`) vira biblioteca pura, sem `Start` nenhum.
- **`WIRTH80`**: automático — `Start` só é gerado se o
  bloco principal `begin...end.` do programa tiver pelo menos um comando.
  Um programa com `begin end.` vazio e algum `PUBLIC` declarado vira
  biblioteca pura, do mesmo jeito que o `DIGNAC`. Antes disso, `WIRTH80`
  gerava `Start` **incondicionalmente**, e um módulo seu nunca conseguia
  entrar no mesmo `.COM` que outro módulo que já definia `Start`.
- **`MUSUBI`**: se, mesmo assim, mais de um módulo linkado definir o mesmo
  símbolo de ponto de entrada, a linkagem é recusada com um erro específico
  ("múltiplos pontos de entrada..."), não um erro genérico de símbolo
  duplicado nem uma escolha silenciosa de qual `Main` vale — decisão
  deliberada (Wilson: "prefiro que exista apenas um único main"). A
  detecção acontece na linkagem (`MUSUBI`), o único ponto onde todos os
  módulos do `.COM` final já estão visíveis juntos.

## 11. Exemplo combinando as três linguagens num mesmo `.COM`

Com `PUBLIC`/`EXTERN` no `WIRTH80` (`docs/manual-pascal.md` §4-5) e a regra
do §10 acima, um módulo `KAJI80` dono do `Start` já linka de verdade com um
módulo-biblioteca `WIRTH80` (sem `Start` próprio) no mesmo `.COM` — testado
e comprovado (`pkg/wirth80/compiler_test.go`,
`TestLinkKaji80AndWirth80LibraryTogether`), não apenas teorizado:

```pascal
{ lib.pas -- biblioteca WIRTH80 pura, sem Start (begin...end. vazio) }
program Lib;
PUBLIC Foo;

procedure Foo(a: Integer);
begin
  WriteLn(a);
end;

begin
end.
```

```bash
wirth80 lib.pas -o lib.mob        # sem Start, só PUBLIC Foo
kaji80 main.asm -o main.mob       # dono do Start, EXTERN Foo, CALL Foo
musubi -o app.com main.mob lib.mob lib/msxlib.hlib
```

O mesmo vale com `DIGNAC` no lugar do `KAJI80` como dono do `Start` — é
literalmente o mesmo mecanismo que já une `KAJI80`+`DIGNAC` em
`sample/obi/Obifile` (§7.2), agora também aberto pro `WIRTH80`. Com
`BANK <n>` também suportado no `WIRTH80` (`docs/manual-pascal.md` §2), o
módulo-biblioteca acima nem precisa ficar no banco comum — `sample/obi/`
é exatamente esse exemplo levado ao fim: `Main` (`KAJI80`, banco 0)
chama `Desenhar` (`DIGNAC`, banco 2) e `Saudacao` (`WIRTH80`, **banco
1**), com o `MUSUBI` gerando os dois trampolins de banco automaticamente
— as três linguagens, cada uma no seu próprio banco, testado e confirmado
em hardware real. `demo/main.pas`/`demo/Obifile` (que ainda usam
`{$USES}`, sintaxe que não existe) continuam sendo só o esboço de uma
visão mais distante — `uses`/units de verdade cruzando arquivos — mas o
mecanismo de exportar/chamar entre as três linguagens num único `.COM`,
cada uma em seu próprio banco, já não é mais teórico.

## 12. Ver também

- `SPEC.md` — decisões de design originais (visão completa do projeto,
  incluindo partes ainda não implementadas).
- `docs/manual-assembly.md`, `docs/manual-basic-dignified.md`,
  `docs/manual-pascal.md` — sintaxe de cada linguagem de entrada.
- `HELP.md` — índice curto apontando pra estes quatro manuais.

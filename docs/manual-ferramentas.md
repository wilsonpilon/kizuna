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

     lib/src/*.asm ──KAJI80──► *.mob ──HAKO──► msxlib.hlib
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
precisa pedir).

### 4.3. Bootstrap multi-banco

Um `.COM` que usa mais de um banco é ainda um único arquivo autocontido: o
bootstrap gerado copia o payload de cada banco pra sua página de RAM
estendida na inicialização, via `ALL_SEG` do EXTBIO quando disponível
(com fallback pro comportamento antigo de mapeamento por identidade).

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

Construída a partir de `lib/src/*.asm` via `lib/build.ps1`. Símbolos
exportados por módulo:

| Módulo | Símbolos | Descrição |
| ------- | --------- | ----------- |
| **`bdos`** | `BDOS_Call`, `BDOS_PrintChar`, `BDOS_PrintString`, `BDOS_ReadChar`, `BDOS_Exit`, `BDOS_FileOpen`, `BDOS_FileCreate`, `BDOS_FileClose`, `BDOS_FileRead`, `BDOS_FileWrite`, `BDOS_FileSeek` | Kernel MSX-DOS (BDOS 0x0005), I/O de arquivo por handle (MSX-DOS 2, funções 43h-4Ah). |
| **`bios`** | `BIOS_Call`, `BIOS_CHPUT`, `BIOS_CHGET`, `BIOS_CLS`, `BIOS_POSIT`, `BIOS_BEEP`, `BIOS_INIT32`, `BIOS_CHGMOD` | Chamadas inter-slot seguras à Main-ROM BIOS via `CALSLT`, preservando o estado do MSX-DOS. |
| **`vdp`** | `VDP_WriteReg`, `VDP_WriteReg_Raw`, `VDP_SetWriteAddr`, `VDP_SetReadAddr`, `VDP_FillVRAM`, `VDP_WriteVRAM`, `VDP_ReadVRAM`, `VDP_CopyVRAM`, `VDP_SetColor`, `VDP_SetScreen`, `VDP_InitScreen0/1/2`, `VDP_InitScreen2_Tables`, `VDP_PSet`, `VDP_PSet_Raw`, `VDP_PSet_HW`, `VDP_CommandWait_Raw`, `VDP_Line`, `VDP_BoxFill`, `VDP_SpriteDefine`, `VDP_SpriteSet`, `VDP_SpriteHide`, `VDP_SpriteHideAll`, `VDP_SpriteSetSize` | V9938/TMS9918: VRAM, paleta, SCREEN 0/1/2, linhas/caixas, sprites (Sprite Mode 1, 8x8/16x16). |
| **`psg`** | `PSG_Write`, `PSG_Read`, `PSG_MuteAll`, `PSG_PlayTone`, `PSG_PlayNoteIndexed`, `PSG_PlaySequence` | AY-3-8910/YM2149: registradores, notas por nome/oitava, sequências de melodia. |
| **`string`** | `StrLen`, `StrCopy`, `StrToUpper`, `PrintHex8`, `PrintHex16`, `PrintDec16`, `PrintDec16ToBuffer`, `StrCopyLen`, `BDOS_PrintLenStr` | Texto terminado em `\0` (rotinas clássicas) e no formato "short string" 1-byte-tamanho+dados (`StrCopyLen`/`BDOS_PrintLenStr`, usadas pelas STRING de verdade do DIGNAC — ver `docs/manual-basic-dignified.md` §3); conversão pra hex/decimal. |
| **`math`** | `Mul16`, `Div16` | Multiplicação/divisão inteira de 16 bits sem sinal. |

### 6.1. Uso com Smart-Linking

```bash
musubi -v -m app.map -o app.com main.mob lib/msxlib.hlib
```

Só os módulos da `MSXLIB` de fato referenciados (direta ou
transitivamente) entram no `.com` final.

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

Sempre que `lib/src/*.asm` muda:

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
`sample/obi/Obifile` (§7.2), agora também aberto pro `WIRTH80`. `demo/
main.pas`/`demo/Obifile` (que ainda usam `{$USES}`, sintaxe que não existe)
continuam sendo só o esboço de uma visão mais distante — `uses`/units de
verdade cruzando arquivos — mas o mecanismo de exportar/chamar entre as três
linguagens num único `.COM` já não é mais teórico.

## 12. Ver também

- `SPEC.md` — decisões de design originais (visão completa do projeto,
  incluindo partes ainda não implementadas).
- `docs/manual-assembly.md`, `docs/manual-basic-dignified.md`,
  `docs/manual-pascal.md` — sintaxe de cada linguagem de entrada.
- `HELP.md` — índice curto apontando pra estes quatro manuais.

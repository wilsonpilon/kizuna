# KIZUNA — Manual da Linguagem Assembly (KAJI80)

> Este manual descreve a sintaxe do Assembly Z80 aceita pelo montador
> **`KAJI80`**, suas diretivas, o subconjunto de instruções suportado e o uso
> da ferramenta de linha de comando. Para o formato de objeto `.MOB` gerado,
> o linker `MUSUBI`, bank switching e o restante da toolchain, veja
> `docs/manual-ferramentas.md`.

## 1. Quando usar Assembly no KIZUNA

`KAJI80` é o único dos três frontends que dá acesso direto a hardware:
portas I/O (`IN`/`OUT`), todos os registradores, timing exato. Use Assembly
para rotinas de baixo nível (a própria `MSXLIB` é inteiramente escrita nele)
e deixe `DIGNAC`/`WIRTH80` para lógica de aplicação — os três produzem o
mesmo formato `.MOB` e linkam juntos sem conversão.

## 2. Estrutura de um módulo

```asm
MODULE NomeDoModulo
BANK 0                      ; 0 = área comum fixa, 1..N = banco paginável
PUBLIC MinhaRotina           ; símbolos exportados
EXTERN BDOS_PrintString      ; símbolos importados (outro módulo ou MSXLIB)

MinhaRotina:
    ; ...
    RET

ENDMOD                       ; opcional
```

Um arquivo-fonte pode conter mais de um `MODULE`/`ENDMOD` (cada `BANK` dentro
dele se aplica aos segmentos declarados a partir daquele ponto).

## 3. Diretivas suportadas

| Diretiva                     | Sintaxe                          | Descrição                                                                                                      |
| ------------------------------ | ----------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `MODULE`                      | `MODULE <nome>`                   | Define o identificador do módulo.                                                                                 |
| `BANK`                        | `BANK <n>`                        | Banco de memória do segmento atual: `0` = área comum (`4000h..7FFFh`), `1..N` = janela paginável (`8000h..BFFFh`). |
| `PUBLIC`                      | `PUBLIC sym1 [, sym2...]`         | Exporta labels para outros módulos e para o linker.                                                               |
| `EXTERN`                      | `EXTERN sym1 [, sym2...]`         | Declara símbolos importados (de outro módulo ou da `MSXLIB`).                                                     |
| `EQU`                         | `<nome> EQU <expressão>`          | Constante simbólica — aceita uma expressão completa (§6), não só um literal isolado. Não ocupa espaço nem gera relocation. |
| `Nome = expressão`            | `<nome> = <expressão>`            | Variável **reatribuível** — diferente de `EQU`, pode ser redefinida quantas vezes quiser (ver §6 e §8).           |
| `ORG`                         | `ORG <endereço>`                  | Ajusta a origem/offset base do segmento atual.                                                                    |
| `DB` / `DEFB` / `BYTE` / `DT` / `DEFT` | `DB item1, item2...`     | Emite bytes, strings literais ou expressões numéricas (§6) — cada item vira 1 byte, ou os bytes literais da string. `DT`/`DEFT` são aliases (compatibilidade com outros assemblers Z80). |
| `DW` / `DEFW` / `WORD`        | `DW val1, val2...`                | Palavras de 16 bits em little-endian — expressão numérica (§6) ou label (gera relocation `ABS16`).                |
| `DS` / `DEFS` / `BLKB`        | `DS <tamanho>`                    | Reserva `N` bytes preenchidos com zero.                                                                            |
| `IF` / `ELSE` / `ENDIF`       | `IF <condição>` ... `ENDIF`       | Montagem condicional — ver §7.                                                                                    |
| `REPT` / `ENDR`               | `REPT <n>` ... `ENDR`             | Repete um bloco de linhas `n` vezes — ver §8.                                                                     |
| `MACRO` / `ENDM`              | `Nome: MACRO @p1, ...` ... `ENDM` | Declara uma macro reutilizável — ver §9.                                                                          |
| `CALLBIOS`                    | `CALLBIOS <rotina>`               | Chama uma rotina da BIOS principal via inter-slot — ver §11.                                                      |
| `CALLDOS`                     | `CALLDOS <código>`                | Chama uma função do MSX-DOS/MSX-DOS2 — ver §11.                                                                   |
| `INCBIN`                      | `INCBIN "arquivo"[, SKIP=x][, SIZE=y]` | Injeta o conteúdo bruto de um arquivo binário — ver §12.                                                     |
| `INCLUDE`                     | `INCLUDE "arquivo.inc"`            | Insere o texto de outro arquivo-fonte (constantes/macros compartilhadas) — ver §12.                                |
| `ENDMOD` / `END`              | `ENDMOD`                          | Finaliza a declaração do módulo (opcional).                                                                       |

## 4. Sintaxe e literais

### Comentários

```asm
; até o fim da linha
ld a, 2 ; comentário inline
```

### Números

- Hexadecimal: `0x100`, `100h`, `100H`, `$100`, `#100`
- Binário: `%10100111`, `0b10100111`, `10100111b`
- Octal: `17o`, `17O` (sufixo apenas — a forma "0" na frente do asMSX foi
  deixada de fora de propósito, ambígua com decimal)
- Decimal: `42`, `255`, `0`
- Decimal de ponto flutuante: `3.14`, `45.0` — só faz sentido dentro de uma
  expressão (§6), normalmente como argumento de uma função matemática ou de
  `FIX()`; o Z80 não tem aritmética de ponto flutuante nativa.

### Strings e literais de caractere

Delimitadas por aspas duplas ou simples, com escapes:

```asm
db "Hello\r\n", 0
db 'MSX', 0x0D, 0x0A, '$'
ld (hl), '$'          ; literal de caractere isolado também funciona como operando de LD/CP/etc.
```

## 5. Rótulos locais

Um rótulo com ponto na frente (`.nome:`) é **local** — válido só a partir
do rótulo global (sem ponto) mais recente até o próximo. Internamente é
renomeado para `<Global>_nome` (maiúsculas/minúsculas preservadas), então
não há custo nenhum em runtime — é só uma forma de reusar nomes triviais
como `.loop`/`.done` em várias rotinas sem precisar inventar um nome
diferente toda vez.

```asm
Funcao1:
.loop:
    NOP
    JR .loop        ; salta pro .loop DESTA função

Funcao2:
.loop:               ; nome igual, escopo diferente -- não colide
    NOP
    JR .loop
```

Um rótulo local usado antes de qualquer rótulo global no arquivo é erro de
montagem claro. `.loop` dentro de um bloco `REPT` (§8) ou de uma macro (§9)
ganha automaticamente um sufixo único por iteração/invocação, então duas
cópias nunca colidem entre si, mesmo com o mesmo nome.

## 6. Avaliador de expressões numéricas

`EQU`, `Nome = expressão`, `DB`/`DW`, a condição de `IF` e os operandos de
`CALLBIOS`/`CALLDOS` aceitam uma expressão numérica completa, avaliada em
tempo de montagem — não só um literal isolado.

**Escopo deliberado: só expressões numéricas puras.** Misturar aritmética
com um rótulo/símbolo relocável (`EQU X + MinhaLabel`, por exemplo) não é
suportado e dá erro de compilação claro — o endereço de um símbolo
`EXTERN` só é conhecido depois da linkagem pelo `MUSUBI`, e misturar isso
com aritmética exigiria estender o formato de relocation do `.MOB` pra
carregar um deslocamento, mudança maior fora do escopo atual. Um nome
isolado que não é uma constante `EQU`/variável conhecida (`LD HL,
MinhaLabel`) continua funcionando normalmente como rótulo/símbolo comum —
só some quando aparece **dentro** de uma expressão maior com operador.

### Operadores (precedência igual ao C, do mais apertado pro mais solto)

```
unário: - + NOT ~
*  /  MOD
+  -
<<  >>
<  <=  >  >=
==  !=
&
^
|
&&
||
```

**Duas diferenças deliberadas do asMSX**, pra não quebrar sintaxe que o
KAJI80 já tinha antes desta leva:

- **`MOD`** é a palavra-chave de módulo, não `%` — `%` já é o prefixo de
  literal binário do KAJI80 (`%1010`).
- Bits: `<<`, `>>`, `|`, `&`, `^`, `~` (complemento binário) — lógicos:
  `&&`, `||`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `NOT` (negação lógica,
  devolve `1`/`0`).

```asm
VAL EQU ((2*8)/(1+3))<<2      ; 16
FLAG EQU 1
IF FLAG == 1 && VAL > 10
    ; ...
ENDIF
```

### Funções matemáticas e constante `PI`

```
SIN(x)  COS(x)  TAN(x)  ASIN(x)  ACOS(x)  ATAN(x)
SQR(x)  SQRT(x)  EXP(x)  LOG(x)  LN(x)  ABS(x)
POW(x,y)
```

`PI` é a constante `π` (double precision). Ângulos em radianos.

```asm
ANGULO EQU sin(pi*45.0/180.0)   ; ~0.7071 -- ver FIX() abaixo pra usar no Z80
```

### Ponto fixo (`FIX`/`FIXMUL`/`FIXDIV`) e `INT`

O Z80 não tem ponto flutuante nativo — um resultado de `SIN`/`COS`/etc.
só é útil dentro de um byte/word depois de convertido pra **ponto fixo
8.8** (1 byte de parte inteira, 1 byte de fração, convenção comum em jogos
Z80):

- `FIX(x)` — converte um float pra sua representação 8.8 (um inteiro de
  16 bits pronto pra `DW`).
- `FIXMUL(a,b)` / `FIXDIV(a,b)` — multiplica/divide dois valores **já em
  ponto fixo** (ex.: `FIXMUL(FIX(1.5), FIX(2.0))`), não floats crus.
- `INT(x)` — trunca um float pra inteiro.

```asm
TABELA_SENO:
X = 0
REPT 91
    DW FIX(sin(X*pi/180.0))
X = X + 1
ENDR
```

### `RANDOM(n)`

Número pseudoaleatório inteiro no intervalo `[0, n)`, gerado **em tempo de
montagem** (uma constante fixa embutida no binário final, não algo que
muda a cada execução no MSX — útil pra tabelas de "ruído"/preenchimento
variado sem escrever os valores à mão).

```asm
DB RANDOM(256)   ; um byte "aleatório" fixo, decidido na hora de montar
```

## 7. Montagem condicional: `IF` / `ELSE` / `ENDIF`

```asm
FORMATO = 1
IF FORMATO == 1
    ; código do formato 1
ELSE
    ; código do formato 2
ENDIF
```

Condição não-zero é verdadeira (usa o avaliador de expressões, §6). `ELSE`
é opcional, `ENDIF` é obrigatório. `IF`/`ELSE`/`ENDIF` precisam estar
sozinhos na linha, sem instrução junto. Aninhamento sem limite artificial.
Um ramo descartado nem chega a existir pras fases seguintes (rótulos
locais, `REPT`, `MACRO`) — uma condição referenciando algo não definido
dentro de um ramo já morto (por já estar dentro de outro `IF` falso) não é
avaliada, não dá erro.

## 8. Repetição: `REPT` / `ENDR`

```asm
REPT 4
    NOP
ENDR
```

Duplica o bloco de linhas `n` vezes. `n` **precisa ser um literal
inteiro** — não pode vir de uma expressão (mesma restrição do asMSX).
Aninhamento permitido. Uma variável reatribuível (`Nome = expressão`, §3)
mutada dentro do bloco funciona normalmente, reavaliada a cada cópia — o
padrão clássico pra gerar tabelas:

```asm
X = 0
Y = 0
REPT 10
    REPT 10
        DB X*Y
X = X + 1
    ENDR
Y = Y + 1
ENDR
```

Um rótulo local (§5) dentro do bloco repetido ganha um sufixo único por
iteração automaticamente — nunca colide entre cópias.

## 9. Macros: `MACRO` / `ENDM`

```asm
m_INC16: MACRO @VARIAVEL
    PUSH HL
    LD HL, @VARIAVEL
    INC (HL)
    POP HL
ENDM

; uso:
m_INC16 MinhaVariavel
MinhaVariavel: DB 0
```

Declara com `Nome: MACRO @param1, @param2, ...`, corpo até `ENDM`. Chama
com `Nome arg1, arg2, ...` — cada `@param` é substituído pelo texto do
argumento **em qualquer lugar** que apareça no corpo, inclusive dentro de
um identificador maior:

```asm
m_INCVALUE_MAX_RESET: MACRO @VARIAVEL, @MAX, @RESET
    LD A, (@VARIAVEL)
    INC A
    CP @MAX
    JR NZ, .naoreseta_@VARIAVEL
    LD A, @RESET
.naoreseta_@VARIAVEL:
    LD (@VARIAVEL), A
ENDM
; chamado como "m_INCVALUE_MAX_RESET Pontuacao, 100, 0", ".naoreseta_@VARIAVEL"
; vira ".naoreseta_Pontuacao" -- rótulo local de verdade, sem colidir com
; outra invocação da mesma macro noutro lugar (cada invocação ganha seu
; próprio sufixo de expansão, igual ao REPT).
```

**Diferença deliberada do asMSX**: o marcador de parâmetro é `@nome`, não
`#nome` — `#` já é o prefixo de literal hexadecimal do KAJI80 (`#100` =
256). Precisa declarar a macro antes de usar (mesma regra de
declare-antes-de-usar do resto do projeto). O corpo de uma macro pode
chamar `IF`/`REPT`/outra macro normalmente.

## 10. Rótulos pré-definidos: BIOS, BDOS e variáveis de sistema

Três tabelas de símbolos sempre disponíveis, sem precisar de `EXTERN` nem
nenhuma declaração — resolvidas como **último recurso**, depois de
qualquer rótulo/constante/variável que você já tenha definido (seu código
sempre vence, sombreamento é silencioso e sem erro):

- **Rotinas da BIOS principal** — endereços de chamada conhecidos
  (MSX/MSX2/MSX2+/Turbo-R), ex.: `CHGMOD`, `CALSLT`, `CHPUT`, `RDVDP`,
  `WRTVDP`. Lista completa em `pkg/kaji80/predefined.go` (`biosLabels`).
- **Variáveis de sistema da BIOS** — ex.: `EXPTBL`, `SCRMOD`, `FORCLR`,
  `LINL32`. Lista completa em `predefined.go` (`biosVars`).
- **Funções do MSX-DOS/MSX-DOS2** — códigos pra `CALLDOS`/`CALL 0005h`,
  ex.: `F_CONOUT` (02h), `F_STROUT` (09h), `F_OPEN` (43h), `F_CLOSE`
  (45h), `F_READ` (48h), `F_WRITE` (49h). Núcleo deliberadamente pequeno
  e conservador (só o que já está verificado em hardware) em
  `predefined.go` (`bdosFuncs`) — não uma lista exaustiva de toda função
  de MSX-DOS.

```asm
LD A, 1
CALLBIOS CHGMOD    ; CHGMOD já é um endereço conhecido, sem EXTERN nenhum
```

## 11. `CALLBIOS` / `CALLDOS`

```asm
CALLDOS F_STROUT      ; LD C,F_STROUT / CALL 0005h -- sempre inlined (5 bytes)
CALLBIOS CHGMOD        ; LD IX,CHGMOD / CALL BIOS_Call (7 bytes)
```

- **`CALLDOS <código>`** — sempre inlined: `LD C,código` / `CALL 0005h`.
  Aceita um código pré-definido (§10) ou um literal numérico.
- **`CALLBIOS <rotina>`** — `LD IX,rotina` / `CALL BIOS_Call`, reusando a
  rotina de chamada inter-slot já testada em hardware
  (`lib/src/bios.asm`, `BIOS_Call`) em vez de repetir a sequência
  completa toda vez — registra `EXTERN BIOS_Call` automaticamente
  (precisa linkar contra `msxlib.hlib`). Aceita uma rotina pré-definida
  (§10) ou qualquer símbolo/expressão que resolva pra um endereço.

## 12. `INCBIN` e `INCLUDE` — incluir arquivos

```asm
Sprites:
    INCBIN "sprites.bin"

Paleta:
    INCBIN "dados.bin", SKIP=16, SIZE=32   ; pula um cabeçalho de 16 bytes, inclui só 32
```

Injeta o conteúdo bruto de um arquivo no objeto, no lugar exato da
diretiva — direto, sem expandir pra uma lista de `DB` em texto (rápido
mesmo pra arquivos maiores, tipo sprites/tiles/samples). `SKIP=x` e
`SIZE=y` são opcionais, em qualquer ordem, separados por vírgula.
**Sintaxe deliberadamente diferente do asMSX**: `SKIP=x`/`SIZE=y` (usando
`=`) em vez de `SKIP x`/`SIZE y` com espaço — evita uma ambiguidade real
na reconstrução de operando do KAJI80.

O caminho é resolvido relativo ao diretório do próprio arquivo-fonte
(`.asm`), não ao diretório de onde você roda `kaji80` — um caminho
absoluto também funciona.

### `INCLUDE` — incluir outro arquivo-fonte

```asm
INCLUDE "vdp.inc"          ; constantes VDP_DATA, VDP_CMD, ...
INCLUDE "../inc/psg.inc"   ; caminho relativo ao arquivo que contém o INCLUDE
```

Substitui a linha pelo **texto** do arquivo indicado, antes de qualquer
outra etapa (IF/REPT/MACRO/rótulos locais já enxergam o texto incluído).
Serve pra compartilhar `EQU`, macros e variáveis entre vários módulos sem
copiar e colar — é assim que a MSXLIB divide as constantes de VDP/PSG/BDOS
entre dezenas de módulos pequenos. Regras:

- O caminho é relativo ao diretório do arquivo que contém o `INCLUDE`
  (includes aninhados resolvem em relação ao incluidor, não ao arquivo
  principal); caminho absoluto também funciona.
- Aceita aspas duplas ou simples e um comentário `;` no fim da linha.
- Inclusão circular é erro ("include circular"), e há um limite de
  profundidade de 16 níveis.
- Arquivo inexistente é erro de montagem, nunca silencioso.
- Limitação: depois da expansão, o "linha N" das mensagens de erro conta
  as linhas do texto já expandido.

## 13. Conjunto de instruções Z80 suportadas

### Controle e estado

`NOP`, `HALT`, `DI`, `EI`, `EXX`, `EX DE,HL`, `EX AF,AF'`

### Fluxo e chamadas

- `RET` e `RET cc` (`NZ`,`Z`,`NC`,`C`,`PO`,`PE`,`P`,`M`)
- `CALL nn` / `CALL sym` e `CALL cc, nn` (gera relocation `ABS16`)
- `JP nn` / `JP sym` e `JP cc, nn`
- `JP (HL)`, `JP (IX)`, `JP (IY)`
- `JR e`, `JR cc, e` (`NZ`,`Z`,`NC`,`C`) — gera relocation `REL8` se for cross-label
- `DJNZ e`

### Pilha

`PUSH`/`POP` para `BC`, `DE`, `HL`, `AF`, `IX`, `IY`

### Entrada e saída (I/O)

`IN A,(n)`, `OUT (n),A`

### Aritmética e lógica

- `INC r` / `DEC r` (8 bits: `A,B,C,D,E,H,L`) e `INC (HL)` / `DEC (HL)`
- `INC rr` / `DEC rr` (16 bits: `BC,DE,HL,SP,IX,IY`)
- `ADD A,...`, `ADC A,...`, `SUB ...`, `SBC A,...`, `AND ...`, `XOR ...`, `OR ...`, `CP ...`
  - Operando registrador (`r` ou `(HL)`), imediato (`n` ou constante `EQU`), **ou indexado** `(IX+d)`/`(IY+d)`
- `ADD HL,rr` (`BC,DE,HL,SP`) e `ADD IX,rr` (`BC,DE,IX,SP`)

### Movimentação de dados (`LD`)

- `LD r,r'`, `LD r,n`, `LD r,(HL)`, `LD (HL),r`, `LD (HL),n`
- `LD A,(BC)` / `LD A,(DE)` / `LD (BC),A` / `LD (DE),A`
- `LD A,(nn)` / `LD (nn),A` e `LD HL,(nn)` / `LD (nn),HL`
- `LD rr,nn` (`BC,DE,HL,SP`), `LD IX,nn`, `LD IY,nn` — gera relocation `ABS16` se `nn` for label
- `LD SP,HL`

### O que **não** é suportado (erro na montagem, nunca silencioso)

- `LD (nn),DE` / `LD (nn),BC` — só `(nn),HL` e `(nn),A` têm forma de 16 bits absoluta. Contorne com `EX DE,HL` antes do `LD (nn),HL`.
- `LDIR`/`LDDR`/`CPIR`/`CPDR` e demais instruções de bloco — para cópias de tamanho conhecido em tempo de compilação, desenrole o laço manualmente (`LD A,(HL)` / `LD (DE),A` / `INC HL` / `INC DE` repetido, opcionalmente com `REPT`, §8).

Se você tentar uma dessas formas, o `KAJI80` recusa a montagem com uma
mensagem de erro — ele nunca cai silenciosamente para uma codificação
diferente da que você escreveu (essa garantia já foi a causa raiz de mais de
um bug real neste projeto — ver `docs/manual-ferramentas.md` §"Depuração e
garantias do KAJI80").

## 14. O montador `KAJI80` (linha de comando)

```bash
kaji80 [opções] <arquivo.asm>
```

| Opção            | Descrição                                                                          |
| ----------------- | ------------------------------------------------------------------------------------ |
| `-o <saida.mob>`  | Nome/caminho do `.mob` de saída (padrão: mesmo nome com extensão `.mob`).             |
| `-v`               | Modo detalhado: contagem de segmentos, símbolos, relocações e bancos.                |
| `--version`        | Exibe a versão atual.                                                                |
| `-h`, `--help`     | Ajuda completa.                                                                       |

```bash
# Montagem padrão
kaji80 sample/hello.asm

# Com saída explícita e modo detalhado
kaji80 -v -o build/hello.mob sample/hello.asm
```

O `.mob` gerado é consumido pelo linker `MUSUBI` (ou empacotado numa
biblioteca `.hlib` pelo `HAKO`) — veja `docs/manual-ferramentas.md`.

## 15. Garantias internas do montador

O `KAJI80` faz uma verificação de consistência interna a cada montagem:
roda a estimativa de tamanho (Pass 1) e a emissão real (Pass 2)
independentemente e compara o tamanho final de cada segmento — uma
divergência denuncia um bug de dessincronia entre os dois passes antes que o
`.mob` saia corrompido. Essa rede de segurança já pegou bugs reais no
histórico do projeto (ver `CHANGELOG.md`), mas ela só detecta diferença de
*tamanho*, não de *significado* — por isso os casos listados acima (como
`(IX+d)` em instruções ALU, que antes virava silenciosamente um imediato
`0`) precisaram de correção explícita no encoder, não só da rede de
segurança.

## 16. Ver também

- `sample/macroasm/` — três programas reais (`expr_labels.asm`,
  `predefined.asm`, `incbin.asm`) exercitando os recursos das §5-12,
  cada um imprimindo `[OK]`/`[FALHOU]` visível no console pra
  confirmação em hardware/openMSX.
- `docs/manual-ferramentas.md` — formato `.MOB`, `MUSUBI` (linker), bank
  switching, `HAKO`/`.HLIB`, `MSXLIB`, `OBI` e o ciclo completo de build.
- `docs/manual-basic-dignified.md` / `docs/manual-pascal.md` — as outras
  duas linguagens de entrada, que compilam para o mesmo Assembly que este
  manual descreve (use `-S` em `dignac`/`wirth80` para ver o Assembly
  gerado).

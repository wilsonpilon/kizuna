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

| Diretiva               | Sintaxe                   | Descrição                                                                                                      |
| ----------------------- | -------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `MODULE`                | `MODULE <nome>`            | Define o identificador do módulo.                                                                                 |
| `BANK`                  | `BANK <n>`                 | Banco de memória do segmento atual: `0` = área comum (`4000h..7FFFh`), `1..N` = janela paginável (`8000h..BFFFh`). |
| `PUBLIC`                | `PUBLIC sym1 [, sym2...]`  | Exporta labels para outros módulos e para o linker.                                                               |
| `EXTERN`                | `EXTERN sym1 [, sym2...]`  | Declara símbolos importados (de outro módulo ou da `MSXLIB`).                                                     |
| `EQU`                   | `<nome> EQU <valor>`       | Constante simbólica — não ocupa espaço nem gera relocation.                                                       |
| `ORG`                   | `ORG <endereço>`           | Ajusta a origem/offset base do segmento atual.                                                                    |
| `DB` / `DEFB` / `BYTE`  | `DB item1, item2...`       | Emite bytes ou strings literais.                                                                                  |
| `DW` / `DEFW` / `WORD`  | `DW val1, val2...`         | Palavras de 16 bits em little-endian. Um label como valor gera relocation `ABS16`.                                |
| `DS` / `DEFS` / `BLKB`  | `DS <tamanho>`             | Reserva `N` bytes preenchidos com zero.                                                                            |
| `ENDMOD` / `END`        | `ENDMOD`                   | Finaliza a declaração do módulo (opcional).                                                                       |

## 4. Sintaxe e literais

### Comentários

```asm
; até o fim da linha
ld a, 2 ; comentário inline
```

### Números

- Hexadecimal: `0x100`, `100h`, `100H`, `$100`, `#100`
- Binário: `%10100111`, `0b10100111`, `10100111b`
- Decimal: `42`, `255`, `0`

### Strings e literais de caractere

Delimitadas por aspas duplas ou simples, com escapes:

```asm
db "Hello\r\n", 0
db 'MSX', 0x0D, 0x0A, '$'
ld (hl), '$'          ; literal de caractere isolado também funciona como operando de LD/CP/etc.
```

## 5. Conjunto de instruções Z80 suportadas

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

- `INC r` / `DEC r` (8 bits: `A,B,C,D,E,H,L`) — **não** suporta `INC (HL)`/`DEC (HL)` (erro claro na montagem, não silencioso)
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
- `INC (HL)` / `DEC (HL)`
- `LDIR`/`LDDR`/`CPIR`/`CPDR` e demais instruções de bloco — para cópias de tamanho conhecido em tempo de compilação, desenrole o laço manualmente (`LD A,(HL)` / `LD (DE),A` / `INC HL` / `INC DE` repetido).

Se você tentar uma dessas formas, o `KAJI80` recusa a montagem com uma
mensagem de erro — ele nunca cai silenciosamente para uma codificação
diferente da que você escreveu (essa garantia já foi a causa raiz de mais de
um bug real neste projeto — ver `docs/manual-ferramentas.md` §"Depuração e
garantias do KAJI80").

## 6. O montador `KAJI80` (linha de comando)

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

## 7. Garantias internas do montador

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

## 8. Ver também

- `docs/manual-ferramentas.md` — formato `.MOB`, `MUSUBI` (linker), bank
  switching, `HAKO`/`.HLIB`, `MSXLIB`, `OBI` e o ciclo completo de build.
- `docs/manual-basic-dignified.md` / `docs/manual-pascal.md` — as outras
  duas linguagens de entrada, que compilam para o mesmo Assembly que este
  manual descreve (use `-S` em `dignac`/`wirth80` para ver o Assembly
  gerado).

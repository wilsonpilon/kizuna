# KIZUNA — Manual da Linguagem MSX-BASIC Dignified (DIGNAC)

> Este manual descreve a sintaxe do **MSX-BASIC Dignified**, a linguagem
> compilada (não interpretada) do KIZUNA que segue de perto os dialetos
> Microsoft de MSX-BASIC — mas sem as amarras de compatibilidade binária com
> um interpretador real: variáveis com nomes longos, tipos explícitos e
> compilação para `.MOB` nativo. O compilador é o `DIGNAC`. Para o formato de
> objeto, o linker e bank switching, veja `docs/manual-ferramentas.md`.

## 1. Filosofia

DIGNAC mira a sintaxe do MSX-BASIC clássico (o mesmo `PRINT`, `FOR...NEXT`,
`IF...THEN`, sufixos `%`/`$`/`!`/`#`) para ficar reconhecível e fácil de
documentar, mas se afasta dele sempre que a compilação nativa permite algo
melhor: nomes de variável maiores que 2 caracteres, tipos reais (não
"tudo é uma palavra de 16 bits"), módulos com `PUBLIC`/`EXTERN` como o
Assembly, e erros de compilação claros em vez do comportamento
"undefined" clássico do BASIC interpretado.

## 2. Estrutura de um módulo

```basic
MODULE NomeDoModulo
BANK 0
PUBLIC Main
EXTERN BIOS_CHGET

DIM contador%, mensagem$        ' variáveis globais (escopo do módulo)

PROCEDURE Main()
    PRINT "Ola, mundo!"
END PROCEDURE
END MODULE
```

- `MODULE <nome>` / `END MODULE`: delimita a unidade de compilação.
- `BANK <n>`: banco de alocação (`0` = área comum fixa, `1..N` = banco
  paginável) — mesmo modelo do Assembly.
- `PUBLIC sym1 [, sym2...]` / `EXTERN sym1 [, sym2...]`: exporta/importa
  símbolos, exatamente como no KAJI80.
- Se existir `PROCEDURE Main`, o compilador gera automaticamente o ponto de
  entrada `Start` — o `.mob` resultante já pode virar um `.COM` sozinho, sem
  precisar de um módulo Assembly adicional.

## 3. Sistema de tipos

| Sufixo | Tipo      | Faixa / tamanho                                             |
| ------ | --------- | ------------------------------------------------------------- |
| `%`    | INTEGER   | -32768..32767 (16 bits, palavra)                              |
| `$`    | STRING    | até 255 caracteres (1 byte de tamanho + dados, 256 bytes no total) |
| `!`    | SINGLE    | ponto flutuante IEEE 754 binary32 (4 bytes)                    |
| `#`    | DOUBLE    | ponto flutuante IEEE 754 binary64 (8 bytes)                    |
| (nenhum, via `AS BOOLEAN`) | BOOLEAN | 1 byte |

Sem sufixo e sem `AS <Tipo>` explícito, uma variável é `INTEGER` por padrão.

```basic
DIM nome$, idade%, altura!, pi#
LOCAL a, b AS BOOLEAN          ' AS no fim força o tipo pra toda a lista
```

Nomes de variável não têm limite de 2 caracteres como o MSX-BASIC clássico —
podem ser tão longos quanto quiser (`DIM pontuacaoTotal%` é válido).

### STRING — funciona de ponta a ponta

Declarar, atribuir (literal ou outra variável STRING) e imprimir:

```basic
DIM nome$
PROCEDURE Main()
    nome$ = "Kizuna"
    PRINT nome$
END PROCEDURE
```

Internamente é um buffer de 256 bytes: 1 byte de tamanho + até 255 bytes de
dados — o mesmo formato "short string" usado nas fronteiras entre as três
linguagens (ver `docs/manual-ferramentas.md`). Limitações atuais:
- **Sem concatenação** (`s$ = a$ + b$` ainda não existe).
- **Não pode ser passada como parâmetro de procedimento** — a ABI de
  parâmetros hoje assume 2 bytes por parâmetro; passar STRING/SINGLE/DOUBLE
  fica para uma leva futura.
- Só pode aparecer como alvo de atribuição/`PRINT`, ou como origem de uma
  atribuição para outra STRING — usá-la em qualquer outra expressão
  (`IF nome$ = ...`, aritmética, argumento de função) é um erro de
  compilação claro, não um resultado errado.

### SINGLE — declarar, `+`/`-`/comparação com operandos simples

```basic
DIM a!, b!, soma!
PROCEDURE Main()
    a! = 1.0
    b! = 0.5
    soma! = a! + b!         ' Float_Add32
    IF soma! > a! THEN      ' Float_Cmp32
        PRINT "maior"
    END IF
END PROCEDURE
```

`SINGLE` (IEEE 754 binary32) tem um motor de aritmética de verdade em
`lib/src/float.asm` (`Float_Add32`/`Float_Sub32`/`Float_Cmp32`,
MSXLIB). Funciona:

- Declarar, atribuir por cópia (literal ou outra variável do mesmo tipo).
- `x! = a! + b!` e `x! = a! - b!` — soma/subtração de verdade.
- Comparação (`=`, `<>`, `<`, `<=`, `>`, `>=`) entre duas variáveis
  `SINGLE`, ou uma variável `SINGLE` e um literal numérico.

**Limitações desta leva, aceitas e documentadas (não escondidas)**:

- **Só operandos simples ("expressões chatas")**: `x! = a! + b!` funciona,
  `x! = (a! + b!) - c!` **não** (erro de compilação claro) — não existe
  alocação de temporários ainda, então uma sub-expressão float aninhada
  dentro de outra é recusada, não computada errado.
- **`*` e `/` ainda não implementados** para `SINGLE` — erro de compilação
  claro ("só '+' e '-' estão implementados... nesta leva"). A normalização
  do produto/quociente de mantissa de 24 bits se mostrou bem mais delicada
  do que somar/subtrair durante a implementação (é fácil inverter a
  direção do ajuste de expoente por engano), e a decisão foi entregar
  soma/subtração/comparação **verificadas** agora e deixar `*`/`/` para uma
  leva futura, em vez de arriscar um bloco de código malfeito.
- **`DOUBLE` continua só declarável/atribuível por cópia** — nenhuma
  aritmética, mesmo erro de antes.
- **`PRINT` de `SINGLE`/`DOUBLE` continua um erro de compilação** — a
  conversão de ponto flutuante pra texto decimal é o item mais difícil de
  todos, fica pra uma leva própria.
- **Comparação mista** (um lado `SINGLE`, outro `INTEGER`, ou `SINGLE`
  contra `DOUBLE`) é erro de compilação claro, não um resultado errado.
- **Sem bit de guarda/arredondamento fino**: os deslocamentos de
  alinhamento/normalização truncam em vez de arredondar ao mais próximo —
  parte dos resultados fica 1 ULP (a última casa binária) diferente do
  IEEE754 "de verdade". Em subtrações de magnitudes muito próximas
  (cancelamento catastrófico) esse erro de 1 bit pode ser amplificado pela
  normalização — também aceito nesta leva.
- **Sem tratamento de `Infinity`/`NaN`/overflow de expoente**; subnormais
  são tratados como zero.

Exemplo real, hardware-testável (usa `IF`/`PRINT` de texto pra tornar o
resultado observável sem depender de `PRINT` de float):
`sample/basic/floatmath.bas`.

## 4. Literais numéricos

```basic
42          ' decimal
&H2A        ' hexadecimal (também aceita $2A com dígito hex logo em seguida)
&O52        ' octal
&B101010    ' binário
3.14        ' SINGLE (ponto decimal, sem sufixo/expoente = SINGLE)
1.5e+10     ' SINGLE explícito via expoente 'e'/'E'
3.14159d0   ' DOUBLE via expoente 'd'/'D'
42!         ' SINGLE via sufixo
42#         ' DOUBLE via sufixo
```

Strings literais vão entre aspas duplas, até 255 caracteres; `""` dentro de
uma string representa uma aspa literal (`"disse ""oi"""`).

## 5. Expressões e operadores

| Categoria    | Operadores                              |
| ------------- | ----------------------------------------- |
| Aritméticos   | `+`, `-`, `*` (via `Mul16`), `/` e `MOD` (via `Div16`) |
| Lógicos       | `AND`, `OR`, `XOR`, `NOT`                |
| Relacionais   | `=`, `<>`, `<`, `<=`, `>`, `>=`           |

**Atenção à precedência**: operadores relacionais têm precedência **mais
baixa** que `AND`/`OR`/`XOR` (o oposto do costume em outras linguagens, onde
comparação normalmente vem primeiro). Isso significa que
`a = 1 AND b = 2` **não** parseia como `(a = 1) AND (b = 2)` — parenteize
sempre que misturar comparação com operador lógico na mesma expressão:

```basic
IF (a% = 1) AND (b% = 2) THEN ...   ' correto
IF a% = 1 AND b% = 2 THEN ...       ' NÃO faz o que parece -- evite
```

## 6. Controle de fluxo

```basic
FOR i% = 1 TO 10 [STEP passo]
    ...
NEXT [i%]

IF condicao THEN
    ...
ELSE
    ...
END IF
' ou de linha única: IF condicao THEN comando [ELSE comando]

WHILE condicao
    ...
WEND

DO [WHILE condicao]
    ...
LOOP
```

`STEP` negativo funciona corretamente desde a v4.8.0 — o teste de término
do laço calcula o sinal do `STEP` uma vez antes de entrar no laço e o
consulta a cada iteração pra decidir se `Var > End` ou `Var < End` é quem
encerra:

```basic
FOR i% = 5 TO 1 STEP -1
    PRINT i%
NEXT i%   ' imprime 5, 4, 3, 2, 1
```

`DO...LOOP` só suporta o teste **antes** do corpo (`DO WHILE cond ... LOOP`)
— não existe a forma `DO ... LOOP WHILE cond` (teste no final) ainda. Um
`DO` sem `WHILE` é um laço infinito (não há `EXIT FOR`/`EXIT DO` — a
palavra-chave `EXIT` está reservada mas ainda não tem efeito nenhum no
parser).

## 7. Sub-rotinas

```basic
PROCEDURE Nome(par1%, par2$)
    LOCAL total%
    ...
END PROCEDURE
' ou: SUB Nome(...) ... END SUB (sinônimo de PROCEDURE)

FUNCTION Calc(n%) AS INTEGER
    RETURN n% * 2
END FUNCTION
```

- Parâmetros são empilhados esquerda→direita; quadro de ativação usa `IX`
  como frame pointer (mesma ABI usada pelo Assembly — ver
  `docs/manual-ferramentas.md` §7).
- `LOCAL var1%, var2$`: variáveis locais no quadro de pilha.
- Chamar um procedimento como comando: `NomeDoProc(arg1, arg2)`.

## 8. Comandos de tela e I/O de console

```basic
PRINT expr1, expr2; expr3    ' ; suprime a quebra de linha final; , também separa argumentos
CLS
BEEP
SCREEN modo
```

`COLOR fg, bg, bd` existe na AST mas **ainda não está ligado ao parser** —
escrever `COLOR ...` hoje dá erro de sintaxe ("instrução desconhecida"), não
um comando silenciosamente ignorado.

## 9. Primitivas gráficas (SCREEN 2)

```basic
LINE (x1, y1)-(x2, y2)[, cor][, B | BF]   ' B = retângulo vazado, BF = preenchido
PSET (x, y)[, cor]
```

## 10. Sprites

```basic
SPRITE PATTERN indice#, b0, b1, ..., bN   ' 8 bytes (8x8) ou 32 bytes (16x16), tudo literal numérico
PUT SPRITE indice, (x, y), cor, padrao
SPRITE OFF                                 ' oculta todos
```

`SPRITE PATTERN` é uma **deliberada divergência** do MSX-BASIC clássico
(que usa `SPRITE$(n)=...` via concatenação de `CHR$`) — impraticável numa
linguagem compilada sem strings dinâmicas de verdade.

## 11. Música (PLAY / MML)

```basic
PLAY "O4 L4 CEG O5 C"
```

A tradução MML **acontece em tempo de compilação** — por isso `PLAY` exige
uma **string literal**, nunca uma variável: não existe interpretador de MML
em tempo de execução no Z80, o compilador já traduz direto para uma tabela
de eventos `PSG_PlaySequence`. Subconjunto MML suportado:

| Elemento | Significado |
| --------- | ------------ |
| `A`-`G` (+ `#`/`+` sustenido ou `-` bemol) | nota, com duração numérica opcional (`C4`, `C8.` com ponto de aumento) |
| `O<n>` | define a oitava (2..6) |
| `<` / `>` | oitava abaixo / acima |
| `L<n>` | duração padrão das notas seguintes |
| `V<n>` | volume padrão (0..15) |
| `R[n]` | pausa |

Um caractere ou combinação não reconhecida é um **erro claro de
compilação**, apontando a posição no texto. Só o canal A é suportado
(`PLAY "a","b","c"` multi-canal e `T` de tempo estão fora de escopo por
enquanto).

## 12. Arquivos

```basic
OPEN "SCORE.TXT" FOR OUTPUT AS #1
PRINT #1, "texto", numero
CLOSE #1
```

- `#n` precisa ser um **literal inteiro em tempo de compilação**, não uma
  expressão — cada `#n` vira um byte global dedicado (`DGN_FileHandle_<n>`).
- `FOR OUTPUT` sempre cria/trunca o arquivo do zero. `FOR APPEND` (desde a
  v4.8.0) abre o arquivo existente sem truncar (criando-o do zero só se
  ainda não existir) e posiciona o ponteiro no fim antes de qualquer
  escrita — duas aberturas em `APPEND` seguidas acrescentam, não
  sobrescrevem uma a outra.
- `FOR INPUT` abre o arquivo para leitura, mas **ainda não existe
  `INPUT #n`/`LINE INPUT #n`** para ler o conteúdo de volta — só dá pra
  abrir e fechar o handle por enquanto. Ler arquivo de dentro do BASIC é o
  próximo passo natural agora que STRING existe de verdade.

## 13. O compilador `DIGNAC` (linha de comando)

```bash
dignac [opções] <arquivo.bas>
```

| Opção          | Descrição                                                                 |
| -------------- | ---------------------------------------------------------------------------- |
| `-o <caminho>` | Arquivo `.mob` de saída (padrão: mesmo nome com extensão `.mob`).            |
| `-S`           | Emite o Assembly Z80 gerado (`.asm`) em vez do `.mob` — ótimo pra aprender/depurar. |
| `-api <f|dir>` | Descritores de API da MSXLIB (`.api`, arquivo ou diretório; repetível). Sem `-api`, usa `../lib/api` ao lado do executável. |
| `-v`           | Modo detalhado.                                                              |
| `--version`    | Versão atual.                                                                |
| `-h`, `--help` | Ajuda completa.                                                              |

```bash
# 1. Compila BASIC Dignified para objeto relocável .mob
dignac sample/basic/hello.bas -o sample/basic/hello.mob

# 2. Linka com a biblioteca padrão gerando o executável .COM
musubi -v -o sample/basic/hello.com sample/basic/hello.mob lib/msxlib.hlib
```

### Chamando rotinas da MSXLIB (`.api`)

Qualquer rotina da MSXLIB descrita num `.api` (`lib/api/`) pode ser chamada como uma
`PROCEDURE`/função qualquer, sem `EXTERN` e sem comando dedicado no compilador:

```basic
PROCEDURE Main()
    PSG_Write(7, 62)              ' proc: comando
    VDP_SetColor(15, 4)           ' texto branco sobre fundo azul
    PrintDec16(Mul16(6, 7))       ' func: usada como expressão
END PROCEDURE
```

Número de argumentos errado, ou usar uma `proc` numa expressão, é erro de compilação.
Uma `PROCEDURE` do próprio programa com o mesmo nome vence a rotina da biblioteca.
Formato do `.api` e regras: `docs/manual-ferramentas.md` §6.3.

## 14. Ver também

- `docs/manual-ferramentas.md` — `.MOB`, `MUSUBI`, bank switching,
  `MSXLIB`, `OBI`, e um exemplo real combinando DIGNAC com as outras
  linguagens.
- `docs/manual-assembly.md` — para quando uma rotina precisa descer a nível
  de hardware (a `MSXLIB` que o DIGNAC chama por baixo é toda escrita nele).

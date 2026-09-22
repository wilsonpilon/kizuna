# KIZUNA — Manual da Linguagem Pascal (WIRTH80)

> Este manual descreve o subconjunto de Pascal **realmente implementado**
> pelo compilador `WIRTH80` hoje. `SPEC.md` registra a visão original (um
> Pascal fiel ao Turbo Pascal 4, com units e compilação separada) — este
> documento descreve o que compila **de fato**, para não prometer o que
> ainda não existe. Para o formato de objeto, o linker e bank switching,
> veja `docs/manual-ferramentas.md`.

## 1. Escopo atual vs. visão do projeto

`WIRTH80` compila um **programa Pascal único por arquivo** — sem units, sem
`uses`, sem arrays/records, sem heap (`New`/`Dispose`). Desde a v4.9.0, um
programa pode declarar suas próprias `procedure`/`function` (§4) e exportar
ou importar símbolos via `PUBLIC`/`EXTERN` (§5) — antes disso, todo programa
só virava seu próprio `Start` isolado, sem nenhuma sub-rotina própria. A
visão mais ampla de `SPEC.md` (units com `interface`/`implementation`,
`{$USES}` cruzando módulos noutras linguagens) ainda **não está
implementada** — o arquivo `demo/main.pas` no repositório é um "croqui
hipotético" explicitamente marcado como não compilável, mostrando pra onde o
projeto está indo, não o que já funciona.

Na prática, isso significa que um módulo Pascal **já pode**, hoje, exportar
uma `procedure`/`function` via `PUBLIC` pra outro módulo `KAJI80`/`DIGNAC`
chamar, ou declarar `EXTERN` pra chamar uma rotina de outro módulo — **desde
que essa rotina siga a mesma ABI Kizuna baseada em pilha** (`PUSH`
esquerda→direita, `IX` como frame pointer, limpeza pelo chamador — ver
`docs/manual-ferramentas.md` §"ABI"). Rotinas da `MSXLIB` com convenção de
registrador própria (`VDP_PSet`, `BDOS_PrintChar` etc.) **não** são
chamáveis assim — só por comandos dedicados como `Write`/`WriteLn`, do mesmo
jeito que o `DIGNAC` só acessa essas rotinas via `PSET`/`LINE`, nunca por
uma chamada genérica. Ver `docs/manual-ferramentas.md` §10 pra um exemplo
prático do que isso já destrava.

## 2. Estrutura de um programa

```pascal
program NomeDoPrograma;

var
  a, b: Integer;

begin
  a := 10;
  b := a + 5;
  WriteLn('Resultado: ', b);
end.
```

Ordem completa (todas as partes depois de `program Nome;` são opcionais,
mas quando presentes seguem sempre esta ordem): `PUBLIC`/`EXTERN` → `var` →
`procedure`/`function` → bloco principal `begin...end.`.

## 3. Declaração de variáveis

```pascal
var
  contador: Integer;
  letra: Char;
  ok: Boolean;
  nome: String;
```

| Tipo palavra-chave | Armazenamento real hoje |
| -------------------- | -------------------------- |
| `Integer`            | palavra de 16 bits          |
| `Char`               | 1 byte                      |
| `Boolean`             | 1 byte                      |
| `String`              | **atualmente idêntico a `Integer`** — palavra de 16 bits, sem nenhuma semântica de string |

`Char`/`Boolean` reservam 1 byte de verdade, mas são impressos por `Write`/
`WriteLn` como **número decimal** (código ASCII/0-ou-1), não como caractere
— não existe ainda uma rotina de impressão específica para eles. `String`
aceita a palavra-chave no `var`, mas o compilador não trata esse tipo de
forma diferente de `Integer` — atribuir um literal de string a uma variável
`String` (`nome := 'oi'`) é hoje um **erro de compilação** ("expressão não
suportada"), porque literais de string só são aceitos como argumento direto
de `Write`/`WriteLn`, nunca como expressão geral. Trate `String` como
reservado-mas-não-funcional por enquanto.

## 4. Procedimentos e funções

```pascal
program Exemplo;

var
  resultado: Integer;

function Dobro(x: Integer): Integer;
begin
  Dobro := x * 2;   { TP4 clássico: atribuir ao próprio nome devolve o valor }
end;

procedure Saudacao(nome: Integer);
begin
  Write('Ola, sujeito numero ');
  WriteLn(nome);
end;

begin
  Saudacao(7);
  resultado := Dobro(21) + 1;
  WriteLn(resultado);
end.
```

- Declaradas depois do `var` do programa, antes do `begin...end.` principal
  — posição clássica de Pascal. Cada uma pode ter seu próprio `var` local.
- **Declare-antes-de-usar**: uma só pode chamar outra já declarada antes
  dela no arquivo — sem forward declarations por enquanto.
- Lista de parâmetros no formato clássico, grupos separados por `;` e nomes
  dentro de um grupo por `,`: `procedure P(a, b: Integer; c: Char);`. Todo
  parâmetro ocupa um slot de 2 bytes na pilha independente do tipo
  declarado (`PUSH` no Z80 só move pares de registrador) — só a leitura de
  volta sabe que `Char`/`Boolean` usa apenas o byte baixo.
- Chamar uma `procedure` como comando **exige parênteses mesmo sem
  argumentos** (`Foo();`, não `Foo;` como o Pascal clássico permitiria) —
  simplificação desta primeira leva.
- `function` devolve valor por **atribuição ao próprio nome** dentro do
  corpo (`Dobro := expr;`), estilo Turbo Pascal — não por um `return`
  estilo C/BASIC. Uma `function` sem nenhuma atribuição ao próprio nome
  devolve lixo em `HL` (mesmo comportamento indefinido do Pascal clássico
  nesse caso — não é erro de compilação).
- Usar uma `function` dentro de uma expressão maior (`Dobro(21) + 1`)
  funciona — o valor de retorno em `HL` é preservado corretamente mesmo com
  a limpeza da pilha depois do `CALL` (isso tinha um bug real no `DIGNAC`
  equivalente, corrigido na mesma sessão que trouxe isso pro `WIRTH80` — ver
  `CHANGELOG.md`).

## 5. PUBLIC e EXTERN

```pascal
program Lib;
PUBLIC Foo, Bar;
EXTERN OutroSimbolo;

procedure Foo(a: Integer);
begin
  WriteLn(a);
end;
```

- Mesma palavra-chave nua do `KAJI80`/`DIGNAC` (não a diretiva `{$...}` do
  Turbo Pascal real) — decisão deliberada, prioriza um modelo mental só
  pra toolchain inteira. Vêm logo após `program Nome;`, antes do `var`.
- Lista separada por vírgula, terminada em `;` (Pascal usa `;` pra terminar
  declarações, diferente do `KAJI80`/`DIGNAC`).
- Só fazem sentido pra nomes de `procedure`/`function` — variável não é
  chamável entre módulos.
- `Start` (o bloco principal `begin...end.`) é **sempre** `PUBLIC`,
  automaticamente, mesmo sem nada declarado — continua sendo o ponto de
  entrada padrão de um `.COM` standalone.
- **Limitação real**: `EXTERN` + chamada genérica só funciona pra símbolos
  que seguem a ABI Kizuna de pilha (outra `PROCEDURE` `DIGNAC`/`KAJI80`, ou
  uma `procedure` `WIRTH80` `PUBLIC`) — não serve pra chamar rotinas da
  `MSXLIB` diretamente (elas usam convenção de registrador, não pilha). Ver
  §1.

## 6. Comandos suportados

```pascal
' Atribuição
identificador := expressao;

' Impressão (aceita literais de string e expressões inteiras)
Write(expressao1, expressao2, ...);
WriteLn(expressao1, ...);
WriteLn;                          ' só a quebra de linha

' Condicional
if condicao then comando [else comando]

' Laço
while condicao do comando

' Bloco
begin
  comando1;
  comando2;
end;
```

**Não implementado ainda** (mesmo com algumas palavras-chave já reservadas
no léxico): laço `for...to/downto...do` (os tokens `for`, `to`, `downto`
existem no lexer mas o parser não tem caso nenhum pra eles — usar `for`
resulta em erro de sintaxe), `repeat...until`, `case`, `uses`/units, arrays,
records, `New`/`Dispose` (heap), `ReadLn`/`Read`/`ReadKey` do teclado,
recursão indireta e forward declarations de `procedure`/`function`.

## 7. Expressões e operadores

| Categoria   | Operadores                                    |
| ------------ | ------------------------------------------------ |
| Aritméticos  | `+`, `-`, `*` (via `Mul16`), `div` (via `Div16`)  |
| Relacionais  | `=`, `<>`, `<`, `<=`, `>`, `>=`                    |
| Unário        | `+`, `-`                                          |

Não existe `and`/`or`/`not` lógico ainda (apenas os operadores acima).

## 8. Literais

```pascal
123        ' inteiro decimal
$FF        ' inteiro hexadecimal
'texto'    ' string literal (só como argumento direto de Write/WriteLn)
'a''b'     ' aspa simples escapada duplicando-a -> "a'b"
```

## 9. Comentários

```pascal
{ bloco de chaves }
(* bloco de parênteses-asterisco *)
// até o fim da linha
```

## 10. O compilador `WIRTH80` (linha de comando)

```bash
wirth80 [opções] <arquivo.pas>
```

| Opção          | Descrição                                                                 |
| -------------- | ---------------------------------------------------------------------------- |
| `-o <caminho>` | Arquivo `.mob` de saída (padrão: mesmo nome com extensão `.mob`).            |
| `-S`           | Emite o Assembly Z80 gerado (`.asm`) em vez do `.mob`.                       |
| `-v`           | Modo detalhado.                                                              |
| `--version`    | Versão atual.                                                                |
| `-h`, `--help` | Ajuda completa.                                                              |

```bash
# 1. Compilar o Pascal gerando o módulo objeto .mob
wirth80 -v sample/pascal/hello.pas -o sample/pascal/hello.mob

# 2. Linkar com a biblioteca padrão gerando o executável .COM
musubi -v -m sample/pascal/hello.map -o sample/pascal/hello.com \
  sample/pascal/hello.mob lib/msxlib.hlib
```

## 11. Roteiro (o que falta pra chegar na visão de `SPEC.md`)

Na ordem que mais desbloqueia o resto:

1. `uses`/`{$USES}` cruzando módulos e linguagens (units de verdade), como
   o `demo/` já esboça — `procedure`/`function`/`PUBLIC`/`EXTERN` dentro de
   um único arquivo (§4-5) são o pré-requisito que já está pronto.
2. `for...to/downto...do` (os tokens já existem, só falta o parser).
3. Diferenciação de tipo de verdade para `Char`/`Boolean`/`String` (mesmo
   trabalho que o DIGNAC já fez — ver `docs/manual-basic-dignified.md` §3 —
   inclusive a decisão já tomada de usar **IEEE 754** para um futuro tipo
   `Real`, para reaproveitar a mesma engine de ponto flutuante que o
   DIGNAC vai precisar, em vez de duas implementações separadas).
4. Chamar `procedure`/`function` sem parênteses quando não há argumentos
   (`Foo;` em vez de `Foo();`) — só uma simplificação de sintaxe, não um
   bloqueio real.
5. Forward declarations e recursão indireta.

## 12. Ver também

- `docs/manual-ferramentas.md` — `.MOB`, `MUSUBI`, bank switching,
  `MSXLIB`, `OBI`.
- `docs/manual-basic-dignified.md` — DIGNAC já passou por essa mesma
  jornada de "tipo real" recentemente; vale de referência para o que o
  Pascal ainda precisa.

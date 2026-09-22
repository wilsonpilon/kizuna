# KIZUNA — Manual da Linguagem Pascal (WIRTH80)

> Este manual descreve o subconjunto de Pascal **realmente implementado**
> pelo compilador `WIRTH80` hoje. `SPEC.md` registra a visão original (um
> Pascal fiel ao Turbo Pascal 4, com units e compilação separada) — este
> documento descreve o que compila **de fato**, para não prometer o que
> ainda não existe. Para o formato de objeto, o linker e bank switching,
> veja `docs/manual-ferramentas.md`.

## 1. Escopo atual vs. visão do projeto

`WIRTH80` hoje compila um **programa Pascal único e autocontido** — sem
units, sem `uses`, sem procedimentos/funções definidos pelo usuário, sem
arrays/records, sem heap (`New`/`Dispose`). Todo programa vira diretamente
seu próprio ponto de entrada `Start`, pronto pra virar um `.COM` sozinho
(linkando com `msxlib.hlib`). A visão de `SPEC.md` (units com
`interface`/`implementation`, `{$USES}` cruzando módulos noutras
linguagens) ainda **não está implementada** — o arquivo `demo/main.pas` no
repositório é um "croqui hipotético" explicitamente marcado como não
compilável, mostrando pra onde o projeto está indo, não o que já funciona.

Isso significa, na prática, que **hoje não é possível** um programa Pascal
chamar uma rotina Assembly ou DIGNAC (ou vice-versa) — só KAJI80 e DIGNAC
têm `PUBLIC`/`EXTERN` e podem ser linkados num único `.COM` multi-módulo. Um
programa WIRTH80 é sempre seu próprio executável completo. Ver
`docs/manual-ferramentas.md` §10 para como isso afeta um projeto que
combina as três linguagens.

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

Um único bloco `var` (opcional) seguido de um único bloco `begin...end.`.

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

## 4. Comandos suportados

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
resulta em erro de sintaxe), `repeat...until`, `case`, procedimentos e
funções definidos pelo usuário, `uses`/units, arrays, records, `New`/
`Dispose` (heap), `ReadLn`/`Read`/`ReadKey` do teclado.

## 5. Expressões e operadores

| Categoria   | Operadores                                    |
| ------------ | ------------------------------------------------ |
| Aritméticos  | `+`, `-`, `*` (via `Mul16`), `div` (via `Div16`)  |
| Relacionais  | `=`, `<>`, `<`, `<=`, `>`, `>=`                    |
| Unário        | `+`, `-`                                          |

Não existe `and`/`or`/`not` lógico ainda (apenas os operadores acima).

## 6. Literais

```pascal
123        ' inteiro decimal
$FF        ' inteiro hexadecimal
'texto'    ' string literal (só como argumento direto de Write/WriteLn)
'a''b'     ' aspa simples escapada duplicando-a -> "a'b"
```

## 7. Comentários

```pascal
{ bloco de chaves }
(* bloco de parênteses-asterisco *)
// até o fim da linha
```

## 8. O compilador `WIRTH80` (linha de comando)

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

## 9. Roteiro (o que falta pra chegar na visão de `SPEC.md`)

Na ordem que mais desbloqueia o resto:

1. Procedimentos e funções definidas pelo usuário (`procedure`/`function`)
   — pré-requisito de tudo que depende de reutilizar código dentro do
   próprio programa.
2. `PUBLIC`/`EXTERN` do lado do Pascal (hoje só o `Start` é exportado) —
   pré-requisito para um módulo Pascal poder ser chamado de fora ou chamar
   uma rotina Assembly/DIGNAC.
3. `uses`/`{$USES}` cruzando módulos e linguagens, como o `demo/` já
   esboça.
4. `for...to/downto...do` (os tokens já existem, só falta o parser).
5. Diferenciação de tipo de verdade para `Char`/`Boolean`/`String` (mesmo
   trabalho que o DIGNAC já fez — ver `docs/manual-basic-dignified.md` §3 —
   inclusive a decisão já tomada de usar **IEEE 754** para um futuro tipo
   `Real`, para reaproveitar a mesma engine de ponto flutuante que o
   DIGNAC vai precisar, em vez de duas implementações separadas).

## 10. Ver também

- `docs/manual-ferramentas.md` — `.MOB`, `MUSUBI`, bank switching,
  `MSXLIB`, `OBI`.
- `docs/manual-basic-dignified.md` — DIGNAC já passou por essa mesma
  jornada de "tipo real" recentemente; vale de referência para o que o
  Pascal ainda precisa.

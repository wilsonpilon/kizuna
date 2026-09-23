# Release Notes — KIZUNA v4.11.0 "Kakuchou" (拡張)

> "Um assembler bom não é o que copia outro de verdade — é o que sabe quando NÃO copiar."

**Kakuchou** (拡張) — "expansão, extensão". Depois de *Jisshou* (実証, a
aritmética `SINGLE` e a primeira ferramenta de execução real de Z80 do
projeto), esta release é inteiramente sobre o `KAJI80` amadurecendo: oito
recursos de macro-assembler inspirados no [asMSX](https://github.com/Fubukimaru/asMSX),
um cross-assembler Z80 consagrado na cena MSX, mas com sintaxe e escolhas
próprias do KIZUNA onde fazia sentido divergir — nunca uma cópia cega.

## KAJI80 ganha um motor de macro-assembler completo, em 8 fases

Wilson pediu pra estudar a documentação do asMSX e trazer recursos
equivalentes pro `KAJI80`. O plano de 8 fases foi desenhado e aprovado
de uma vez antes de qualquer código, e entregue fase por fase, cada uma
verificada e commitada antes da próxima:

1. **Avaliador de expressões numéricas** em tempo de montagem —
   aritmética, bits, lógica, comparação, funções matemáticas (`SIN`,
   `COS`, `SQRT`, `POW`...), a constante `PI`, conversão de ponto fixo
   8.8 (`FIX`/`FIXMUL`/`FIXDIV`/`INT` — necessário porque o Z80 não tem
   ponto flutuante nativo) e `RANDOM(n)` (constante pseudoaleatória
   embutida em tempo de montagem). `EQU` passa a aceitar a expressão
   inteira; nova forma `Nome = expressão` pra variáveis reatribuíveis.
2. **Rótulos locais** (`.nome:`) — escopados ao rótulo global mais
   recente, sem custo em runtime, evitam ter que inventar um nome novo
   pra cada `.loop`/`.done` trivial de cada rotina.
3. **Montagem condicional** (`IF`/`ELSE`/`ENDIF`).
4. **Repetição de bloco** (`REPT n`/`ENDR`), com desambiguação
   automática de rótulo local entre iterações.
5. **Macros** (`MACRO @param.../ENDM`), incluindo substituição de
   parâmetro dentro de um identificador maior (`.reset_@VAR:`).
6. **Rótulos pré-definidos** de BIOS (~120), variáveis de sistema
   (~130) e uma tabela nova de códigos de função do MSX-DOS/MSX-DOS2 —
   sem precisar de `EXTERN` nem `INCLUDE`.
7. **`CALLBIOS`/`CALLDOS`** — chamada inter-slot de BIOS e chamada de
   MSX-DOS reduzidas a uma linha.
8. **`INCBIN "arquivo", SKIP=x, SIZE=y`** — inclusão direta de um
   arquivo binário no objeto.

## "Nossa própria sintaxe": divergências deliberadas do asMSX

Cada uma resolve um conflito real com algo que o `KAJI80` já tinha, não
capricho: `MOD` em vez de `%` pra módulo (`%` já é o prefixo de literal
binário, `%1010`), `@param` em vez de `#param` nas macros (`#` já é
prefixo de literal hexadecimal), `SKIP=x`/`SIZE=y` em vez de espaço no
`INCBIN` (evita uma ambiguidade real de como o assembler reconstrói um
operando internamente). A tabela de códigos `BDOS` é inteiramente nova —
o asMSX nem tem equivalente.

## Bugs reais encontrados no caminho

Dois, como efeito colateral de reproduzir exemplos reais da própria
documentação do asMSX em testes: `INC (HL)`/`DEC (HL)` — instrução Z80
padrão — nunca tinham sido implementados; e a separação de operandos por
vírgula não respeitava parênteses (`DB POW(2,3)` quebrava errado em dois
operandos).

Um terceiro, mais interessante, só apareceu **testando os recursos
juntos**, não isolados — exatamente o tipo de bug que testes por fase não
pegam. `EQU FIX(1.5)` (Fase 1) dentro de um arquivo com rótulos locais
(Fase 2) quebrava com "rótulo local usado antes de qualquer rótulo
global", mesmo sem nenhum rótulo local de verdade no arquivo. Causa raiz:
o lexer nunca soube reconhecer um ponto decimal dentro de um número —
`"1.5"` sempre virou dois tokens, já que `.` é caractere de identificador
válido desde sempre (usado pelos próprios rótulos locais). Inofensivo
antes da Fase 2 existir; um bug de verdade assim que ela passou a tratar
qualquer identificador com ponto na frente como referência de rótulo
local. Corrigido na raiz, no lexer.

## Três programas reais pra testar em hardware

`sample/macroasm/` — `expr_labels.asm`, `predefined.asm` e `incbin.asm`,
em `KAJI80` puro (sem `DIGNAC`/`WIRTH80`), cada um imprimindo
`[OK]`/`[FALHOU]` visível no console via BDOS. Verificados antes da
entrega num emulador Z80 (o mesmo harness deste projeto usado pro motor
de float): saída completa capturada e conferida pros dois primeiros;
`predefined.asm` (que troca `SCREEN 1`/`0` de verdade via `CALLBIOS
CHGMOD`) teve a chamada inter-slot confirmada até o ponto de entrada real
da ROM — a mudança de tela em si só é verificável em hardware/openMSX de
verdade, e é exatamente o que foi pedido a Wilson pra fechar o ciclo.

## Próximos passos

- Confirmação em hardware/openMSX dos três programas de
  `sample/macroasm/` (pendente).
- `parseImm8` continua sem forma de reportar erro, e os operandos de
  porta de `IN`/`OUT` e o número de bit de `BIT`/`RES`/`SET` passam por
  ela sem guarda contra a mesma classe de bug já fechada em outros
  pontos do `KAJI80` — risco prático baixo (portas de I/O no projeto
  inteiro sempre usam constantes `EQU`), documentado, não escondido.
- `INCLUDE` (inclusão de outro arquivo-fonte `.asm`) nunca foi pedido
  nem entrou no plano das 8 fases — candidato natural pra uma leva
  futura se fizer falta.

---

# Release Notes — KIZUNA v4.10.0 "Jisshou" (実証)

> "Não basta parecer certo no papel — só conta o que roda de verdade, no emulador e na placa."

**Jisshou** (実証) — "prova empírica, verificação pela prática". Depois de
*Yuugou* (融合, as três linguagens linkando juntas), esta release entrega
aritmética `SINGLE` de verdade no `DIGNAC` — e, ao testá-la em hardware,
um lembrete direto do porquê desse nome: um bug real que sobreviveu a
várias rodadas de verificação no papel só caiu depois que o projeto
ganhou sua primeira ferramenta de execução real de Z80.

## DIGNAC ganha aritmética SINGLE de verdade: soma, subtração e comparação

```basic
DIM a!, b!, soma!
a! = 1.0 : b! = 0.5
soma! = a! + b!        ' Float_Add32
IF soma! > a! THEN      ' Float_Cmp32
    PRINT "maior"
END IF
```

`x! = a! + b!`, `x! = a! - b!` e comparação (`=`,`<>`,`<`,`<=`,`>`,`>=`)
agora funcionam para `SINGLE` (IEEE 754 binary32), via um motor de
software novo em `lib/src/float.asm` — `Float_Add32`/`Float_Sub32`/
`Float_Cmp32`, três novas rotinas na `MSXLIB`.

**Escopo desta leva, deliberado**: só operandos simples — `x! = a! + b!`
funciona, `x! = (a!+b!)*c!` continua um erro de compilação claro (sem
alocação de temporários ainda). `*`/`/` e `DOUBLE` ficam pra uma leva
futura: a normalização de mantissa 24×24→48 bits do produto/quociente se
mostrou bem mais delicada do que soma/subtração — decidido entregar
Add/Sub/Cmp **verificados** agora em vez de arriscar `*`/`/` malfeitos.
`PRINT` de float continua fora de escopo (conversão decimal é o item mais
difícil de todos).

**Metodologia**: o algoritmo (unpack/align/normalize/repack) foi escrito
primeiro num protótipo em Go, testado contra ~200 mil pares aleatórios
comparados com a aritmética `float32` nativa do Go — achou e corrigiu 2
bugs de normalização antes de qualquer linha de Z80 ser escrita. Na
transliteração pro Z80, mais 3 bugs reais: `KAJI80` não suporta
aritmética `Label+N` em operandos (virava `EXTERN` fantasma silencioso),
`LD DE,(nn)` não existe no Z80 de verdade, e comparar o expoente (signed)
com `CP` direto é uma comparação *unsigned* e errava sempre que os dois
expoentes tinham sinais diferentes — só descoberto ao escolher
deliberadamente um exemplo pra hand-trace que exercitasse esse caminho.

## Bug real que só apareceu em hardware — e a primeira ferramenta de execução real de Z80 do projeto

Wilson testou `sample/basic/floatmath.bas` em hardware/openMSX: soma
falhava, uma das duas comparações falhava, subtração e a outra
comparação funcionavam. Um padrão que resistiu a várias re-verificações
por hand-tracing — não existia emulador Z80 no repositório até então, só
montagem, leitura de código e raciocínio manual.

A causa raiz só foi encontrada depois de montar um harness em Go usando
`github.com/remogatto/z80` (scratchpad, fora do repositório) pra
**executar de verdade** o `.com` linkado, com registradores controlados —
primeira vez que este projeto teve acesso a execução real de Z80 pra
depuração. Em minutos: `Float_Cmp32` usava `LD B, (Float_MantHi)` — uma
forma que **não existe** no Z80 de verdade (só `LD A,(nn)` tem
endereçamento absoluto de 16 bits pra um registrador de 8 bits) — e o
`KAJI80` montava isso em silêncio como `LD B, 0`. Só quebrava quando o
byte mais significativo (sinal + topo do expoente) de A e B já eram
iguais — ou seja, **qualquer comparação de igualdade**, ou dois positivos
com faixa de expoente parecida, exatamente o padrão observado.

Corrigido e **reconfirmado em hardware pelo usuário** no mesmo dia.
Verificado também com 20.000 pares aleatórios rodados no emulador contra
a aritmética `float32` nativa do Go: `Cmp32` foi de "falhava em quase
todo caso de byte-alto igual" pra 0 falhas em 20.000.

## Auditoria do KAJI80: uma classe inteira de bug fechada de uma vez

A pedido de Wilson, logo depois do bug acima: existem outros pontos no
assembler com a mesma forma — um operando que deveria ser rejeitado
caindo em silêncio num fallback de imediato/símbolo em vez de dar erro?

Achados e corrigidos dois:

1. `emitAddressOrReloc`/`emitRelativeOrReloc` (usadas por `CALL`, `JP`,
   `JR`, `DJNZ`, `LD (nn),A/HL`, `LD HL,(nn)`, `LD rr,nn`, `LD IX/IY,nn`,
   `DW`) não validavam o nome do símbolo — qualquer string virava uma
   "relocation" válida em silêncio. Mesma causa raiz do bug de `Label+N`
   já conhecido, agora fechada na fonte, não só contornada num arquivo.
   De brinde, fecha também `LD DE,(nn)`/`LD BC,(nn)`.
2. `encodeAlu8` (`ADD`/`ADC`/`SUB`/`SBC`/`AND`/`XOR`/`OR`/`CP`) ganhou o
   mesmo tipo de guarda pra endereço absoluto entre parênteses — `CP
   (Algo)` virava `CP 0` em silêncio.

Rebuild completo de `MSXLIB` e todos os exemplos `BASIC`/`Pascal`/`OBI`
confirma: pura proteção, nada legítimo dependia do comportamento antigo.

## WIRTH80 ganha a diretiva BANK <n>

Item já sinalizado como "próximo passo" na v4.9.0. Programas Pascal
(`WIRTH80`) agora podem declarar `BANK <n>;` logo após `program Nome;`
(mesma posição do `KAJI80`/`DIGNAC`, terminada com `;` pra combinar com a
sintaxe `PUBLIC`/`EXTERN` já estabelecida). `sample/obi/` — as 3
linguagens linkadas num único `.COM` — agora tem cada linguagem no seu
próprio banco de verdade, não mais duas forçadas a dividir o banco 0.

## Próximos passos

- `Float_Mul32`/`Float_Div32` — deliberadamente deferidos nesta leva.
- Expressões `SINGLE` aninhadas (`(a!+b!)*c!`) — precisa de alocação de
  temporários.
- `DOUBLE` com aritmética de verdade (mesmo algoritmo, mantissa de 52
  bits).
- `PRINT`/conversão decimal de `SINGLE`/`DOUBLE`.
- `uses`/units de verdade cruzando arquivos no `WIRTH80` (`{$USES}`).
- `for...to/downto...do` no `WIRTH80` (tokens já reservados, sem parser).

---

# Release Notes — KIZUNA v4.9.0 "Yuugou" (融合)

> "O laço (絆) finalmente amarra as três: Assembly, BASIC e Pascal, linkados de verdade num único .COM, testado em hardware real."

**Yuugou** (融合) — "fusão, integração, as partes se tornando um todo".
Depois de *Minori* (実り, a colheita — tipos reais no DIGNAC), esta release
entrega o que o próprio nome do projeto promete: **as três linguagens de
entrada do KIZUNA — Assembly, MSX-BASIC Dignified e Pascal — linkando
juntas num único `.COM`, executando de verdade em hardware real.** Não é
mais teoria nem só teste unitário isolado.

## WIRTH80 ganha procedimentos, funções, PUBLIC e EXTERN

Até aqui, todo programa `WIRTH80` era seu próprio `Start` autocontido —
sem `procedure`/`function` definidas pelo usuário, sem nada pra exportar
ou importar. Agora:

- **`procedure`/`function`** na posição clássica de Pascal, reaproveitando
  a mesma ABI de pilha já provada pelo `KAJI80`/`DIGNAC` (parâmetros
  empilhados esquerda→direita, frame pointer `IX`). `function` devolve
  valor por atribuição ao próprio nome (`Dobro := x * 2;`), estilo Turbo
  Pascal clássico.
- **`PUBLIC`/`EXTERN`** com a mesma palavra-chave nua do resto da
  toolchain — decisão deliberada, prioriza um modelo mental só.
- **Um único ponto de entrada por executável, garantido**: `Start` do
  `WIRTH80` agora só é gerado se o programa tiver de fato um bloco
  principal (mesma regra que o `DIGNAC` já usa pra `PROCEDURE Main`), e
  `MUSUBI` recusa a linkagem com um erro claro ("múltiplos pontos de
  entrada...") se mais de um módulo tentar assumir esse papel — pedido
  explícito de Wilson: "prefiro que exista apenas um único main".

`sample/obi/` — já o exemplo real de `KAJI80`+`DIGNAC` linkados num único
`.COM` multi-banco — ganha um terceiro módulo, `greet_lib.pas`
(`WIRTH80`, biblioteca pura, sem `Start` próprio). `main.asm` (dono do
`Start`) agora chama tanto uma rotina `DIGNAC` (cross-bank, via
trampolim) quanto uma rotina `WIRTH80` (mesmo banco, `CALL` direto). Não
é mais só uma prova de conceito.

**De brinde**: achado e corrigido um bug real ao escrever o `CallExpr` do
`WIRTH80` — o mesmo padrão já existia no `DIGNAC` com um bug nunca pego
(corrompia o valor de retorno de uma função usada dentro de uma expressão
maior, nunca exercitado por nenhum teste antes). Corrigido nos dois.

## Bug real corrigido: tela preta ao desenhar em SCREEN 2 a partir de um banco paginável

Wilson testou o `sample/obi/main.com` de 3 linguagens em hardware e
reportou: banner e `SCREEN 2` apareciam, mas o gráfico nunca desenhava —
tela preta, depois voltava pra tela de texto e encerrava normalmente.

Isolado o problema reconstruindo a versão anterior (só 2 linguagens, sem
o módulo `WIRTH80` novo): **o mesmo bug já acontecia** — confirmando que
era um bug pré-existente, não algo introduzido nesta sessão.

**Causa raiz**: o módulo `DIGNAC` que desenha o gráfico (banco 2) chama
de volta rotinas da `MSXLIB` que vivem no banco comum (`VDP_PSet`,
`VDP_BoxFill`, `VDP_Line`). O `MUSUBI` gerava um trampolim de troca de
banco pra cada uma dessas chamadas — mesmo o alvo sendo o banco comum,
que está **sempre** presente na memória, não importa o que estiver
mapeado na Página 2. Esse trampolim lia uma tabela de bancos que o
bootstrap nunca inicializava pra essa entrada específica (o comentário no
próprio código já admitia a entrada como "desperdício", mas a suposição
de que nunca seria lida estava errada) — resultado, cada chamada de
desenho trocava a Página 2 pra um segmento físico aleatório antes de
rodar.

Esse caminho específico (um banco paginável chamando de volta pro banco
comum) nunca tinha sido exercitado por nenhum teste confirmado em
hardware antes — o bug do bootstrap multi-banco corrigido no dia anterior
foi validado com um exemplo que só chama a BDOS diretamente, nunca uma
rotina `MSXLIB` de volta. Corrigido: uma chamada mirando o banco comum
agora sempre vira `CALL` direto, não importa de onde vem. **Confirmado
corrigido em hardware pelo usuário** — primeira vez na história do
projeto que esse caminho específico funciona de verdade.

## Dois bugs pequenos do DIGNAC, prometidos na v4.8.0, corrigidos

`FOR...STEP` negativo (o teste de término assumia sempre passo positivo)
e `OPEN...FOR APPEND` (se comportava igual a `FOR OUTPUT`, truncando em
vez de acrescentar).

## Próximos passos

- `WIRTH80` ainda não suporta a diretiva `BANK <n>` — todo módulo seu cai
  sempre no banco comum. Não bloqueia a integração das 3 linguagens (que
  já funciona), só limita onde um módulo `WIRTH80` pode morar num projeto
  multi-banco.
- `uses`/units de verdade cruzando arquivos (`{$USES}`) — o mecanismo de
  `PUBLIC`/`EXTERN` dentro de um único arquivo já está pronto como
  pré-requisito.
- `for...to/downto...do` no `WIRTH80` (tokens já reservados, sem parser).

---

# Release Notes — KIZUNA v4.8.0 "Minori" (実り)

> "A flor de Kaika amadurece: variáveis de verdade, e uma documentação honesta sobre o que ainda não é fruto."

**Minori** (実り) — "colheita, fruição, o fruto que amadurece depois da flor".
Depois de *Kaika* (開花, o florescimento — sprites, música e arquivo chegando
na `MSXLIB`), esta release colhe o que sustenta tudo: o `DIGNAC` ganha um
sistema de tipos de verdade (STRING, INTEGER, SINGLE, DOUBLE), e a
documentação do projeto inteiro é reorganizada em manuais — inclusive
documentando, pela primeira vez de forma explícita, o quanto o `WIRTH80`
ainda está longe da visão original do `SPEC.md`.

## DIGNAC: STRING, INTEGER, SINGLE, DOUBLE — um sistema de tipos de verdade

Até aqui, toda variável não-BOOLEAN no `DIGNAC` era tratada como uma
palavra de 16 bits, sempre — `LOCAL a%, b$, c!` dava o mesmo tipo pra todos
(o primeiro sufixo "vencia"), e só literais de string funcionavam de
verdade. Agora cada variável resolve seu próprio tipo pelo sufixo
(`%`/`$`/`!`/`#`), e **STRING funciona de ponta a ponta**: declarar,
atribuir (literal ou outra variável STRING) e imprimir, num buffer de 256
bytes (1 byte de tamanho + até 255 de dados — o formato "short string" já
usado nas fronteiras entre as três linguagens do projeto).

SINGLE e DOUBLE usam **IEEE 754** (binary32/binary64), não o MBF do
MSX-BASIC real — decisão deliberada, já que o `DIGNAC` compila pra código
nativo standalone. São declaráveis, aceitam literais (`3.14`, `1.5e+10`,
`2.71828d0`, sufixos `!`/`#`) e são atribuíveis por cópia de bytes — mas
**qualquer aritmética ou `PRINT` de SINGLE/DOUBLE é um erro de compilação
claro**, nunca um resultado errado calado. A engine de ponto flutuante
completa fica para uma sessão futura dedicada — sozinha, é do tamanho de
uma biblioteca de ponto flutuante em Z80 escrita do zero.

Literais `&O` (octal) e `&B` (binário) também chegam, no mesmo padrão do
`&H` já existente.

**Confirmado em hardware real pelo usuário**: `sample/basic/types.bas`
imprimiu as duas strings e os dois valores inteiros corretamente, sem lixo
na tela.

## Documentação reorganizada em manuais por assunto

`HELP.md` tinha virado um arquivo único e desatualizado (faltava `OBI`,
faltavam as novidades recentes do `DIGNAC`). Vira quatro manuais novos em
`docs/`, cada um cobrindo seu próprio assunto e atualizado contra o estado
real do código:

- **`docs/manual-assembly.md`** — `KAJI80`: diretivas, instruções Z80
  suportadas, e as formas explicitamente **não** suportadas (erro claro,
  nunca silencioso).
- **`docs/manual-basic-dignified.md`** — `DIGNAC`: o sistema de tipos
  novo, sprites, música (MML), arquivos, e as limitações conhecidas (como
  `FOR` com `STEP` negativo, que ainda não funciona corretamente).
- **`docs/manual-pascal.md`** — `WIRTH80`: o escopo **real** hoje, que é
  bem menor que a visão original do `SPEC.md` sugere.
- **`docs/manual-ferramentas.md`** — formato `.MOB`/`.MAP`, `MUSUBI`, um
  capítulo dedicado de bank switching, `HAKO`/`.HLIB`, `MSXLIB`, `OBI`, e
  um exemplo end-to-end honesto sobre misturar as três linguagens.

**Achado real durante a pesquisa**: `WIRTH80` hoje não tem procedimentos,
funções, `uses`/units, nem `PUBLIC`/`EXTERN` do lado do programador — só
compila um programa único e autocontido. O laço `for` tem os tokens
reservados no lexer mas nenhum caso no parser (é erro de sintaxe hoje). O
tipo `String` em `var` é aceito mas tratado internamente como `Integer`,
sem nenhuma semântica de string de verdade. Isso significa que, hoje, só
`KAJI80`+`DIGNAC` conseguem formar um único `.COM` multi-banco de verdade
(`sample/obi/Obifile` já faz isso) — `WIRTH80` ainda não pode ser um
terceiro módulo nesse mesmo binário. `demo/main.pas`/`demo/Obifile`
continuam sendo, como sempre foram, um croqui hipotético — agora
documentado de forma explícita, não só implícita.

## Próximos passos

- Corrigir os bugs pequenos que a documentação acabou de expor: `FOR` com
  `STEP` negativo (o teste de término do laço assume passo positivo),
  `OPEN ... FOR APPEND` que hoje se comporta igual a `OUTPUT` (não faz
  seek até o fim do arquivo).
- Dar ao `WIRTH80` `PUBLIC`/`EXTERN` e procedimentos/funções definidos
  pelo usuário — o item mais estrutural, e o que destrava o exemplo real
  das três linguagens num único `.COM`.

---

# Release Notes — KIZUNA v4.7.0 "Kaika" (開花)

> "Depois de fechar o roadmap, o laço floresce: sprites, música e arquivos, provados um por um em hardware real."

**Kaika** (開花) — "florescimento, o momento em que a flor desabrocha". Depois
de *Kansei* (完成, a conclusão do roadmap original), esta release é o
primeiro capítulo do que vem depois: a `MSXLIB` deixa de ser fina e ganha
as três capacidades de maior valor prático para quem for escrever um
programa de verdade — **sprites**, **música PSG** e **I/O de arquivo** —
todas verificadas contra fontes primárias antes de escrever qualquer
código, e todas testadas em hardware real pelo usuário no mesmo dia.

## MSXLIB: sprites, música e arquivos

- **Sprites** (`lib/src/vdp.asm`): `VDP_SpriteDefine`, `VDP_SpriteSet`,
  `VDP_SpriteHide`, `VDP_SpriteHideAll`, `VDP_SpriteSetSize`. Endereços de
  tabela confirmados contra `resource/MSXgl/engine/src/vdp_reg.h` — os
  mesmos que o MSX-BASIC já usa por padrão em SCREEN 2.
- **Música PSG** (`lib/src/psg.asm`): `PSG_NoteTable` (5 oitavas, 60
  períodos calculados pela fórmula padrão, conferidos contra o valor
  público A4=440Hz→254), `PSG_PlayNoteIndexed`, `PSG_PlaySequence`.
- **I/O de arquivo** (`lib/src/bdos.asm`): `BDOS_FileOpen/Create/Close/
  Read/Write/Seek` — funções de MSX-DOS 2 baseadas em handle, números
  confirmados contra o protocolo oficial (diferentes dos genéricos de
  CP/M/MS-DOS que se assumiria por padrão).
- Três exemplos novos, um por área: `sample/sprites/`, `sample/music/`,
  `sample/fileio/` — todos em `KAJI80` puro (nem `WIRTH80` nem `DIGNAC`
  ainda sabem chamar uma rotina externa por convenção de registrador
  arbitrária; dar essa capacidade aos compiladores é o próximo passo
  natural depois desta release).

**Testado em hardware real**: sprites e música funcionaram corretamente
de primeira. I/O de arquivo criava e escrevia o arquivo certo, mas a
leitura de volta imprimia caracteres bagunçados na tela.

## Bug real do KAJI80 encontrado e corrigido: literal de caractere virava zero em silêncio

O sample de I/O de arquivo usava `LD (HL), '$'` para marcar onde parar de
imprimir a string lida de volta do arquivo (a função 09h da BDOS exige um
terminador `'$'`, que o conteúdo do arquivo não tem). O assembler
reconstruía os operandos de uma instrução perdendo as aspas de um literal
de caractere — `'$'` virava só `$`, indistinguível de um prefixo
hexadecimal malformado, e o parser de imediatos falhava em silêncio,
devolvendo `0`. `LD (HL),'$'` virava `LD (HL),0x00` em vez de
`LD (HL),0x24`, **sem nenhum erro de montagem** — a mesma classe de bug já
vista com `(IX+d)` na saga da SCREEN 2: sintaxe não reconhecida virando 0
em silêncio ao invés de erro. Com um byte nulo no lugar do terminador
real, a rotina de impressão nunca parava e mostrava memória adiante
indefinidamente.

Corrigido na raiz do assembler (`pkg/kaji80/assembler.go`), não só
contornado no exemplo: o parser agora preserva as aspas de um literal de
caractere ao reconstruir o operando. Novo teste de regressão
`TestCharLiteralImmediate`. **Confirmado pelo usuário em hardware real
após a correção.**

## Limpeza do disco de teste

`sample/` é usado como disco MSX-DOS 2 montado diretamente pelo openMSX
para testar os `.com` do projeto. A pedido do usuário, reduzido para caber
no limite de 720KB de um disquete DD padrão: `sample/screen2_test.*`
removido (saga da SCREEN 2 encerrada), `sample/UTILS/` reduzido de 21 para
5 arquivos essenciais, e `sample/HELP/` (433K de textos de ajuda do
MSX-DOS 2) removido por inteiro. `sample/` caiu de ~1.1M para 526K.

## Próximos passos

- Dar ao `DIGNAC` (e eventualmente ao `WIRTH80`) acesso às novas rotinas
  de sprite/música/arquivo, para que MSX-BASIC Dignified consiga usá-las
  sem cair para Assembly puro — o que realmente destrava escrever um jogo
  simples só em BASIC.
- Investigar o não-determinismo residual de `WIRTH80`/`DIGNAC` (ordem dos
  literais de string na pool de deduplicação).
- Considerar suporte a units/linkagem externa em `WIRTH80`.

---

# Release Notes — KIZUNA v4.6.0 "Kansei" (完成)

> "O laço agora amarra até os próprios bancos de memória — e prova, de novo, que só a execução real fecha um bug."

**Kansei** (完成) — "conclusão, obra completa, acabamento". Depois de *Kuyashii*
(悔しい, a frustração), *Yoake* (夜明け, o amanhecer) e *Kaisei* (快晴, o céu
limpo) da saga de SCREEN 2, esta release fecha o **roadmap original inteiro**
(`SPEC.md` Seção 8) com o envio de `OBI` — e, ao testá-lo em hardware real,
encontra e corrige uma regressão séria e há muito adormecida no suporte
multi-banco.

## Fase 6: OBI, o orquestrador de build declarativo

`OBI` (`pkg/obi` + `cmd/obi`) lê uma receita `Obifile` e invoca
`KAJI80`/`WIRTH80`/`DIGNAC` + `MUSUBI` (ou `HAKO`, quando o alvo é uma
biblioteca `.hlib`) na ordem certa — tudo em processo, reusando as mesmas
APIs que cada ferramenta já expõe, sem lançar subprocessos. Sem dependências
externas (o projeto inteiro é zero-dependency), o parser do `Obifile` é
hand-rolled, no mesmo espírito do `KAJI80`/`WIRTH80`/`DIGNAC`.

`sample/obi/` é a prova: uma receita real (não ilustrativa, diferente do
`demo/Obifile` original) combinando `KAJI80` (banco 0, dono do `Start`),
`DIGNAC` (banco 2, módulo biblioteca sem `PROCEDURE Main`), um resource
binário embutido, e a `MSXLIB` via `.hlib` — tudo com um único comando
`obi build`.

**Achado fora de escopo, registrado**: `WIRTH80` sempre emite seu próprio
`Start` e não tem sintaxe para chamar uma rotina externa arbitrária, então
não pode ser um módulo "biblioteca" numa ligação multi-módulo hoje — é por
isso que `demo/main.pas` continua só um croqui ilustrativo.

## Regressão real: o bootstrap multi-banco do MUSUBI não executava

Testar `sample/obi` em hardware real revelou um sintoma preocupante: o
programa carregava e voltava limpo ao prompt do MSX-DOS **sem executar
nada** — nem imprimir o `[L]` que o bootstrap multi-banco sempre imprime
primeiro. O mesmo teste com o já existente `sample/multibank` confirmou que
não era específico do OBI: **era geral a qualquer programa multi-banco**.

Investigação por histórico do Git revelou a causa: o multi-banco funcionava
desde a v4.1.0 "Akatsuki" com um mapeamento **identidade** simples (segmento
físico = número de banco do linker). A sessão da v4.5.2 "Yoake" trocou isso
por alocação **dinâmica** via `ALL_SEG` do EXTBIO — em teoria mais correta,
mas **nunca executada de verdade**, só validada por análise estática. Essa
troca foi a regressão.

Uma primeira correção (registrador `B` não inicializado antes de `ALL_SEG`,
confirmada contra a documentação oficial do protocolo EXTBIO) não resolveu
o problema. A correção definitiva foi reverter `buildBootstrapCode` para o
mapeamento identidade original — a única versão deste bootstrap já
confirmada funcionando em hardware — mantendo as melhorias de qualidade de
código da v4.5.2 que não tinham relação com o bug.

**Confirmado em hardware real pelo usuário**: `multibank.com` recompilado
imprime corretamente a sequência completa Banco 0 → Banco 1 → Banco 2 →
Banco 0, com retorno limpo ao MSX-DOS. Ver a listagem completa e a captura
de tela ao final destas notas.

## SCREEN 2: confirmação visual documentada (v4.5.3)

A v4.5.3 "Kaisei" já havia confirmado visualmente, em hardware real, os
dois bugs de SCREEN 2 corrigidos na v4.5.2 — ver a entrada anterior destas
notas e `README.md` para a listagem e a captura de tela completas.

## Estado do projeto

Com `OBI` enviado e o bootstrap multi-banco confirmado funcionando de novo,
**todo o roadmap original está implementado e validado em hardware**:
`KAJI80`, `WIRTH80`, `DIGNAC`, `MUSUBI`, `HAKO`, `MOBDUMP`, `MSXLIB`
(incluindo SCREEN 2 e multi-banco) e `OBI`.

### Próximos passos

- Expandir a `MSXLIB` (sprites, I/O de arquivo, som mais rico) — a camada
  que todo programa de aplicação realmente usa.
- Investigar o não-determinismo residual de `WIRTH80`/`DIGNAC` (ordem dos
  literais de string na pool de deduplicação).
- Considerar suporte a units/linkagem externa em `WIRTH80`, o que
  finalmente tornaria `demo/main.pas` realizável.

## Confirmação visual: bank switching automático real

```asm
; main.asm -- Banco 0 (Área Comum)
MODULE MAIN
BANK 0

PUBLIC Start
EXTERN PrintBank1, PrintBank2

BDOS       EQU 0x0005
C_WRITE    EQU 0x09

Start:
    ; 1. Mensagem a partir do Banco 0 (Área Comum)
    ld   de, MsgCommon
    ld   c, C_WRITE
    call BDOS

    ; 2. Chamada Cross-Bank: Salta para rotina no Banco 1
    ; O MUSUBI intercepta esta chamada e gera o trampolim automático!
    call PrintBank1

    ; 3. Chamada Cross-Bank: Salta para rotina no Banco 2
    ; O MUSUBI intercepta e troca para o Banco 2 na Página 2!
    call PrintBank2

    ; 4. Mensagem final do Banco 0 e retorno limpo ao MSX-DOS
    ld   de, MsgDone
    ld   c, C_WRITE
    call BDOS

    ret

MsgCommon:
    db 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "[Banco 0 - Area Comum] KIZUNA Multi-Banco Iniciado", 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "$"

MsgDone:
    db 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "[Banco 0] Execucao Multi-Banco Concluida com Sucesso!", 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "$"

ENDMOD
```

```asm
; bank1.asm -- Banco 1 (paginável na Página 2, 0x8000..0xBFFF)
MODULE BANK1
BANK 1

PUBLIC PrintBank1

BDOS       EQU 0x0005
C_WRITE    EQU 0x09

PrintBank1:
    ; Imprime mensagem oficial do Banco 1 na Página 2
    ld   de, MsgBank1
    ld   c, C_WRITE
    call BDOS
    ret

MsgBank1:
    db 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db ">>> [BANCO 1] ROTINA DO BANCO 1 EXECUTADA!", 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "$"

ENDMOD
```

```asm
; bank2.asm -- Banco 2 (paginável na Página 2, 0x8000..0xBFFF)
MODULE BANK2
BANK 2

PUBLIC PrintBank2

BDOS       EQU 0x0005
C_WRITE    EQU 0x09

PrintBank2:
    ; Imprime mensagem oficial do Banco 2 na Página 2
    ld   de, MsgBank2
    ld   c, C_WRITE
    call BDOS
    ret

MsgBank2:
    db 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db ">>> [BANCO 2] ROTINA DO BANCO 2 EXECUTADA!", 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "$"

ENDMOD
```

Resultado da execução em openMSX (MSX2+ Boosted, MSX-DOS 2) — `[L]` do
bootstrap, seguido das três mensagens (Banco 0 → Banco 1 → Banco 2 → Banco
0), e retorno limpo ao prompt:

![multibank.com rodando no openMSX, mostrando a troca de banco 0 -> 1 -> 2 -> 0 com sucesso](images/kizuna-01.png)

---

# Release Notes — KIZUNA v4.5.3 "Kaisei" (快晴)

> "Da frustração ao céu limpo — o mesmo laço que amarra as linguagens agora também desenha certo."

**Kaisei** (快晴) — "céu completamente limpo, tempo perfeito, sem uma nuvem".
O fechamento natural do arco iniciado em *Kuyashii* (悔しい, a frustração) e
continuado em *Yoake* (夜明け, o instante em que a escuridão cede): agora que
o sol nasceu, o céu está limpo — **o bug gráfico de SCREEN 2 foi confirmado
visualmente, em hardware/emulador real, pela primeira vez em toda a saga**.

A v4.5.2 já havia encontrado e corrigido, em software, os dois bugs reais que
explicavam o caos visual em `VDP_Line`/`VDP_BoxFill` (endereçamento indexado
`(IX+d)` mal codificado como imediato no `KAJI80`, e `VDP_PSet_Raw`
sobrescrevendo o byte inteiro do padrão em vez de preservar os demais pixels
da célula) — mas isso havia sido validado só por remontagem e leitura direta
dos bytes do `.MOB`, sem confirmação visual. A v4.5.3 fecha essa lacuna: o
pipeline completo **BASIC (`chart.bas`) → `DIGNAC` → `MUSUBI` → SCREEN 2**
agora desenha moldura, eixos, grade e curva corretamente na tela, sem nenhum
artefato — ver a captura de tela ao final destas notas.

Esta versão também traz uma correção de robustez encontrada na mesma
investigação, sem relação direta com o bug gráfico: `KAJI80` gravava a tabela
de símbolos `PUBLIC`/`EXTERN` do `.MOB` iterando um `map` do Go, cuja ordem
de iteração é embaralhada a cada execução do processo — inofensivo para o
linker (que resolve símbolos por nome), mas produzia arquivos `.mob`/`.com`/
`.map` byte-a-byte diferentes a cada remontagem do mesmo fonte, gerando
diffs espúrios permanentes em builds versionados no Git. Corrigido ordenando
os nomes alfabeticamente antes de serializar; confirmado remontando toda a
MSXLIB e os exemplos duas vezes seguidas e comparando os bytes (idênticos).

## Estado técnico atual - SCREEN 2

- **Confirmado.** `sample/basic/chart.bas` renderiza corretamente em
  hardware/emulador real: moldura, eixos cartesianos, grade e a curva de
  pontos calculada, sem pixels dispersos nem artefatos.
- Toda a toolchain (`KAJI80`, `WIRTH80`, `DIGNAC`, `MUSUBI`, `HAKO`,
  `MOBDUMP`) e a `MSXLIB` (incluindo as rotinas gráficas de VDP) estão agora
  **concluídas e validadas de ponta a ponta** — o bloqueio que se arrastava
  desde a v4.5.1 está fechado.

### Não-determinismo residual conhecido (fora de escopo desta release)

A correção de ordenação de símbolos acima resolve o `KAJI80` especificamente.
`WIRTH80` e `DIGNAC` (que geram Assembly intermediário e chamam o mesmo
`Assemble()`) têm uma fonte de não-determinismo própria e diferente: a ordem
dos literais de string na pool de deduplicação desses dois frontends muda
entre execuções, afetando o layout real do segmento de dados (não só a
tabela de símbolos). Registrado para uma sessão futura caso builds
reproduzíveis do lado Pascal/BASIC se tornem prioridade.

### Próximos passos

1. Avançar para a **Fase 6**: `OBI`, o orquestrador de build declarativo
   (`Obifile`) — última fase pendente do roadmap em `SPEC.md` §8.
2. Opcionalmente, investigar e eliminar o não-determinismo residual de
   `WIRTH80`/`DIGNAC` descrito acima.

As fontes em `resource/MSXgl` e `resource/MSXFusionC` permanecem material de
referência; não fazem parte do build da MSXLIB.

## Confirmação visual: `sample/basic/chart.bas` em SCREEN 2

Listagem completa do programa de exemplo cuja execução está na captura de
tela logo abaixo:

```basic
' ============================================================
' KIZUNA sample -- Modulo Chart em MSX-BASIC Dignified
' Compilador: DIGNAC
' ============================================================

MODULE Chart
BANK 0
PUBLIC Main, Desenhar
EXTERN BIOS_CHGET

' Ponto de entrada para demonstracao grafica standalone
PROCEDURE Main()
    SCREEN 2
    Desenhar(10)
    BIOS_CHGET()
    SCREEN 0
END PROCEDURE

' Desenhar: recebe um valor e traça um gráfico com moldura,
' eixos cartesianos, grade e curva calculada.
PROCEDURE Desenhar(valor%)
    LOCAL x%, y%

    ' Checkpoints: um PSET de cor unica por etapa, na coluna x=2 (fora da
    ' area da moldura/eixos/grade/curva, ninguem mais escreve ali), cada um
    ' numa linha Y diferente para nao dar color clash entre eles.
    ' Ordem/cores: 15 branco, 8 vermelho medio, 5 azul claro, 11 amarelo
    ' claro, 13 magenta, 7 ciano, 3 verde claro, 10 amarelo escuro,
    ' 6 vermelho escuro, 12 verde escuro, 9 rosa/vermelho claro.

    ' 1. Limpa a tela com fundo preto
    LINE (0,0)-(255,191), 1, BF
    PSET (2, 2), 15   ' checkpoint 1: BoxFill (limpar tela) OK
    BIOS_CHGET()

    ' 2. Moldura retangular externa branca (cor 15)
    LINE (8, 8)-(247, 8), 15
    PSET (2, 6), 8    ' checkpoint 2: borda topo OK
    BIOS_CHGET()
    LINE (247, 8)-(247, 183), 15
    PSET (2, 10), 5   ' checkpoint 3: borda direita OK
    BIOS_CHGET()
    LINE (247, 183)-(8, 183), 15
    PSET (2, 14), 11  ' checkpoint 4: borda baixo OK
    BIOS_CHGET()
    LINE (8, 183)-(8, 8), 15
    PSET (2, 18), 13  ' checkpoint 5: borda esquerda OK
    BIOS_CHGET()

    ' 3. Eixos cartesianos em Ciano (cor 7)
    LINE (24, 20)-(24, 165), 7
    PSET (2, 22), 7   ' checkpoint 6: eixo vertical OK
    BIOS_CHGET()
    LINE (24, 165)-(236, 165), 7
    PSET (2, 26), 3   ' checkpoint 7: eixo horizontal OK
    BIOS_CHGET()

    ' 4. Linhas de grade horizontais em Cinza (cor 14)
    LINE (24, 130)-(236, 130), 14
    PSET (2, 30), 10  ' checkpoint 8: grade 1 OK
    BIOS_CHGET()
    LINE (24, 95)-(236, 95), 14
    PSET (2, 34), 6   ' checkpoint 9: grade 2 OK
    BIOS_CHGET()
    LINE (24, 60)-(236, 60), 14
    PSET (2, 38), 12  ' checkpoint 10: grade 3 OK
    BIOS_CHGET()

    ' 5. Curva de pontos do grafico em Amarelo (cor 10)
    FOR x% = 25 TO 235
        y% = 160 - (x% MOD (valor% + 1)) * 9
        PSET (x%, y%), 10
    NEXT x%
    PSET (2, 42), 9   ' checkpoint 11: curva (loop completo) OK
    BIOS_CHGET()

END PROCEDURE
END MODULE
```

Resultado da execução em openMSX (MSX2+ Boosted, SCREEN 2, 256x192):

![chart.bas rodando em SCREEN 2 no openMSX — moldura branca, eixos em ciano, grade cinza e curva de pontos amarela, todos renderizados corretamente](images/kizuna-00.png)

---

# Release Notes — KIZUNA v4.5.2 "Yoake" (夜明け)

**Yoake** (夜明け) — "o romper da madrugada, o instante em que a escuridão
finalmente cede". A resposta direta a *Kuyashii* (悔しい, a frustração da
release anterior): depois de várias sessões investigando na direção errada,
a causa raiz real do bug gráfico de SCREEN 2 foi encontrada e corrigida.

A versão **v4.5.2** corrige **dois bugs independentes** que juntos explicam
todo o caos visual observado em `VDP_Line`/`VDP_BoxFill` desde o início da
investigação: o assembler `KAJI80` codificava operações ALU com operando
indexado (`CP (IX+d)`, `SUB (IX+d)`, etc.) como um imediato de 8 bits em
silêncio — as únicas ocorrências dessa forma no projeto inteiro estavam
justamente dentro das rotinas de linha/área da `MSXLIB` — e `VDP_PSet_Raw`
sobrescrevia o byte inteiro do padrão em vez de preservar os outros pixels
da mesma célula. De brinde, a mesma auditoria corrigiu dois pontos de
robustez no linker `MUSUBI`: um bug de offset de arquivo com segmentos BSS
em build multi-banco, e a falta de alocação real de segmento (`ALL_SEG` via
EXTBIO) para bancos pagináveis, que antes assumia que o número lógico de
banco do linker já era um segmento físico livre da Memory Mapper.

## Estado técnico atual - SCREEN 2

- **Corrigido em software, confirmado por remontagem e leitura direta dos
  bytes do `.MOB`/`.COM` gerados** (não apenas análise estática): os dois
  bugs acima descritos em detalhe no `CHANGELOG.md`.
- **Ainda não confirmado visualmente em hardware/emulador real** — esse é o
  único passo que falta para fechar definitivamente o problema.
- Todo o restante da toolchain (`KAJI80`, `WIRTH80`, `DIGNAC`, `MUSUBI`,
  `HAKO`, `MOBDUMP`) segue **concluído e validado**; o bug estava isolado às
  rotinas gráficas de VDP da `MSXLIB`.

### Retomada da próxima sessão

1. Rodar `sample/basic/chart.bas` em hardware/emulador real e confirmar
   visualmente a moldura, os eixos e a curva.
2. Se confirmado: fechar o bloqueio de vez e seguir para a Fase 6 (`OBI`).
3. Se ainda houver artefato: já não é mais nenhuma das causas descartadas em
   sessões anteriores (timing de VRAM, atomicidade, motor de comando de
   hardware) nem os dois bugs corrigidos aqui — investigar do zero com um
   dump de VRAM durante a execução real.

As fontes em `resource/MSXgl` e `resource/MSXFusionC` permanecem material de
referência; não fazem parte do build da MSXLIB.

---

# Release Notes — KIZUNA v4.5.1 "Kuyashii" (悔しい)

**Kuyashii** (悔しい) — sentimento profundo de frustração honrosa e inconformismo por não ter atingido o resultado gráfico esperado no momento, mas acompanhado da convicção e energia para retornar, perseverar e conquistar a solução definitiva.

A versão **v4.5.1** consolida a integração completa do compilador **`DIGNAC`** (MSX-BASIC Dignified) com o ecossistema KIZUNA, exporta os símbolos de desenho em tela da biblioteca **`MSXLIB`** e prepara o terreno para a depuração fina do subsistema gráfico TMS9918/V9938 na próxima iteração.

## Estado técnico atual - SCREEN 2

- **Validado:** entrada em SCREEN 2, escrita direta na VRAM, exibição de um
  ponto e retorno ao prompt do MSX-DOS 2.
- **Em investigação:** artefatos fixos na tela durante a inicialização e o
  cálculo completo das relações entre Pattern Generator, Name Table, Color
  Table e as três páginas verticais da SCREEN 2.
- **Ainda não considerar concluído:** `VDP_PSet`, `VDP_Line`, `VDP_BoxFill`,
  texto gráfico e o exemplo `sample/basic/chart.bas`.

### Retomada da próxima sessão

1. Reproduzir o teste mínimo e registrar os bytes de VRAM observados.
2. Executar uma inicialização de SCREEN 2 baseada em uma única página e uma
   única célula 8x8, sem `PSET` ou `LINE`.
3. Confirmar o mapeamento entre Name Table, Pattern Table e Color Table.
4. Revalidar `VDP_PSet` com uma única célula antes de implementar linhas.
5. Só então avançar para `VDP_Line`, texto gráfico e o `chart.bas`.

As fontes em `resource/MSXgl` e `resource/MSXFusionC` permanecem material de
referência; não fazem parte do build da MSXLIB.

---

# Release Notes — KIZUNA v4.5.0 "Hinode" (日の出)

**Hinode** (日の出) — "o nascer do sol, a alvorada radiante". Na continuidade do amanhecer do KIZUNA, a versão **v4.5.0** traz o nascimento do compilador **`DIGNAC`** (MSX-BASIC Dignified para Z80 / MSX2+), completando o trio de linguagens fundamentais da toolchain: **Assembly Z80 (`KAJI80`)**, **Pascal (`WIRTH80`)** e **BASIC Estruturado (`DIGNAC`)**.

---

## O que é esta release

Esta é a release oficial **v4.5.0 pre-alpha**, trazendo uma toolchain Z80 completa para MSX2+ / MSX-DOS 2 com suporte à compilação de código **Assembly Z80**, **Pascal** e **MSX-BASIC Dignified**, geração de objetos relocáveis `.MOB`, paginação automática na página 2 com Memory Mapper, suporte ao frame pointer `IX` (ABI de pilha) e biblioteca padrão **MSXLIB** com rotinas gráficas de alta velocidade.

---

## Destaques da Release

1. **DIGNAC (Compilador MSX-BASIC Dignified Z80 - Fase 5.3)**:
   - Compila código BASIC estruturado diretamente para módulos relocáveis `.MOB`.
   - Suporte a `MODULE`, `BANK`, `PUBLIC`, `EXTERN`, `PROCEDURE`, `SUB`, `FUNCTION`.
   - Quadro de pilha compatível com ABI Kizuna usando ponteiro de frame `IX` e variáveis `LOCAL x%, y%`.
   - Controle de fluxo estruturado: `FOR ... NEXT`, `WHILE ... WEND`, `DO ... LOOP`, `IF ... THEN ... ELSE ... END IF`.
   - Primitivas gráficas nativas: `LINE (x1,y1)-(x2,y2)[, color][, B | BF]` e `PSET (x, y)[, color]`.
   - Geração automática de ponto de entrada `Start` para criação direta de executáveis `.COM` standalone via `musubi`.
   - Opção `-S` para inspeção do código Assembly Z80 gerado.

2. **Suporte a Endereçamento Indexado no KAJI80**:
   - Implementado suporte completo a `LD r, (IX+d)`, `LD (IX+d), r`, `LD (IX+d), n` (e registradores `IY`).
   - Implementado suporte a `LD SP, IX` e `LD SP, IY` (`DD/FD F9`).

3. **Primitivas Gráficas na MSXLIB (`lib/src/vdp.asm`)**:
   - `VDP_PSet`: Plotagem de pixel em SCREEN 2 (256x192) calculando endereços na Pattern e Color Table.
   - `VDP_BoxFill`: Preenchimento de retângulos e limpeza de tela cheia ultrarrápida.
   - `VDP_Line`: Desenho de retas e segmentos na tela gráfica.

4. **WIRTH80 (Compilador Pascal Z80)**:
   - Suporte a variáveis `Integer`, `Char` e `Boolean`, blocos `begin ... end.`.
   - Aritmética de 16 bits, chamadas de sistema, `WriteLn` e integração com `MSXLIB`.

5. **Exemplos em BASIC (`sample/basic/`)**:
   - `hello.bas`: Hello World em BASIC compilado (apenas 186 bytes).
   - `calc.bas`: Demonstração de variáveis locais, aritmética e chamadas à biblioteca padrão.
   - `chart.bas`: Módulo gráfico paginado no banco 2 para desenho de curvas e gráficos.
   - `build.ps1`: Script de compilação e smart-linking automatizado.

---

## Conteúdo do Pacote de Distribuição (`kizuna-v4.5.0-dist.zip`)

```
distribute/
  ├── bin/
  │    ├── kaji80.exe     (Assembler Z80)
  │    ├── wirth80.exe    (Compilador Pascal)
  │    ├── dignac.exe     (Compilador BASIC Dignified)
  │    ├── musubi.exe     (Linker com Smart-Linking)
  │    ├── hako.exe       (Bibliotecário)
  │    └── mobdump.exe    (Inspecionador de objetos)
  ├── lib/
  │    └── msxlib.hlib    (Biblioteca Padrão MSX)
  ├── sample/
  │    ├── hello.asm      (Exemplo Assembly monobanco)
  │    ├── multibank/     (Exemplo Assembly multi-banco com 3 módulos)
  │    ├── libdemo/       (Exemplo Assembly consumindo MSXLIB)
  │    ├── pascal/        (Exemplos em Pascal: hello.pas e calc.pas)
  │    └── basic/         (Exemplos em BASIC: hello.bas, calc.bas, chart.bas)
  ├── docs/
  │    ├── README.md
  │    ├── HELP.md
  │    └── CHANGELOG.md
  ├── install.exe / install.cmd
  └── LICENSE
```

---

## Próximos Passos

- **Fase 6**: Orquestrador de build declarativo **OBI** (`Obifile`) para compilar e linkar projetos multi-linguagem em um único comando.
- Construção da demo poliglota completa (`demo/`) unindo Pascal (`main.pas`), Assembly (`screen.asm`) e BASIC (`chart.bas`).

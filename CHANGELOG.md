# Changelog

Todas as mudanças notáveis deste projeto são documentadas aqui.
Formato baseado em [Keep a Changelog](https://keepachangelog.com/).

## [Não lançado]

### MSXLIB dividida em módulos pequenos (um por rotina/família)

`lib/src/{bdos,bios,vdp,psg,string,math}.asm` deram lugar a 56 módulos em
`lib/src/<area>/*.asm` (+ `lib/src/float.asm`), com as constantes de
porta/endereço compartilhadas em `lib/inc/*.inc`. O conjunto de símbolos
públicos é **idêntico** ao anterior (120), então nenhum programa muda, mas
o linker agora só traz o que cada programa usa: `predefined.com` 1774→382
bytes, `music.com` 1895→522, `sprites.com` 1603→642, `libdemo.com` 893→384.
Todos os samples foram relinkados e comparados no emulador Z80 (saída de
console + traço de escritas de porta VDP/PSG) contra a versão anterior:
comportamento idêntico. Também: `BIOS_Call` não arrasta mais o módulo de
VDP inteiro (havia uma dependência circular bios↔vdp em nível de módulo).
Base para a expansão da MSXLIB no estilo MSXgl.

### MSXLIB Fase 1 — memória (`lib/src/mem`, 22 rotinas)

Cópia com tratamento de sobreposição (`MEM_Copy`, como `memmove`) e variantes diretas, preenchimento
(8 e 16 bits), zerar, trocar, comparar, buscar, `MEM_GetSP`/`MEM_TPATop`, e um **heap dinâmico**
(`MEM_HeapInit`/`InitToStack`, `MEM_Alloc`, `MEM_Free` com recusa de ponteiro inválido e liberação
dupla, `MEM_BlockSize`, `MEM_HeapCompact`, `MEM_HeapSize/Free/Largest`). O heap é comparado **byte a
byte** com um modelo em Go depois de cada operação (12 sequências aleatórias de 600 passos) mais
checagens independentes (blocos ladrilham a região, conteúdo dos blocos vivos intacto). Descritores em
`lib/api/mem.api`. Itens do guia deixados de fora de propósito (alocador estático, `Mutex_*`) estão
em `docs/cobertura.map` com o motivo.

### KAJI80: conjunto Z80 documentado completo

Faltavam instruções básicas: operações de bloco (`LDIR`, `LDDR`, `CPIR`, `CPDR`, `INIR`, `OTIR`…),
`LD BC/DE/SP/IX/IY,(nn)` e inversas, `DAA`, `IM`, `RST`, `RLD`/`RRD`, `RETI`/`RETN`, `LD I/R,A`,
`JP (HL)`/`(IX)`/`(IY)`, `IN/OUT r,(C)`, `ADD IX,IX`, e as formas indexadas de rotação/`BIT`/`RES`/
`SET`/`INC`/`DEC`. Todas com teste de bytes contra a documentação oficial. **Bug silencioso
corrigido**: `IN A,(C)` e `OUT (C),A` eram montados como `IN A,(0)`/`OUT (0),A` (`DB 00`/`D3 00`);
`EX AF,AF'` não montava (o `'` abria um literal de caractere). O manual (§13) dizia que
`JP (HL)` era suportado, mas não era — agora é.

### MSXLIB Fase 1 — matemática (`lib/src/math`, ~45 rotinas novas)

Inteiros de 8/16/32 bits com e sem sinal (`MATH_*`): resto, divisão e resto com sinal, produtos
completos de 32 bits, `Neg/Abs/Sign`, comparação/mínimo/máximo/`Clamp`, deslocamentos de 8 e 16
bits com contagem variável, raiz quadrada inteira, divisão por 10, inversão de bits, ponto fixo
8.8 (`FixMul88/FixDiv88`), divisão 32/16 e gerador pseudoaleatório (xorshift, período 65535, com
faixa e intervalo). Cada rotina foi testada no simulador Z80 contra uma referência em Go
(exaustivamente quando barato) **incluindo a checagem de que os registradores não-saída foram
preservados**. Descritores em `lib/api/math.api` (chamáveis do BASIC/Pascal).
KAJI80: novo `EX (SP),HL/IX/IY`. Cobertura: 25 dos 29 itens de matemática do guia
(`docs/cobertura.md`); `QMN_Get8/16` ficam pendentes de propósito (semântica de arredondamento
diferente do nosso deslocamento aritmético).

### Fase 0 da expansão da MSXLIB (ver `docs/plano-expansao-msxlib.md`)

- **Descritores de API (`.api`)**: DIGNAC e WIRTH80 passam a chamar rotinas da MSXLIB de
  convenção de registradores como qualquer procedure/função (`PSG_Write(7, 62)`,
  `x% = Mul16(6, 7)`), sem comando dedicado no compilador e sem `EXTERN`. Novo pacote
  `pkg/api`, opção `-api` nos dois compiladores, chave `api:` no `Obifile`, e
  `lib/api/{bdos,bios,vdp,psg,string,math}.api` descrevendo as rotinas existentes. Erros de
  contagem de argumentos e `proc` em expressão são erros de compilação; o compilador
  salva/restaura IX em volta da chamada. Exemplo: `sample/api/` (BASIC e Pascal).
- **Simulador Z80 no repositório** (`pkg/z80sim`, `pkg/libtest`): roda `.COM` ou uma rotina só,
  com BDOS/CALSLT simulados e registro de escritas de porta — a mesma técnica que achou o
  bug `LD B,(nn)`, agora usada nos testes de tudo o que é novo.
- **Constantes dos chips em `lib/inc/`**: portas, VDP (registradores, comandos, status), PSG,
  YM2413, Y8950, SCC, V9990 (portas), ASCII e cores.
- **`docs/cobertura.md`** (gerado por `tools/cobertura`): checklist do que o MSXgl e o
  Fusion-C oferecem, com o estado de cada item no KIZUNA; `docs/CREDITS.md`.

### KAJI80: `INCLUDE "arquivo"`

Insere o texto de outro arquivo-fonte (relativo ao arquivo que inclui,
aninhável, com detecção de inclusão circular). Ver `docs/manual-assembly.md` §12.

## [4.11.0] - 2026-09-23 - Release Kakuchou (拡張)

### KAJI80 ganha recursos de macro-assembler ao estilo asMSX, com sintaxe própria do KIZUNA

8 fases entregues, aprovadas de antemão como um plano único: avaliador de
expressões numéricas em tempo de montagem (aritmética, bits, lógica,
funções matemáticas, `PI`, ponto fixo `FIX`/`FIXMUL`/`FIXDIV`/`INT`,
`RANDOM`), rótulos locais (`.nome:`, escopados ao rótulo global mais
recente), montagem condicional (`IF`/`ELSE`/`ENDIF`), repetição de bloco
(`REPT`/`ENDR`), macros (`MACRO @param.../ENDM`), rótulos pré-definidos
de BIOS/BDOS/variáveis de sistema (sem precisar de `EXTERN`),
`CALLBIOS`/`CALLDOS` (chamada inter-slot de BIOS e chamada de MSX-DOS) e
`INCBIN "arquivo", SKIP=x, SIZE=y`.

Vários desvios deliberados da sintaxe literal do asMSX, sempre pra evitar
quebrar uma convenção que o KAJI80 já tinha antes desta leva: `MOD` em
vez de `%` (já é o prefixo de literal binário), `@param` em vez de
`#param` nas macros (já é o prefixo de literal hexadecimal), `SKIP=x`/
`SIZE=y` em vez de espaço (evita uma ambiguidade real na reconstrução de
operando). A tabela `BDOS` de códigos de função do MSX-DOS é nova — o
asMSX não tem equivalente.

Dois bugs reais encontrados e corrigidos como efeito colateral de
reproduzir exemplos reais do asMSX em testes, não como objetivo direto de
nenhuma fase: `INC (HL)`/`DEC (HL)` nunca tinham sido implementados
(instrução Z80 padrão), e a separação de operandos por vírgula não
respeitava profundidade de parênteses (`DB POW(2,3)` quebrava errado em
dois operandos).

Um terceiro bug, mais sutil, só apareceu testando os recursos **juntos**:
um literal decimal com ponto (`FIX(1.5)`) colidia com a resolução de
rótulo local, porque o lexer nunca soube reconhecer ponto decimal em
número (`"1.5"` virava dois tokens, `NUMBER "1"` + `IDENTIFIER ".5"`, já
que `.` sempre foi caractere de identificador válido). Corrigido na raiz
no lexer — nenhuma fase individual tinha esse bug isolada, só a
combinação.

Novo `sample/macroasm/` — três programas reais em KAJI80 puro (sem
DIGNAC/WIRTH80) exercitando os 8 recursos, cada um imprimindo
`[OK]`/`[FALHOU]` visível via BDOS pra confirmação em hardware real.
Verificados num emulador Z80 (harness próprio deste projeto, usado antes
pro motor de float) antes da entrega: saída impressa capturada e
conferida por completo pra dois dos três programas; o terceiro
(`CALLBIOS CHGMOD`, troca de tela de verdade) teve a chamada inter-slot
confirmada até o ponto de entrada real da ROM — a mudança de tela em si
só é verificável em hardware/openMSX de verdade.

### DIGNAC ganha aritmética SINGLE de verdade: Float_Add32/Sub32/Cmp32

`x! = a! + b!`, `x! = a! - b!` e comparação (`=`,`<>`,`<`,`<=`,`>`,`>=`)
agora funcionam para `SINGLE` (IEEE 754 binary32), via um motor de
software novo em `lib/src/float.asm` (`Float_Add32`/`Float_Sub32`/
`Float_Cmp32`, MSXLIB). Escopo desta leva: operandos simples (variável
ou literal, sem aninhar — `x! = (a!+b!)*c!` continua erro de compilação
claro), `*`/`/` e `DOUBLE` ficam pra uma leva futura (a normalização de
mantissa 24×24→48 bits se mostrou bem mais delicada que soma/subtração).
`PRINT` de float continua fora de escopo.

Algoritmo verificado num protótipo em Go contra ~200 mil pares
aleatórios antes de tocar em Z80 — achou e corrigiu 2 bugs de
normalização antes de escrever a primeira linha de assembly. Na
transliteração, mais 3 bugs reais encontrados e corrigidos: KAJI80 não
suporta aritmética `Label+N` em operandos (virava `EXTERN` fantasma
silencioso), `LD DE,(nn)` não existe no Z80 de verdade, e comparação de
expoente signed usando `CP` direto (comparação unsigned) errava sempre
que os dois expoentes tinham sinais diferentes.

### Bug real encontrado só em hardware — e uma ferramenta nova pro projeto: um emulador Z80

Wilson testou o exemplo (`sample/basic/floatmath.bas`) em hardware/
openMSX: soma falhava, uma das duas comparações falhava, subtração e a
outra comparação funcionavam — um padrão que resistiu a várias
re-verificações por hand-tracing (não existe emulador Z80 no repositório
até então). Causa raiz só foi achada montando um harness em Go com
`github.com/remogatto/z80` (scratchpad, primeira vez que o projeto tem
acesso a execução real de Z80 pra depuração) pra *rodar* o `.com`
linkado de verdade: `Float_Cmp32` usava `LD B, (Float_MantHi)` — forma
que não existe no Z80 (só `LD A,(nn)` tem endereçamento absoluto de 16
bits pra um registrador de 8 bits) — e o KAJI80 montava isso em
silêncio como `LD B, 0`. Só quebrava quando o byte mais significativo
(sinal+topo do expoente) de A e B já eram iguais — qualquer comparação
de igualdade, ou dois positivos com faixa de expoente parecida.
Corrigido e **reconfirmado em hardware pelo usuário** no mesmo dia.
Verificado também com 20.000 casos aleatórios rodados no emulador contra
a aritmética float32 nativa do Go (`Cmp32` 0 falhas).

### Auditoria do KAJI80 fecha uma classe inteira de bug de "operando de memória vira imediato em silêncio"

A pedido de Wilson, depois do bug acima: qualquer operando que parece
um endereço de memória (`(algo)`) mas não bate com nenhuma forma real do
Z80 pra aquela instrução caía num fallback de imediato/símbolo sem
nenhuma validação, virando `0` ou um "símbolo" fantasma em silêncio —
mesma causa raiz do já conhecido bug de `Label+N`. `emitAddressOrReloc`/
`emitRelativeOrReloc` (usadas por `CALL`, `JP`, `JR`, `DJNZ`, `LD
(nn),A/HL`, `LD HL,(nn)`, `LD rr,nn`, `LD IX/IY,nn`, `DW`) agora validam
o nome do símbolo antes de aceitar — fecha essa classe inteira de uma
vez (inclusive `LD DE,(nn)`/`LD BC,(nn)`, de brinde). `encodeAlu8`
ganhou o mesmo tipo de guarda pra endereço absoluto entre parênteses
(`CP (Algo)` virava `CP 0` em silêncio). Rebuild completo de MSXLIB e
todos os exemplos confirma que nada legítimo dependia do comportamento
antigo.

### WIRTH80 ganha a diretiva BANK <n>

Programas Pascal (`WIRTH80`) agora podem declarar `BANK <n>;` logo após
`program Nome;` (mesma posição do `KAJI80`/`DIGNAC`, terminada com `;`
pra combinar com a sintaxe `PUBLIC`/`EXTERN` já estabelecida) — antes,
todo módulo `WIRTH80` caía sempre no banco comum. `sample/obi/`
(`KAJI80`+`DIGNAC`+`WIRTH80` linkados num único `.COM`) agora tem cada
linguagem em seu próprio banco de verdade, não mais duas linguagens
forçadas a dividir o banco 0.

## [4.9.0] - 2026-09-23 - Release Yuugou (融合)

### As 3 linguagens de entrada linkam juntas num único .COM — pela primeira vez, testado e confirmado em hardware

`WIRTH80` ganha `procedure`/`function` definidas pelo usuário (posição
clássica de Pascal, valor de retorno por atribuição ao próprio nome —
`Nome := expr;`, estilo Turbo Pascal) e `PUBLIC`/`EXTERN` com a mesma
palavra-chave nua do `KAJI80`/`DIGNAC`, reaproveitando a ABI de pilha já
provada (parâmetros empilhados esquerda→direita, frame pointer `IX`,
limpeza pelo chamador). `Start`/`PUBLIC Start` agora só é gerado se o
bloco principal (`begin...end.`) tiver algum comando — mesma regra que o
`DIGNAC` já usa pra `PROCEDURE Main` — permitindo que um módulo `WIRTH80`
vire uma "biblioteca" pura, sem ponto de entrada próprio.

Isso, combinado com **exatamente um Main por executável agora sendo
garantido**: `MUSUBI` dá um erro específico ("múltiplos pontos de
entrada...") em vez do genérico "símbolo duplicado" quando mais de um
módulo tenta definir o ponto de entrada — decisão explícita de Wilson,
já que a alternativa (deixar o usuário escolher no link, ou avisar sem
bloquear) abriria espaço pra ambiguidade silenciosa.

`sample/obi/` ganha um terceiro módulo, `greet_lib.pas` (`WIRTH80`,
biblioteca pura) — prova real, não só testes unitários isolados, de que
as 3 linguagens linkam juntas: `main.asm` (`KAJI80`, banco 0, dono do
`Start`) chama tanto `Desenhar` (`chart_lib.bas`, `DIGNAC`, banco 2,
cross-bank) quanto `Saudacao` (`greet_lib.pas`, `WIRTH80`, banco 0,
`CALL` direto).

**Achado real no processo**: `WIRTH80` ainda não suporta a diretiva
`BANK <n>` (todo módulo seu cai sempre no banco comum) — documentado,
não bloqueia o objetivo desta leva.

**Bug real encontrado e corrigido ao escrever o `CallExpr` do `WIRTH80`**:
o mesmo padrão já existia no `DIGNAC` (`pkg/dignac/codegen.go`) com um
bug nunca pego — somava o tamanho de limpeza da pilha direto no `HL` do
retorno antes de trocar registradores, corrompendo o valor de qualquer
chamada de função usada dentro de uma expressão maior (nunca exercitado
por nenhum teste/sample antes, só chamadas como comando isolado).
Corrigido nos dois compiladores.

### Bug real corrigido: trampolim de banco gerado por engano pra chamadas de volta ao banco comum — tela preta em SCREEN 2

Reportado por Wilson ao testar o `sample/obi/main.com` de 3 linguagens:
banner e `SCREEN 2` apareciam, mas o gráfico nunca desenhava (tela
preta), voltava pra tela de texto e encerrava normalmente. Isolado
reconstruindo a versão de 2 linguagens (pré-`WIRTH80`) — **mesmo bug**,
confirmando que era pré-existente, não introduzido pelas mudanças desta
sessão.

Causa raiz: `DIGNAC PROCEDURE Desenhar` (banco 2) chama de volta rotinas
da `MSXLIB` no banco comum (`VDP_PSet`, `VDP_BoxFill`, `VDP_Line`,
`Mul16`, `Div16`) — e `MUSUBI` gerava um trampolim de troca de banco pra
cada uma dessas chamadas, mesmo o alvo sendo o banco comum, que está
**sempre** presente na memória independente do que a Página 2 tem
mapeado no momento. Esse trampolim lê uma tabela de bancos pra saber
qual segmento físico ativar — só que o bootstrap nunca escreve a entrada
do banco 0 (só popula 1..N; o próprio comentário no código já dizia
"desperdício na entrada 0 (banco comum, nunca paginado)", mas a suposição
de que nenhum trampolim miraria o banco 0 estava errada). Resultado: um
`OUT` pra Página 2 com um segmento físico nunca inicializado antes de
cada chamada de desenho.

Nunca tinha sido pego porque o bug do bootstrap multi-banco corrigido no
dia anterior foi validado via `sample/multibank`, que só chama `BDOS`
diretamente a partir dos bancos pagináveis — nunca uma rotina `MSXLIB` no
banco comum. Esse caminho específico (banco paginável chamando de volta
pro banco comum) nunca tinha sido exercitado por nenhum teste
confirmado antes.

Corrigido: uma chamada mirando o banco comum agora sempre vira `CALL`
direto, não importa de onde vem — mesma otimização que já existia pra
"mesmo banco", estendida. `sample/obi/main.com` caiu de 5 trampolins (4
deles quebrados) pra 1 só, 2499 → 2399 bytes. **Confirmado corrigido em
hardware pelo usuário** — primeira vez na história do projeto que esse
caminho (banco paginável chamando de volta pro banco comum) funciona de
verdade em execução real.

### Dois bugs pequenos do DIGNAC corrigidos

`FOR...STEP` negativo (o teste de término do laço assumia sempre passo
positivo) e `OPEN...FOR APPEND` (se comportava igual a `FOR OUTPUT`,
truncando em vez de acrescentar) — ambos documentados como limitação
conhecida na v4.8.0, corrigidos logo em seguida como prometido.

## [4.8.0] - 2026-09-22 - Release Minori (実り)

### DIGNAC ganha um sistema de tipos real: STRING, INTEGER, SINGLE, DOUBLE

Até aqui o `DIGNAC` tratava toda variável não-BOOLEAN como uma palavra de
16 bits, sempre — `LOCAL a%, b$, c!` dava o mesmo tipo pra todo mundo na
lista (o primeiro sufixo "vencia"), e só literais de string funcionavam de
verdade (nenhuma variável STRING existia). `LocalDeclNode`/`DimDeclNode`
foram redesenhados para `Decls []VarDecl{Name,Type}`, resolvendo o tipo de
cada variável pelo próprio sufixo (`%`=INTEGER, `$`=STRING, `!`=SINGLE,
`#`=DOUBLE), com `AS <Tipo>` continuando como override explícito pra toda
a lista.

STRING funciona de ponta a ponta agora: declarar, atribuir (literal ou
outra variável STRING) e imprimir — buffer de 256 bytes (1 byte de tamanho
+ até 255 de dados, o mesmo formato "short string" já documentado em
`SPEC.md` §7), via duas rotinas novas na `MSXLIB`
(`StrCopyLen`/`BDOS_PrintLenStr`, `lib/src/string.asm`).

SINGLE/DOUBLE usam **IEEE 754** (binary32/binary64), não o MBF que o
MSX-BASIC real usa internamente — decisão deliberada, já que o `DIGNAC`
compila pra código nativo standalone, sem necessidade de compatibilidade
binária com um interpretador. São declaráveis, aceitam literais (ponto
decimal, expoente `e`/`E`=SINGLE `d`/`D`=DOUBLE, sufixo `!`/`#`) e são
atribuíveis por cópia crua de bytes — mas **qualquer aritmética ou `PRINT`
de uma variável SINGLE/DOUBLE é um erro de compilação claro** ("ainda não
implementada"), nunca um resultado errado calado. A engine de ponto
flutuante completa (+,-,\*,/, comparação, conversão pra texto decimal) fica
para uma sessão futura dedicada — sozinha, é do tamanho de escrever uma
biblioteca de ponto flutuante em Z80 do zero.

Também adiciona literais `&O` (octal) e `&B` (binário), no mesmo padrão do
`&H` já existente.

Seis testes novos em `pkg/dignac/compiler_test.go`. Novo sample
`sample/basic/types.bas` — compila, linka e **foi confirmado em hardware
real (openMSX)**: as duas strings e os dois valores inteiros apareceram
corretos, sem lixo na tela.

### Documentação reorganizada em manuais por assunto

`HELP.md` tinha virado um arquivo monolítico desatualizado (sem `OBI`, sem
as novidades recentes do `DIGNAC`). Dividido em quatro manuais novos em
`docs/`, cada um atualizado contra o estado real do código — não a visão
aspiracional do `SPEC.md`:

- `docs/manual-assembly.md` (`KAJI80`), `docs/manual-basic-dignified.md`
  (`DIGNAC`), `docs/manual-pascal.md` (`WIRTH80`) — sintaxe de cada
  linguagem, incluindo suas próprias limitações conhecidas.
- `docs/manual-ferramentas.md` — formato `.MOB`/`.MAP`, `MUSUBI`, bank
  switching, `HAKO`/`.HLIB`, `MSXLIB`, `OBI`, `MOBDUMP`, e um exemplo
  end-to-end combinando as três linguagens.

`MANUAL.md` e `HELP.md` viraram índices curtos apontando pros manuais.

**Achado real durante a pesquisa para o manual do Pascal**: `WIRTH80` hoje
é um subconjunto bem menor do que `SPEC.md` descreve — sem
procedimentos/funções definidos pelo usuário, sem `uses`/units, sem
`PUBLIC`/`EXTERN` do lado do programador, e o laço `for...to/downto...do`
tem os tokens reservados no lexer mas **nenhum caso no parser** (usar `for`
hoje é erro de sintaxe). O tipo `String` em `var` é aceito sintaticamente
mas tratado internamente como um `Integer` sem nenhuma semântica de string
— atribuir um literal de string a ele é erro de compilação. `demo/main.pas`
e `demo/Obifile` (que mostram Pascal chamando um módulo Assembly via
`{$USES}`) continuam corretamente marcados como "croqui hipotético, não
compila" — documentado agora de forma explícita em vez de implícita.

## [4.7.0] - 2026-09-22 - Release Kaika (開花)

### MSXLIB expandida: sprites, música PSG e I/O de arquivo — todos confirmados em hardware

Primeira expansão real da `MSXLIB` desde que o roadmap original fechou na
v4.6.0. Toda a informação de hardware foi verificada contra fontes
primárias antes de escrever qualquer código (registradores de sprite
contra `resource/MSXgl/engine/src/vdp_reg.h`; números de função de
arquivo contra o protocolo oficial de MSX-DOS 2, diferentes dos números
genéricos de CP/M/MS-DOS que se assumiria por padrão).

- **`lib/src/vdp.asm`**: `VDP_SpriteDefine`, `VDP_SpriteSet`,
  `VDP_SpriteHide`, `VDP_SpriteHideAll`, `VDP_SpriteSetSize`. Sprite
  Attribute Table em `1B00h`, Sprite Pattern Table em `3800h` — os mesmos
  endereços-padrão que o MSX-BASIC já usa em SCREEN 2, encaixando nos vãos
  livres entre as tabelas existentes.
- **`lib/src/psg.asm`**: `PSG_NoteTable` (5 oitavas, C2..B6, 60 períodos
  calculados pela fórmula padrão e conferidos contra A4=440Hz→254),
  `PSG_PlayNoteIndexed`, `PSG_PlaySequence`.
- **`lib/src/bdos.asm`**: `BDOS_FileOpen/Create/Close/Read/Write/Seek` —
  funções de MSX-DOS 2 baseadas em handle (43h-4Ah).
- **`sample/sprites/`, `sample/music/`, `sample/fileio/`**: um exemplo
  KAJI80 puro por área (nem `WIRTH80` nem `DIGNAC` ainda conseguem chamar
  rotinas externas por convenção de registrador arbitrária — dar acesso
  aos compiladores fica para uma sessão futura).

**Confirmado em hardware real pelo usuário**: sprites e música funcionaram
de primeira. I/O de arquivo criava e escrevia o arquivo corretamente, mas
a leitura de volta imprimia caracteres bagunçados na tela — ver o bug
real abaixo, encontrado e corrigido no mesmo dia.

### Bug real no KAJI80: literal de caractere entre aspas simples virava zero em silêncio

`sample/fileio/main.asm` usava `LD (HL), '$'` para marcar onde a função
09h da BDOS deveria parar de imprimir (o arquivo lido não tem esse
terminador). `pkg/kaji80/assembler.go`'s `parseLine` reconstrói os
operandos de uma instrução concatenando `token.Value` diretamente, e o
lexer já devolve o conteúdo de uma string **sem** aspas (correto para
`DB`, que lê os tokens crus, não os operandos reconstruídos aqui) — então
`'$'` virava só `$`, indistinguível de um prefixo hexadecimal malformado
para `parseImm8`, que falhava em silêncio e devolvia `0`. `LD (HL),'$'`
virava `LD (HL),0x00` (`36 00`) em vez de `LD (HL),0x24` (`36 24`), **sem
nenhum erro de montagem** — mesma classe de bug já vista com `(IX+d)` na
saga da SCREEN 2. Com o terminador virando um byte nulo em vez de `'$'`
de verdade, a rotina de impressão nunca parava e imprimia memória adiante
indefinidamente — exatamente o "monte de caracteres bagunçados" relatado.

Corrigido na raiz: `parseLine` agora devolve as aspas simples ao redor do
valor de um `TokenString` ao reconstruir o operando, para que o suporte a
literal de caractere que `parseImm8` já tinha (mas nunca recebia a string
com aspas) passe a funcionar de verdade. `DB` não é afetado — usa os
tokens crus diretamente. Novo teste `TestCharLiteralImmediate`
(`pkg/kaji80/assembler_test.go`). `sample/fileio` também trocou `'$'` por
`24h` explícito, por segurança. **Confirmado pelo usuário em hardware
real após a correção.**

### Limpeza do disco de teste

`sample/` é usado como "disk in a folder" pelo openMSX para rodar os
`.com` do projeto durante os testes. Durante essa mesma sessão de
depuração, um bug lateral: vários arquivos antigos e sem relação com o
KIZUNA (`sample/MSXDOS.SYS`, `sample/UTILS/*`, `sample/hello.mob`/`.map`,
`sample/screen2_test.*`) apareceram deletados do disco — quase certamente
efeito colateral de algum comando MSX-DOS 2 rodado no openMSX contra essa
mesma pasta montada como disco real. Restaurados via `git restore` antes
de qualquer commit.

A pedido do usuário, o disco de teste foi então deliberadamente reduzido
para caber no limite de 720KB de um disquete DD padrão: removidos
`sample/screen2_test.*` (saga da SCREEN 2 já encerrada), `sample/UTILS/`
reduzido de 21 para 5 arquivos (mantendo só `CHKDSK`, `COPY.BTM`, `DIR.BTM`,
`MORE`, `TREE` — o resto não tem relação com testar programas KIZUNA), e
`sample/HELP/` (433K de textos de ajuda do MSX-DOS 2) removido por
inteiro. `sample/` caiu de ~1.1M para 526K.

## [4.6.0] - 2026-09-22 - Release Kansei (完成)

### Fase 6 concluída: OBI, o orquestrador de build declarativo

Todo o roadmap original (`SPEC.md` Seção 8) está agora implementado. `OBI`
(`pkg/obi` + `cmd/obi`) lê uma receita `Obifile` declarativa e invoca
`KAJI80`/`WIRTH80`/`DIGNAC` + `MUSUBI` (ou `HAKO`, quando o alvo é uma
biblioteca `.hlib`) na ordem certa — tudo **em processo**, reusando
exatamente as mesmas APIs que `cmd/kaji80`/`cmd/wirth80`/`cmd/dignac`/
`cmd/musubi`/`cmd/hako` já chamam, sem lançar subprocessos.

- `pkg/obi/parser.go`: micro-parser de linha/indentação para o subconjunto
  de sintaxe do `Obifile` (não é um parser YAML genérico — o projeto é
  deliberadamente livre de dependências externas, `go.mod` não tem nenhum
  `require`, no mesmo espírito hand-rolled do `KAJI80`/`WIRTH80`/`DIGNAC`).
  Suporta `target`/`entry`/`base` como campos escalares, `resources:`/
  `modules:` como listas de `- campo: valor`, `link:` como mapa aninhado, e
  bibliotecas via `library: {archive: x.hlib}` (forma singular, compatível
  com o `demo/Obifile` ilustrativo original) ou `libraries: [x.hlib, ...]`
  (forma plural).
- `pkg/obi/build.go`: compila cada módulo declarado, sintetiza um objeto
  `.MOB` mínimo por *resource* binário bruto (um segmento `DATA` + um
  símbolo `PUBLIC`/`DATA`, via a API já existente de `pkg/mob`), resolve
  bibliotecas `.hlib` e despacha para `musubi.LinkToFile` (`target: *.com`)
  ou `hako.Pack` (`target: *.hlib`).
- `sample/obi/`: prova real (não ilustrativa, diferente do `demo/Obifile`
  original) — `main.asm` (`KAJI80`, banco 0, dono do `Start`) chama
  `ChartLib.Desenhar`, um módulo `DIGNAC` sem `PROCEDURE Main` (logo sem
  `Start` próprio, sem conflito de símbolo) no banco 2, com um resource
  binário embutido e a `MSXLIB` via `.hlib`. `MUSUBI` gera o trampolim de
  troca de banco automaticamente.

**Achado registrado, fora de escopo**: `WIRTH80` sempre emite seu próprio
`Start` e não tem sintaxe para declarar/chamar uma rotina externa
arbitrária — por isso não pode ser um módulo "biblioteca" numa ligação
multi-módulo hoje (só pode ser o único módulo, dono do programa). É por
isso que `demo/main.pas` continua só ilustrativo.

### Regressão real encontrada e corrigida: o bootstrap multi-banco do MUSUBI não executava

Construir e testar `sample/obi` em hardware real (banco 0 + banco 2)
revelou que o programa carregava e voltava limpo ao prompt do MSX-DOS sem
executar nada — nem imprimir o `[L]` que o bootstrap sempre imprime
primeiro. O mesmo teste com o já existente `sample/multibank` confirmou:
**não era específico do OBI, era geral a qualquer programa multi-banco**.

Investigação por histórico do Git: o multi-banco funcionava desde a
v4.1.0 "Akatsuki" (quando esse suporte foi introduzido) usando um
mapeamento **identidade** simples — segmento físico da Memory Mapper =
número de banco do linker. A sessão da v4.5.2 "Yoake" trocou isso por
alocação **dinâmica** de segmento via `ALL_SEG` do EXTBIO, em teoria mais
correta (evita colidir com um segmento que o MSX-DOS 2 ou outro processo
já esteja usando), mas **nunca tinha sido executada de verdade** — só
validada por análise estática e um script Python de conferência de bytes.
Essa troca foi a regressão.

Duas tentativas de correção:

1. Confirmado contra a documentação oficial do protocolo EXTBIO
   ([map.grauw.nl/resources/dos2_environment.php](http://map.grauw.nl/resources/dos2_environment.php))
   que a rotina `ALL_SEG` exige `B` = seleção de mapper (0 = mapper
   primário) como parâmetro de entrada, além de `A` = tipo de segmento.
   `buildBootstrapCode` fazia `XOR A` (`A=0`, correto) mas nunca definia
   `B` antes de chamar `ALL_SEG` — `B` ficava com o que a chamada EXTBIO
   anterior (que busca a tabela de saltos) tivesse deixado lá, que segundo
   a mesma documentação é "o slot do mapper primário", não necessariamente
   0. Corrigido adicionando `LD B, 0`. **Reteste em hardware: continuou
   quebrado, sintoma idêntico.**
2. Com a correção de registrador não resolvendo, e a confirmação de que
   essa era uma regressão pós-v4.1.0 sem histórico de funcionamento,
   `buildBootstrapCode` foi revertido por inteiro para o mapeamento
   identidade original — removida a célula `Musubi_CallHL`/`JP (HL)`, o
   scratch da tabela de saltos do EXTBIO, o laço de chamadas `ALL_SEG` e o
   handler `[NOMEM]` (nenhum necessário sem alocação dinâmica). Mantida a
   detecção de EXTBIO/HOKVLD que ajusta `Musubi_PutP2`/`GetP2` para as
   rotinas oficiais do EXTBIO quando disponível (ortogonal à numeração de
   segmento, não implicada em nenhuma das duas falhas) e a resolução de
   saltos por endereço absoluto via `patchJP()` em vez de deslocamentos
   `JR` contados à mão (boa prática mantida da v4.5.2).

**Confirmado em hardware real**: `sample/multibank/multibank.com`
recompilado imprime corretamente `[L]`, a mensagem do Banco 0, a do Banco
1, a do Banco 2 e a mensagem final do Banco 0, com retorno limpo ao
MSX-DOS. `sample/obi/chart_lib.bas` voltou de `BANK 0` para `BANK 2` (a
versão single-bank era só uma mitigação temporária enquanto o bootstrap
estava sob suspeita), restaurando a demonstração completa de multi-banco
via OBI.

`TestBootstrapAllocSegStructure` foi reescrito como
`TestBootstrapIdentityMappingStructure` para o novo layout de bytes
(menor: sem as células/laço de `ALL_SEG`). `go test ./...` limpo.

### SCREEN 2 confirmada visualmente em hardware (v4.5.3, não documentada até agora)

A causa raiz do bug gráfico de SCREEN 2 (dois bugs independentes,
encontrados e corrigidos na v4.5.2 — ver entrada abaixo) foi confirmada
**visualmente**, pela primeira vez em toda a saga, rodando
`sample/basic/chart.bas` em hardware/emulador real: moldura, eixos,
grade e curva de pontos todos renderizados corretamente, sem nenhum
artefato. Ver `README.md` para a captura de tela e a listagem completa.
De brinde, corrigida uma não-determinismo no `.MOB` gerado pelo `KAJI80`
(a tabela de símbolos `PUBLIC`/`EXTERN` era serializada iterando um `map`
do Go, cuja ordem muda a cada execução — inofensivo para o linker, mas
gerava diffs espúrios em builds versionados a cada remontagem do mesmo
fonte; corrigido ordenando os nomes alfabeticamente antes de serializar).
Não-determinismo residual e diferente ainda existe em `WIRTH80`/`DIGNAC`
(ordem dos literais de string na pool de deduplicação), não corrigido.

## [4.5.2] - 2026-09-21 - Release Yoake (夜明け)

### Causa raiz real do bug gráfico de SCREEN 2 encontrada e corrigida

Depois da sessão de 2026-09-10 (registrada abaixo) ter investigado e
descartado timing de VRAM, atomicidade de `DI`/`EI` e o motor de comando de
hardware do V9938, sem achar a causa, esta sessão encontrou **dois bugs
independentes**, confirmados por remontagem e leitura direta dos bytes do
`.MOB` gerado (não apenas análise estática):

1. **`pkg/kaji80/assembler.go`, `encodeAlu8`/`estimateSize`**: operações ALU
   de 8 bits (`ADD`, `ADC`, `SUB`, `SBC`, `AND`, `XOR`, `OR`, `CP`) com
   operando indexado (`(IX+d)`/`(IY+d)`) não eram reconhecidas e caíam no
   ramo de imediato de 8 bits — `parseImm8("(IX+8)")` não entende essa
   sintaxe e devolve `0`, então `CP (IX+8)` virava `CP 0` (`FE 00`) em vez
   do `DD BE 08` correto, **sem erro de montagem**. O agravante: `estimateSize`
   também devolvia 2 bytes para essa forma (mesmo tamanho do Pass 2), então
   a verificação de consistência Pass1/Pass2 adicionada em 2026-09-10 não
   detectava a divergência — só verifica TAMANHO, não semântica. As únicas
   9 ocorrências dessa forma no projeto inteiro estavam todas dentro de
   `VDP_Line`/`VDP_BoxFill` (`lib/src/vdp.asm`), que por isso calculavam
   Bresenham/limites de laço sempre a partir de `0` em vez do X/Y real —
   isso também explica por que um dump de VRAM de uma sessão anterior
   parecia (erradamente) indicar corrupção do registrador `DE` entre
   chamadas de `VDP_PSet`: na verdade era `VDP_Line` desenhando uma "escada"
   de parâmetros errados, não uma linha reta.
   Corrigido adicionando detecção de operando indexado (`isIndexedOperand`)
   antes do fallback de imediato, nos três blocos de `estimateSize` e em
   `encodeAlu8`.
2. **`lib/src/vdp.asm`, `VDP_PSet_Raw`**: escrevia a máscara do pixel
   diretamente no byte do padrão (`OUT (VDP_DATA), C`), sobrescrevendo os
   outros 7 pixels da mesma linha da célula 8x8 a cada chamada — por isso
   sobrava só 1 pixel a cada 8 em qualquer `LINE`/`BOXFILL`. A Color Table já
   recebia leitura-modificação-escrita corretamente; faltava fazer o mesmo
   para o byte de padrão. Esse bug ficava mascarado pelo bug 1 (só teria
   efeito visível depois de corrigi-lo).

Ainda **não confirmado visualmente em hardware/emulador real** — próximo
passo é rodar `sample/basic/chart.bas` de novo.

### Robustez do MUSUBI (dois pontos encontrados na mesma auditoria, sem relação com o bug gráfico)

- **`copyData` / offset de arquivo com BSS**: o formato `.MOB` nunca grava
  bytes para segmentos `BSS` (só reserva `Size` no header do segmento), mas
  o linker somava esse tamanho ao endereço do próximo item mesmo assim. Em
  build multi-banco, qualquer coisa posicionada depois de um BSS na área
  comum (dispatcher/trampolins) ficaria deslocada para trás no `.COM` real
  em relação ao endereço que os relocs apontam, porque o arquivo ficava
  `Size` bytes mais curto do que o endereço pressupõe. Corrigido
  materializando BSS como zeros reais no binário final (bônus: garante
  memória zerada, que o MSX-DOS não garante por si só). Nenhum frontend
  (KAJI80/WIRTH80/DIGNAC) emite BSS hoje, então este bug nunca havia sido
  disparado na prática.
- **`buildBootstrapCode` / números de banco usados como segmento físico**:
  o bootstrap multi-banco usava o número lógico de banco do linker (1, 2,
  3…, decidido só pela ordem de descoberta dos módulos) diretamente como
  número de segmento físico da Memory Mapper na Página 2, sem checar se
  esse segmento já estava em uso pelo MSX-DOS 2 nas Páginas 0/1/3 (ou por
  outro processo). Corrigido: quando o EXTBIO está disponível, o bootstrap
  agora aloca um segmento real por banco pagineável via `ALL_SEG` (D=4,E=2
  → tabela de saltos, offset 0; calling convention confirmada contra
  `resource/MSXgl/engine/src/dos_mapper.c/h`) e guarda o mapeamento
  banco-lógico → segmento-físico em `Musubi_BankTable`, lida tanto pelo
  loop de cópia do bootstrap quanto por todo trampolim gerado por
  `buildTrampolineCode`. Sem EXTBIO, cai no mapeamento identidade de antes
  (não há allocator disponível para consultar).
  Como consequência, a antiga fórmula de tamanho do bootstrap por contagem
  manual de bytes (a mesma classe de bug já vista no KAJI80!) foi eliminada:
  `bootstrapSize` agora é *medido* chamando `buildBootstrapCode` com bancos
  fictícios (mesmos tamanhos, endereços zero), e uma verificação de
  consistência ao preencher o bootstrap real confere que o tamanho bate,
  falhando a linkagem com erro claro em vez de corromper o layout de
  memória em silêncio caso divirja no futuro. Nenhum salto no bootstrap
  usa mais deslocamento relativo (`JR`) calculado à mão — cada `JP`/`JP cc`
  é emitido com um endereço reservado e corrigido (`patchJP`) a partir da
  posição real onde os bytes acabaram.
  Novo teste `TestBootstrapAllocSegStructure` (`pkg/musubi/linker_test.go`)
  decodifica estruturalmente os bytes do bootstrap gerado para um build
  multi-banco, e uma verificação byte-a-byte manual (script Python) do
  `.COM` real de `sample/multibank` confirmou que ALL_SEG, `Musubi_BankTable`,
  trampolins e o handler de falha `[NOMEM]` batem exatamente com o
  projetado.

`go build ./...`, `go vet ./...` e `go test ./...` limpos; toolchain inteira
(`build.ps1` raiz, `lib/build.ps1`, `sample/*/build.ps1`) reconstruída sem
erro, incluindo `sample/multibank` (exercita o novo bootstrap ALL_SEG) e
`sample/basic/chart.bas` (exercita os dois bugs de VDP).

## Histórico da investigação (sessão de 2026-09-10, antes da causa raiz ser encontrada)

### Intervenções do Claude Code (2026-09-10)

Sessão de depuração profunda do problema gráfico em SCREEN 2, que levou a uma
auditoria mais ampla do assembler `KAJI80` e do linker `MUSUBI` a pedido do
usuário. Resultado: **quatro bugs reais e confirmados corrigidos** (nenhum
deles a causa raiz final do problema gráfico, que segue em aberto), mais duas
tentativas de correção do problema gráfico em si que não resolveram.

**Bugs reais corrigidos (`pkg/kaji80/assembler.go`, `pkg/musubi/linker.go`,
`lib/src/bios.asm`):**

1. `estimateLdSize` (Pass 1 do KAJI80): `LD A, (rótulo)` era subestimado em 2
   bytes (tratado como `LD A, n` imediato) em vez dos 3 bytes corretos de
   `LD A, (nn)`. Também afetava `LD A,(BC)`/`LD A,(DE)` (1 byte, não 2).
2. `estimateSize`, caso `DB`: para uma linha com rótulo antes (`rotulo: DB
   valor`), o próprio token do mnemónico `DB` era contado como um byte de
   dado extra, superestimando em +1.
3. `encodeInstruction`, caso `DB`: o mesmo bug do item 2, só que na emissão
   real dos bytes (Pass 2) — chegava a **escrever** um byte espúrio.
4. `MUSUBI`, `buildBootstrapCode`: um `JR Z` no carregador multi-banco
   calculado como `+17` (0x11) quando deveria ser `+16` (0x10) — só afeta
   programas multi-banco (`BANK 1+`); os exemplos de teste atuais são todos
   monobanco.

Qualquer um dos itens 1-3, isoladamente, corrompe silenciosamente o endereço
de **todo rótulo declarado depois no mesmo módulo** — sem erro de montagem.
Foi exatamente isso que corrompia o próprio `RET` do `VDP_PSet` (sobrescrito
pela cor a cada chamada), explicando boa parte do caos visual observado antes
desta sessão. **Rede de segurança permanente adicionada**: `Assemble()` agora
recompara, para cada linha, o tamanho estimado no Pass 1 contra os bytes
realmente emitidos no Pass 2, e falha a montagem com erro claro apontando a
linha exata em caso de divergência — pegou o bug 3 automaticamente assim que
foi ativado. Testes de regressão em `pkg/kaji80/assembler_test.go`.

Também corrigido: `BIOS_CHGET` (`lib/src/bios.asm`) só saía do laço de espera
com a tecla ESC — qualquer outra tecla caía no mesmo caminho de "sem tecla" e
o laço continuava. E: `VDP_PSet` nunca escrevia a cor do pixel na Color Table
(só o bit do padrão) — agora faz leitura-modificação-escrita do nibble de
frente, preservando o de fundo.

**O que NÃO resolveu o problema gráfico em si** (para não repetir na próxima
sessão): aumentar a margem de `NOP`s entre operações de VRAM; tornar o laço
inteiro de `VDP_Line`/`VDP_BoxFill` atômico com um único `DI`/`EI` em vez de
por-chamada; trocar `VDP_PSet` pelo motor de comando de hardware do V9938
(descartado de vez — esse motor só funciona em Graphic 4-7/SCREEN 5-8, nunca
em SCREEN 2, confirmado via documentação técnica externa). O laço de
`VDP_Line` foi revisado byte a byte contra o `.MOB` montado e está
semanticamente correto; uma chamada isolada e duas chamadas separadas a
`VDP_PSet` funcionam perfeitamente, mas o laço interno de muitos pontos
ainda produz pixels dispersos. Causa raiz não identificada nesta sessão.

### Intervenções do GitHub Copilot

- Corrigida a configuração manual do SCREEN 2 em `lib/src/vdp.asm`: `R1=E2h`
  e `R7=F1h`, mantendo `R0=02h`.
- Atualizado `sample/basic/chart.bas` para executar `Desenhar(10)` sem emitir
  `PRINT` enquanto o SCREEN 2 está ativo.
- O exemplo BASIC completo foi recompilado com sucesso e a suíte `go test ./...`
  passou.

### Estado

- O teste mínimo de SCREEN 2 e o exemplo `sample/basic/chart.bas` estão
  prontos para validação visual no OpenMSX.

### Política de Versionamento (`MAJOR.MINOR.COMPILAÇÃO`)

- **MAJOR**: Incrementado a cada encerramento de fase da toolchain (ex.: Fase 1 = MOB, Fase 2 = KAJI80, Fase 3 = MUSUBI monobanco, Fase 4 = Multi-banco & Memory Mapper).
- **MINOR**: Incrementado a cada feature ou subsistema novo adicionado.
- **COMPILAÇÃO (BUILD)**: Incrementado a cada compilação / build realizado no projeto.

## [4.5.1] - 2026-09-04 - Release Kuyashii (悔しい)

### Modificado & Corrigido

- **MSX-BASIC Dignified DIGNAC (`pkg/dignac`) e MSXLIB (`lib/src/`)**:
  - **Exportação de Símbolos**: Exportação das rotinas `VDP_PSet`, `VDP_Line`, `VDP_BoxFill` e `VDP_InitScreen2_Tables` na diretiva `PUBLIC` de `lib/src/vdp.asm`, permitindo resolução completa e correta pelo smart-linker `musubi`.
  - **Suporte de Vídeo e BIOS ([lib/src/bios.asm](lib/src/bios.asm))**:
    - Ajuste em `BIOS_CHGMOD`: integração com `INIGRP (0072h)` da Main-ROM BIOS para SCREEN 2 e rotinas de tabela VRAM (`VDP_InitScreen2_Tables`).
    - Integração com `INITXT (006Ch)` e `CHGCLR (0062h)` na saída para SCREEN 0, garantindo restauração da fonte de caracteres ROM e das cores de tela ao retornar ao MSX-DOS.
    - Temporizador em `BIOS_CHGET` com dreno de caracteres residuais e saída antecipada via tecla `ESC`.
- **Estado dos Testes Gráficos**:
  - O pipeline de montagem e linkagem do exemplo gráfico (`sample/basic/chart.bas` -> `chart.com`) compila e empacota perfeitamente de ponta a ponta.
  - A renderização em tela real/emulador sob o MSX-DOS 2 em hardware MSX2+ permanece em investigação para ajustes de inicialização de VRAM/VDP (motivo do codinome _Kuyashii (悔しい)_ — expressando o sentimento de frustração respeitosa e determinação para a próxima sessão).

---

## [4.5.0] - 2026-09-04 - Compilador MSX-BASIC Dignified DIGNAC

### Adicionado

- **Compilador MSX-BASIC Dignified DIGNAC (`pkg/dignac` e `cmd/dignac`) - Fase 5.3**:
  - Nova ferramenta da toolchain para compilação de código BASIC estruturado para MSX2+ / MSX-DOS 2, emitindo objetos relocáveis `.MOB`.
  - **Modularidade e Paginação**: Suporte a diretivas `MODULE`, `BANK`, `PUBLIC` e `EXTERN`.
  - **Sub-rotinas e ABI Kizuna**: Suporte a `PROCEDURE`, `SUB`, `FUNCTION`, passagem de parâmetros em pilha, alocação de quadro de pilha com frame pointer `IX` e variáveis locais (`LOCAL x%, y%`).
  - **Ponto de Entrada Automático**: Geração automática do ponto de entrada `Start` chamando `PROCEDURE Main` e finalizando com `BDOS_Exit`, permitindo linkagem direta de `.COM` standalone.
  - **Controle de Fluxo Estruturado**: Laços `FOR ... TO ... STEP ... NEXT`, `WHILE ... WEND`, `DO ... LOOP`, e condicionais `IF ... THEN ... ELSE ... END IF` (em linha única ou blocos).
  - **Primitivas Gráficas**: Comandos `LINE (x1,y1)-(x2,y2)[, color][, B | BF]` e `PSET (x, y)[, color]`.
  - **Expressões e Aritmética de 16 bits**: Operações `+`, `-`, `*` (via `Mul16`), `/` e operador `MOD` (via `Div16`), além de operadores lógicos `AND`, `OR`, `XOR`, `NOT` e comparações relacionais (`=`, `<>`, `<`, `<=`, `>`, `>=`).
  - **CLI `dignac`**: Suporte às flags `-o <saida.mob>`, `-S` (código assembly intermediário formatado), `-v` (modo detalhado) e `--version`.
- **Endereçamento Indexado Z80 no Assembler KAJI80 ([pkg/kaji80/assembler.go](pkg/kaji80/assembler.go))**:
  - Suporte completo a instruções `LD r, (IX+d)` e `LD (IX+d), r` para registradores de 8 bits.
  - Suporte a `LD (IX+d), n` (imediato).
  - Suporte análogo para o registrador de índice `IY` (`0xFD`).
  - Suporte a `LD SP, IX` (`DD F9`) e `LD SP, IY` (`FD F9`).
- **Primitivas Gráficas na Biblioteca Padrão MSXLIB ([lib/src/vdp.asm](lib/src/vdp.asm))**:
  - `VDP_PSet`: Cálculo de endereços nas tabelas de padrão e cor em SCREEN 2 (TMS9918/V9938) e plotagem de pixel.
  - `VDP_BoxFill`: Preenchimento de retângulos e aceleração para limpeza de tela cheia via `VDP_FillVRAM`.
  - `VDP_Line`: Traçado de segmentos de reta horizontais e diagonais.
- **Exemplos em BASIC Dignified (`sample/basic/`)**:
  - `hello.bas`: Hello World em BASIC compilado (apenas 186 bytes).
  - `calc.bas`: Demonstração de variáveis locais, aritmética de 16 bits, divisão, módulo e exibição decimal.
  - `chart.bas`: Módulo paginável no banco 2 para traçado de curvas.
  - `build.ps1`: Automação de compilação e smart-linking dos exemplos BASIC.
- **Empacotamento Global**:
  - `dignac.exe` integrado ao script mestre [build.ps1](build.ps1) e empacotado em `distribute/bin/`.
  - Pasta `sample/basic/` incluída na distribuição oficial.

---

## [4.4.0] - 2026-09-04 - Compilador Pascal WIRTH80

### Adicionado

- **Compilador Pascal WIRTH80 (`pkg/wirth80` e `cmd/wirth80`)**:
  - Nova ferramenta da toolchain para compilação de código-fonte Pascal nativo para MSX2+ / MSX-DOS 2, gerando módulos objeto no formato `.MOB`.
  - **Análise Léxica e Sintática**: Suporte a identificadores, números (decimais e `$hex`), strings literais (`'...'`), comentários (`{ ... }`, `(* ... *)`, `// ...`), `program`, `var`, `begin ... end.`.
  - **Tipos de Dados**: Suporte a `Integer` (16 bits sinalizado) e `Char` / `Boolean` (8 bits).
  - **Comandos e Controle de Fluxo**: Atribuição (`:=`), `Write`, `WriteLn`, condicionais `if ... then ... else` e laços `while ... do`.
  - **Expressões**: Aritmética de 16 bits (`+`, `-`, `*` via `Mul16`, `div` via `Div16`), unários e comparações relacionais (`=`, `<>`, `<`, `<=`, `>`, `>=`).
  - **Integração com MSXLIB**: Geração automática de chamadas externas para `BDOS_PrintString`, `BDOS_PrintChar`, `PrintDec16`, `Mul16`, `Div16` e `BDOS_Exit`.
  - **CLI `wirth80`**: Suporte às flags `-o <saida.mob>`, `-S` (emissão de código assembly Z80 legível), `-v` (modo detalhado) e `--version`.
- **Exemplos em Pascal (`sample/pascal/`)**:
  - `hello.pas`: Exemplo clássico Hello World em Pascal para MSX.
  - `calc.pas`: Demonstração de declaração de variáveis, multiplicação de 16 bits e exibição de números decimais.
  - `build.ps1`: Script de automação para compilação e smart-linking dos exemplos Pascal.
- **Empacotamento Global**:
  - `wirth80.exe` integrado ao script mestre [build.ps1](build.ps1) e empacotado em `distribute/bin/`.
  - Pasta `sample/pascal/` incluída na distribuição oficial.

---

## [4.3.3] - 2026-09-04

### Corrigido

- **Suporte Aritmético de 16 bits no Assembler KAJI80 ([pkg/kaji80/assembler.go](pkg/kaji80/assembler.go))**:
  - Implementada a codificação e estimativa precisa de `SBC HL, ss` (`ED 42/52/62/72`) e `ADC HL, ss` (`ED 4A/5A/6A/7A`) de 16 bits.
  - Anteriormente, `SBC HL, DE` caía no caso ALU de 8 bits e montava incorretamente como `0xDE 0x00` (`SBC A, 00h`). Isso fazia o loop de cálculo de dígitos em `PrintDecDigit` entrar em um loop infinito no Z80 sem subtrair HL, travando a máquina antes de imprimir qualquer caractere e impedindo a execução de prosseguir para o som e retorno ao DOS.
  - Implementado suporte a `LD HL, (nn)` (`0x2A`) e `LD (nn), HL` (`0x22`) no KAJI80.
- **Ajuste de Duração do Loop de Pausa de Áudio ([sample/libdemo/main.asm](sample/libdemo/main.asm))**:
  - Reduzido o contador do loop externo `LD B, 18h` (que durava ~11.5 segundos) para `LD B, 02h` (~1 segundo audível), tornando a execução e retorno ao DOS fluídos e imediatos.
- **Validação Completa via Emulação de CPU**:
  - Execução validada instrução-a-instrução em simulador Z80, comprovando o cálculo de `123 x 45 = 5535`, ativação correta do PSG e encerramento com sucesso via BDOS 00h.

---

## [4.3.2] - 2026-09-04

### Corrigido

- **Sincronização de Offsets no Assembler KAJI80 ([pkg/kaji80/assembler.go](pkg/kaji80/assembler.go))**:
  - Corrigida a estimativa de tamanho no Pass 1 para instruções ALU imediatas de 1 operando (ex: `CP n`, `SUB n`, `AND n`), que retornava 1 byte em vez de 2 bytes.
  - Corrigida a estimativa de tamanho para `LD (DE), A`, `LD (BC), A`, `LD A, (DE)` e `LD A, (BC)` que retornava 3 bytes em vez de 1 byte.
  - A discrepância acumulada causava desvio de endereço nos símbolos de `string.mob`, fazendo o linker `musubi` calcular chamadas como `PrintDec16` com desvio de 2 a 3 bytes no meio de outra instrução.
- **Canal de Som do PSG ([lib/src/psg.asm](lib/src/psg.asm))**:
  - Inserido `LD A, C` após `INC C` no registrador grosso de período da frequência.
  - Configuração exata do misturador do Registrador 7: ativa Canal A (`0xBE`), Canal B (`0xBD`) e Canal C (`0xBB`) sem afetar as portas de joystick e teclado.
- **Pausa Audível e Frequência no Demo ([sample/libdemo/main.asm](sample/libdemo/main.asm))**:
  - Ajustada a frequência para A440 (`0x00FE`), volume máximo 15 (`0x0F`) e loop de pausa aninhado audível (~0.5s).

---

## [4.3.1] - 2026-09-03

### Corrigido

- **Conversão Decimal em `PrintDec16` ([lib/src/string.asm](lib/src/string.asm))**:
  - Reescrita da rotina de conversão para utilizar subtração sucessiva direta de potências de 10 sem endereçamento indexado `(IX+d)`, eliminando falha que impedia a exibição do resultado numérico.
- **Preservação de Registradores em Chamadas BDOS ([lib/src/bdos.asm](lib/src/bdos.asm))**:
  - `BDOS_PrintChar` e `BDOS_PrintString` agora salvam e restauram explicitamente todos os registradores (`AF`, `BC`, `DE`, `HL`, `IX`, `IY`), prevenindo corrupção de registradores durante chamadas de console do MSX-DOS.
- **Controle Seguro do Misturador PSG ([lib/src/psg.asm](lib/src/psg.asm))**:
  - `PSG_PlayTone` e `PSG_MuteAll` agora gravam valores fixos e seguros no Registrador 7 (`0xB8` / `0xBF`), eliminando leituras na porta `0xA2` e preservando intactas as direções das portas de I/O (joystick e teclado) do MSX.
- **Lexer do KAJI80 ([pkg/kaji80/lexer.go](pkg/kaji80/lexer.go))**:
  - Ajustada a precedência na análise de literais para verificar o sufixo `H` antes do prefixo `0b`, corrigindo erro de montagem em constantes hexadecimais iniciadas em `0B` (ex: `0BFh`, `0B8h`).

---

## [4.3.0] - 2026-09-03 - Biblioteca Padrão MSXLIB

### Adicionado

- **Biblioteca Padrão MSXLIB (`lib/src/` e `lib/msxlib.hlib`)**:
  - Pacote de 6 módulos modulares escritos em Z80 Assembly e empacotados em arquivo de biblioteca `.HLIB`:
    - `bdos`: Rotinas de chamada ao kernel MSX-DOS (`BDOS_Call`, `BDOS_PrintChar`, `BDOS_PrintString`, `BDOS_ReadChar`, `BDOS_Exit`).
    - `bios`: Chamadas inter-slot seguras à Main-ROM via `CALSLT` (`BIOS_Call`, `BIOS_CHPUT`, `BIOS_CHGET`, `BIOS_CLS`, `BIOS_POSIT`, `BIOS_BEEP`, `BIOS_INIT32`).
    - `vdp`: Controle de portas e registradores do processador de vídeo TMS9918/V9938/V9958 (`VDP_WriteReg`, `VDP_SetWriteAddr`, `VDP_SetReadAddr`, `VDP_FillVRAM`, `VDP_WriteVRAM`, `VDP_ReadVRAM`, `VDP_SetColor`).
    - `psg`: Manipulação do gerador de som programável AY-3-8910 (`PSG_Write`, `PSG_Read`, `PSG_MuteAll`, `PSG_PlayTone`).
    - `string`: Funções de texto e conversão numérica rápida com supressão de zeros (`StrLen`, `StrCopy`, `StrToUpper`, `PrintHex8`, `PrintHex16`, `PrintDec16`).
    - `math`: Multiplicação e divisão inteira não sinalizada de 16 bits (`Mul16`, `Div16`).
- **Suporte a Novas Instruções Z80 no Assembler KAJI80 (`pkg/kaji80`)**:
  - Instruções de rotação e deslocamento: `RLCA`, `RRCA`, `RLA`, `RRA`, `RLC`, `RRC`, `RL`, `RR`, `SLA`, `SRA`, `SRL`.
  - Operações bit-a-bit: `BIT`, `RES`, `SET`.
  - Flags e complemento aritmético: `CPL`, `SCF`, `CCF`, `NEG`.
- **Exemplo Prático com Smart-Linking ([sample/libdemo](sample/libdemo))**:
  - Exemplo demonstrativo em Assembly importando e executando rotinas da `msxlib.hlib`, comprovando eliminação de código morto no mapa de memória.
- **Integração no Script de Distribuição ([build.ps1](build.ps1))**:
  - Geração automática de `msxlib.hlib` e inclusão em `distribute/lib/` e no pacote ZIP de distribuição.

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
  - Eliminação de código morto (_dead-code elimination_): módulos não referenciados no `.hlib` são descartados, mantendo o executável `.com` no menor tamanho possível.
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
  - Montador de dois passos (_two-pass assembler_) emitindo arquivos de objeto `.MOB`.
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

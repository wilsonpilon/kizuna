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

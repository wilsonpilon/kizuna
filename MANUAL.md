# KIZUNA — Manual (rascunho conceitual, v0.0.0)

> Este manual descreve o uso **pretendido** da toolchain. Nenhuma das
> ferramentas abaixo está implementada nesta versão — serve como guia de
> referência para a implementação futura.

## 1. Visão geral do fluxo

```
fonte.asm  ──KAJI80──►  fonte.mob  ─┐
fonte.pas  ──WIRTH80─►  fonte.mob  ─┼──MUSUBI──►  programa.com
fonte.bas  ──DIGNAC──►  fonte.mob  ─┘
```

Um projeto KIZUNA é descrito por um arquivo `Obifile` na raiz, que lista
os módulos-fonte, o compilador de cada um, o banco de memória alvo e o
tamanho esperado. O comando `obi build` lê essa receita, invoca cada
compilador na ordem correta e chama `MUSUBI` para gerar o `.COM` final.

## 2. Estrutura de um `Obifile`

```yaml
target: programa.com
entry: NomeDoModulo.PontoDeEntrada

resources:
  - file: recurso.bin
    bank: N
    size: TAMANHO
    desc: "descrição livre"

modules:
  - name: NomeDoModulo
    source: arquivo-fonte
    compiler: kaji80 | wirth80 | dignac
    bank: N        # 0 = area comum fixa; 1..N = banco paginavel
    size: TAMANHO

link:
  tool: musubi
  map: nome.map     # relatorio de enderecos + trampolins gerados
  trampoline: auto   # MUSUBI decide sozinho onde precisa

library:
  tool: hako
  archive: runtime.hlib
```

## 3. Comandos

| Comando       | O que faz                                                  |
|---------------|--------------------------------------------------------------|
| `obi build`   | Compila todos os módulos e linka o `.COM` final              |
| `obi clean`   | Remove artefatos intermediários (`.mob`, `.map`)              |
| `obi map`     | Mostra o mapa de memória/bancos resultante do último build   |
| `kaji80 arq.asm`  | Compila um módulo Assembly isoladamente para `.mob`       |
| `wirth80 arq.pas` | Compila um módulo/unit Pascal isoladamente para `.mob`    |
| `dignac arq.bas`  | Compila um módulo BASIC Dignified isoladamente para `.mob`|
| `musubi *.mob -o saida.com` | Linka objetos `.mob` diretamente, sem `Obifile`  |
| `hako pack runtime.hlib arq1.mob arq2.mob` | Empacota objetos numa biblioteca |

## 4. Interoperabilidade entre linguagens

- Um módulo Pascal declara uma dependência externa com `{$USES Nome in
  'arquivo'}`, apontando para um módulo Assembly ou BASIC Dignified.
- Símbolos exportados (`PUBLIC` no Assembly, `interface` no Pascal,
  `PUBLIC PROCEDURE` no BASIC Dignified) ficam visíveis para os demais
  módulos linkados.
- Chamadas entre módulos em bancos diferentes são resolvidas
  automaticamente por `MUSUBI` via trampolim (transparente para quem
  escreve o código-fonte).
- Strings passadas entre linguagens seguem sempre o formato short string
  (1 byte de tamanho + dados) — não há conversão necessária nas
  fronteiras.

## 5. Exemplo mínimo

Veja o diretório `demo/` para um croqui completo: um programa Pascal que
chama uma rotina de ajuste de tela em Assembly, lê uma entrada do
usuário, chama um módulo BASIC Dignified para desenhar um gráfico, e
retorna ao Pascal — com os três módulos em bancos de memória diferentes.

## 6. Ver também

- `SPEC.md` — especificação técnica completa (formato `.MOB`, ABI,
  modelo de bancos de memória).

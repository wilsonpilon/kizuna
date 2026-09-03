# KIZUNA (絆)

> "Um laço entre linguagens, amarrado num único `.COM`."

**KIZUNA** é uma proposta de toolchain para MSX2+ (Z80 / MSX-DOS 2 / 256Kb
ou mais com memory mapper) que permite escrever partes de um mesmo programa em
**Assembly Z80**, **Pascal** (estilo Turbo Pascal 4, com units) e
**MSX-BASIC Dignified**, compilar cada parte separadamente e linkar tudo
num único executável — inclusive distribuindo módulos por bancos de
memória diferentes, com troca de banco resolvida automaticamente pelo
linker.

Este repositório é, por enquanto, **conceitual**: contém a especificação
técnica completa e um croqui de demonstração, sem implementação funcional.
Versão `0.0.0`.

## Por quê

MSX-DOS 2 herda compatibilidade de BDOS do CP/M-80, plataforma onde
linguagens como Macro-80, Fortran-80 e Cobol-80 já linkavam juntas via um
formato de objeto comum. O Turbo Pascal 3 (até a 3.3) rodava nesse mesmo
mundo, mas **não permitia linkar com outras linguagens**; foi o Turbo
Pascal 4, com suas units e a diretiva `{$L}` para `.OBJ` externos, que
passou a **permitir** essa integração — só que ficou restrito a
MS-DOS/x86. O KIZUNA propõe recuperar essa ideia — um Pascal fiel ao TP4, com units —
mas nativo para MSX2+, e indo além: aproveitando os 256Kb (ou mais) via bank
switching, coisa que os compiladores originais de CP/M nunca precisaram
resolver, e permitindo combinar isso com Assembly e um BASIC estruturado
no mesmo binário.

## As ferramentas

| Nome        | O que faz                              |
|-------------|------------------------------------------|
| `KAJI80`    | Assembler Z80                            |
| `WIRTH80`   | Compilador Pascal (estilo Turbo Pascal 4) |
| `DIGNAC`    | Compilador do MSX-BASIC Dignified         |
| `MUSUBI`    | Linker (resolve símbolos e bancos)        |
| `HAKO`      | Bibliotecário / empacotador de objetos    |
| `OBI`       | Orquestrador de build (lê o `Obifile`)    |

Cada compilador gera um objeto relocável no formato próprio `.MOB`;
`MUSUBI` linka tudo (com trampolins automáticos de bank switching quando
necessário) e produz um `.COM` MSX-DOS 2.

## Estrutura deste repositório

```
SPEC.md            Especificação técnica completa
README.md          Este arquivo
MANUAL.md           Rascunho de manual de uso (conceitual)
CHANGELOG.md        Histórico de versões
RELEASE_NOTES.md    Notas desta release
demo/               Croqui de exemplo: Pascal + Assembly + BASIC Dignified
  main.pas           Programa principal (WIRTH80)
  screen.asm          Rotina de ajuste de tela (KAJI80)
  chart.bas           Módulo gráfico (DIGNAC)
  Obifile             Receita de build de exemplo
```

## Status

Nada aqui compila ainda — é a planta baixa do projeto, documentada para
servir de referência antes de começar a implementação em Go. Veja
`SPEC.md` para os detalhes técnicos e `demo/` para uma visão de como um
programa real usaria o toolchain depois de pronto.

## Licença

A definir.

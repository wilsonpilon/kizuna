# KIZUNA — Especificação Técnica (v0.0.0 / conceito)

**KIZUNA** (絆, "laço", "vínculo") é uma proposta de toolchain multi-linguagem
para MSX2+ (Z80, MSX-DOS 2, 256Kb ou mais de RAM com memory mapper), permitindo escrever
um único programa combinando **Assembly Z80**, **Pascal** (estilo Turbo
Pascal 4) e **MSX-BASIC Dignified**, compilando cada módulo separadamente e
linkando tudo num único `.COM` final — inclusive através de bancos de memória
diferentes.

Este documento registra as decisões de design discutidas até aqui. Nada
aqui está implementado; é a planta baixa do projeto.

---

## 1. Motivação

- MSX-DOS 2 é compatível a nível de BDOS com CP/M-80, o que já permitia,
  historicamente, rodar `.COM` gerados por toolchains cruzadas de linguagens
  diferentes (Macro-80, Fortran-80, Cobol-80 linkavam juntos via formato
  REL da Microsoft).
- Turbo Pascal 3 (até a versão 3.3) rodava em CP/M-80 (Z80), mas **não
  permitia linkar com outras linguagens** — não havia compilação separada
  nem suporte a objetos externos. O Turbo Pascal 4 introduziu units
  (compilação separada) e a diretiva `{$L}` para linkar `.OBJ` externos,
  passando a **permitir** essa integração entre linguagens — mas ficou
  restrito a MS-DOS/x86. Não existe hoje um Pascal moderno, fiel ao TP4
  (com units), gerando `.COM` nativo para MSX2+.
- O objetivo final é uma linguagem melhor que o TP3/TP4 originais **porque**
  passa a usar os 256Kb (ou mais) de RAM com bank switching — algo que os compiladores
  originais de CP/M nunca precisaram resolver.
- Como ponto de partida mais rápido de validar o pipeline completo, antes do
  Pascal, avaliou-se também um dialeto Forth como "linguagem mais rápida de
  entregar resultado" em Z80 (não faz parte do escopo desta v0.0.0, mas fica
  registrado como direção futura).

## 2. Escopo da linguagem Pascal

- Fiel ao **Turbo Pascal 4**: units com seção `interface`/`implementation`,
  compilação separada, `uses`.
- Runtime (`System`) provê I/O de arquivo via BDOS (compatível CP/M),
  console via BIOS, heap (`New`/`Dispose`).
- Uma unit adicional (`CRT`/`Graph`-like) explora BIOS/VRAM do MSX2+ para
  gráficos, algo que o TP3/TP4 originais nunca tiveram.

## 3. Plataforma alvo

| Item              | Valor                                             |
|-------------------|----------------------------------------------------|
| Máquina            | MSX2+                                               |
| CPU               | Z80                                                 |
| Sistema            | MSX-DOS 2                                           |
| Memória            | 256Kb RAM ou mais, com memory mapper                |
| Formato de saída  | `.COM` executável                                   |
| Paginação          | 4 páginas de 16Kb (`0000h-3FFFh` .. `C000h-FFFFh`)  |

### Mapa de memória proposto

```
0000h-3FFFh  Página 0  Fixa — reservada para MSX-DOS 2/BIOS
4000h-7FFFh  Página 1  Fixa — runtime comum + código sempre presente
8000h-BFFFh  Página 2  Janela comutável — bancos pagináveis entram aqui
C000h-FFFFh  Página 3  Fixa — pilha, heap, buffers
```

Módulos grandes (units Pascal, módulos BASIC Dignified, blocos de Assembly)
são alocados em bancos pagináveis mapeados na página 2. O linker decide a
alocação de cada módulo em um banco e resolve chamadas entre bancos
automaticamente via trampolim (ver §6).

## 4. Componentes da toolchain

Nomeados no mesmo espírito "Japão retrô-futurista anos 80" da identidade
visual já usada no projeto irmão `msxbasica`.

| Ferramenta | Papel                                 | Origem do nome                          |
|------------|----------------------------------------|------------------------------------------|
| **KAJI80** | Assembler Z80                          | *kaji* (鍛冶) — ferreiro, forja o código |
| **WIRTH80**| Compilador Pascal (TP4-like)           | Niklaus Wirth                            |
| **DIGNAC** | Compilador do MSX-BASIC Dignified      | "Dignified" + "-ac" (compilador)         |
| **MUSUBI** | Linker                                 | *musubi* (結び) — atar, amarrar          |
| **HAKO**   | Bibliotecário / empacotador de objetos | *hako* (箱) — caixa                      |
| **OBI**    | Orquestrador de build (arquivo `Obifile`) | *obi* (帯) — faixa que amarra o conjunto |

Cada frontend (`KAJI80`, `WIRTH80`, `DIGNAC`) gera um arquivo objeto no
formato `.MOB` (ver §5). `MUSUBI` resolve símbolos entre módulos e gera o
`.COM` final. `HAKO` empacota bibliotecas de runtime reutilizáveis
(`.hlib`). `OBI` lê a receita declarativa (`Obifile`) e invoca as
ferramentas na ordem certa.

**Decisão registrada:** o formato de objeto é **novo, próprio do projeto**
(não reaproveita o formato REL/M80 já usado pelo `msxbasica`), para poder
já nascer com o conceito de banco de memória embutido.

## 5. Formato de objeto `.MOB`

```
Header
  0   4   Magic "MOB1"
  4   1   Versão do formato
  5   2   Nº de segmentos
  7   2   Nº de símbolos
  9   2   Nº de relocations
  11  2   Offset da tabela de strings (nomes de símbolos)

Tabela de segmentos (um por trecho de código/dado)
  Tipo: CODE | DATA | BSS
  Banco: 0 = área comum (fixa), 1..N = banco paginável
  Tamanho
  Bytes brutos (CODE/DATA) — BSS só reserva espaço

Tabela de símbolos
  Nome (índice na tabela de strings)
  Classe: PUBLIC (exportado) | EXTERN (importado)
  Kind: PROC | DATA   ← usado pelo linker para decidir se precisa trampolim
  Segmento + offset (se PUBLIC)

Tabela de relocations
  Segmento + offset onde aplicar
  Símbolo alvo
  Tipo: ABS16 | REL8 | BANKNUM
```

Regras:
- `REL8` (usado por `JR`/`DJNZ`) só é permitido dentro do mesmo
  segmento/banco.
- Chamadas cross-bank nunca usam salto relativo; sempre viram
  `CALL`/`JP` absoluto por trampolim.

## 6. Trampolim de bank switching

Gerado automaticamente por `MUSUBI` para todo símbolo `PUBLIC` cujo
chamador está em banco diferente do chamado. Fica alocado na área comum
(página 1).

```asm
CALL_simbolo:
    push af
    in   a,(MAPPER_PAGE2)   ; salva banco atual da pagina 2
    push af
    ld   a, N               ; banco onde o simbolo mora
    out  (MAPPER_PAGE2), a
    call real_simbolo       ; endereco dentro de 8000h-BFFFh
    pop  af
    out  (MAPPER_PAGE2), a  ; restaura banco anterior
    pop  af
    ret
```

Se chamador e chamado estão no mesmo banco, `MUSUBI` emite `CALL` direto,
sem trampolim (otimização automática).

## 7. ABI própria — convenção de chamada

**Decisão registrada:** ABI própria, tudo via pilha (não replica a
convenção de registradores do Turbo Pascal original).

- **Parâmetros**: empilhados esquerda → direita (primeiro parâmetro fica
  mais fundo na pilha).
- **Frame pointer**: `IX`.
  ```asm
  push ix
  ld   ix, 0
  add  ix, sp
  ```
- **Limpeza da pilha**: feita pelo **chamador**, logo após o `CALL`
  (mais barato em Z80 do que fazer o callee limpar, já que não existe
  `RET n`):
  ```asm
  ld   hl, N        ; N = bytes dos parametros empilhados
  add  hl, sp
  ld   sp, hl
  ```
- **Retorno**: `A` para byte/boolean, `HL` para word/ponteiro. Tipos
  maiores (records, strings longas) retornam via parâmetro oculto: o
  chamador aloca o espaço e passa o endereço como argumento extra.
- **Strings entre linguagens**: formato neutro comum — 1 byte de tamanho
  + até 255 bytes de dados (short string clássico). Assembly, Pascal e
  BASIC Dignified leem/escrevem esse mesmo layout, sem conversão nas
  fronteiras.

## 8. Ordem de implementação sugerida

1. Structs do formato `.MOB` (leitura/escrita) — base de tudo.
2. `KAJI80` mínimo: assembler Z80 com subconjunto de mnemônicos, já
   emitindo `.MOB`, validando o pipeline ponta a ponta.
3. `MUSUBI` resolvendo um único banco (sem trampolim) → gera `.COM` simples.
4. Suporte a múltiplos bancos + geração automática de trampolins.
5. `WIRTH80` (Pascal/TP4-like) e `DIGNAC` (BASIC Dignified) como frontends
   adicionais emitindo o mesmo `.MOB`.
6. `HAKO` e bibliotecas de runtime comuns.
7. `OBI` como orquestrador declarativo (`Obifile`).

## 9. Fora de escopo nesta versão (ideias registradas para o futuro)

- Dialeto Forth como linguagem adicional (mais rápida de bootstrapar em
  Z80 por não exigir parser/AST tradicional).
- Interpretador Lisp simples.
- Um "BASIC compilado" mais amplo, além do subset do Dignified.
- COMAL (dialeto estruturado de BASIC com porte histórico real para MSX)
  como alternativa de linguagem de entrada mais acessível que Forth/Lisp.

## 10. Ver também

- `demo/` — croqui de um programa combinando as três linguagens, e um
  `Obifile` de exemplo mostrando a receita de build (não compila —
  é só ilustrativo do formato final).

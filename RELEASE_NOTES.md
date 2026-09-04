# Release Notes — KIZUNA v4.4.0 "Hinode" (日の出)

**Hinode** (日の出) — "o nascer do sol, a alvorada radiante". Seguindo a transição poética iniciada em *Akatsuki* (a aurora antes do amanhecer), *Hinode* representa o sol surgindo plenamente no horizonte, iluminando a toolchain com o nascimento de sua primeira linguagem de alto nível: o compilador Pascal **WIRTH80**, integrando-se nativamente com a biblioteca padrão **MSXLIB** e o linker **MUSUBI**.

---

## O que é esta release

Esta é a release oficial **v4.4.0 pre-alpha**, trazendo uma toolchain Z80 completa para MSX2+ / MSX-DOS 2 com suporte à compilação de código **Assembly Z80** e **Pascal (Turbo Pascal 4 style)**, geração de objetos relocáveis `.MOB`, criação de bibliotecas `.HLIB`, eliminação de código morto (Smart-Linking) e suporte à Memory Mapper do MSX-DOS 2.

---

## Destaques da Release

1. **WIRTH80 (Compilador Pascal Z80)**:
   - Compila programas Pascal nativamente para módulos `.MOB`.
   - Suporte a `program`, declaração de variáveis `Integer`, `Char` e `Boolean`, blocos `begin ... end.`.
   - Atribuição `:=`, `Write`, `WriteLn`, condicionais `if ... then ... else` e laços `while ... do`.
   - Expressões aritméticas de 16 bits (`+`, `-`, `*`, `div`) e comparações relacionais (`=`, `<>`, `<`, `<=`, `>`, `>=`).
   - Opção `-S` para inspecionar o código Assembly Z80 gerado pelo compilador.

2. **HAKO (Bibliotecário de Objetos)**:
   - Empacota múltiplos módulos `.MOB` em arquivos de biblioteca estática `.HLIB`.
   - Tabela global de símbolos com validação anti-duplicação.

3. **MSXLIB (Biblioteca Padrão MSX)**:
   - Coleção de 6 módulos modulares e reutilizáveis (`bdos`, `bios`, `vdp`, `psg`, `string`, `math 16-bit`).
   - Multiplicação e divisão inteira de 16 bits rápida (`Mul16`, `Div16`).
   - Formatação e conversão de números para decimal com supressão de zeros (`PrintDec16`).
   - Controle do gerador de som PSG (AY-3-8910) e do processador de vídeo VDP.

4. **MUSUBI com Smart-Linking & Multi-Banco**:
   - Resolução inteligente de dependências: extrai da biblioteca apenas o código que seu programa realmente utiliza, reduzindo drasticamente o tamanho do `.COM` final.
   - Suporte a paginação de memória na Página 2 (`0x8000..0xBFFF`) com chaveamento e trampolins automáticos via porta `0xFE` do MSX-DOS 2.

5. **Instalador TUI**:
   - Utilitário interativo `install.exe` e atalho `install.cmd` para instalação com 2 cliques e configuração automática do `PATH` no Windows.

---

## Conteúdo do Pacote de Distribuição (`kizuna-v4.4.0-dist.zip`)

```
distribute/
  ├── bin/
  │    ├── kaji80.exe     (Assembler Z80)
  │    ├── wirth80.exe    (Compilador Pascal)
  │    ├── musubi.exe     (Linker)
  │    ├── hako.exe       (Bibliotecário)
  │    └── mobdump.exe    (Inspecionador de objetos)
  ├── lib/
  │    └── msxlib.hlib    (Biblioteca Padrão MSX)
  ├── sample/
  │    ├── hello.asm      (Exemplo Assembly monobanco)
  │    ├── multibank/     (Exemplo Assembly multi-banco com 3 módulos)
  │    ├── libdemo/       (Exemplo Assembly consumindo MSXLIB)
  │    └── pascal/        (Exemplos em Pascal: hello.pas e calc.pas)
  ├── docs/
  │    ├── README.md
  │    ├── HELP.md
  │    └── CHANGELOG.md
  ├── install.exe / install.cmd
  └── LICENSE
```

---

## Próximos Passos

- **Fase 5.3**: Início do compilador **DIGNAC** (MSX-BASIC Dignified para Z80).
- **Fase 6**: Orquestrador de build declarativo **OBI** (`Obifile`).


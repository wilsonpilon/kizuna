# Créditos e referências

O KIZUNA é uma toolchain escrita do zero (Go + Assembly Z80), licenciada sob a
GPL-3.0 (ver `LICENSE`).

## Projetos usados como guia de cobertura da MSXLIB

Ao planejar a expansão da MSXLIB (ver `docs/plano-expansao-msxlib.md`), dois
projetos de código aberto foram consultados **apenas para saber o que uma
biblioteca de MSX completa costuma oferecer** — que funções existem, o que cada
uma faz, que casos de borda os autores consideraram:

- **MSXgl** — Guillaume "Aoineko" Blanchard e colaboradores.
  https://github.com/aoineko-fr/MSXgl — licença CC BY-SA 4.0.
- **Fusion-C** (versão 1.3) — Eric Boez (EBSOFT) e colaboradores.
  Licença CC BY-SA 4.0.

Nenhum código-fonte, assembly ou estrutura interna desses projetos é copiado ou
convertido para a MSXLIB: as rotinas do KIZUNA são escritas a partir da
documentação dos chips e do sistema (ver abaixo), com nomes e APIs próprios. A
lista de nomes consultada e o estado de cada item ficam em `docs/cobertura.md`
(gerado por `go run ./tools/cobertura`); as cópias desses projetos em `resource/`
são material de consulta e **não fazem parte** do build nem da distribuição.

## Documentação técnica usada como fonte das rotinas

- Folhas de dados: TMS9918A, Yamaha V9938 / V9958 / V9990, General Instrument
  AY-3-8910 e Yamaha YM2149, Yamaha YM2413 (OPLL) e Y8950 (MSX-AUDIO), Konami SCC,
  Ricoh RP-5C01 (relógio).
- Mapa de portas de I/O e tabelas de BIOS/variáveis de sistema do MSX
  (MSX Assembly Page — map.grauw.nl; MSX Resource Center — msx.org).
- Chamadas do MSX-DOS 2 (`map.grauw.nl/resources/dos2_functioncalls.php`).

## Outros

- Simulador Z80 usado nos testes: `github.com/remogatto/z80` (Andrea Fazzi, MIT),
  dependência apenas de teste (`pkg/z80sim`).
- A ideia de vários recursos do montador KAJI80 (`INCBIN`, `REPT`, macros,
  rótulos locais, `CALLBIOS`…) veio do asMSX (https://github.com/Fubukimaru/asMSX);
  a sintaxe e a implementação são próprias.

; =============================================================================
; KIZUNA MSXLIB - psg/playnote
; tabela de notas + toca nota por indice
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE psg_playnote
BANK 0

PUBLIC PSG_PlayNoteIndexed
EXTERN PSG_PlayTone

; -----------------------------------------------------------------------------
; PSG_NoteTable: períodos de 12 bits (word) para 5 oitavas (C2..B6, índices
; 0..59, 12 notas por oitava começando em Dó). Calculados pela fórmula padrão
; período = (clock do PSG / 32) / frequência, clock = 3.579.545 Hz (NTSC) --
; conferido contra o valor publicamente conhecido de A4=440Hz -> período 254
; (0FEh), índice 33 (oitava 4, nota 9 = Lá).
; -----------------------------------------------------------------------------
PSG_NoteTable:
    DW 06AEh, 064Eh, 05F4h, 059Eh, 054Dh, 0501h, 04B9h, 0475h, 0435h, 03F9h, 03C0h, 038Ah  ; C2 C#2 D2 D#2 E2 F2 F#2 G2 G#2 A2 A#2 B2
    DW 0357h, 0327h, 02FAh, 02CFh, 02A7h, 0281h, 025Dh, 023Bh, 021Bh, 01FCh, 01E0h, 01C5h  ; C3 C#3 D3 D#3 E3 F3 F#3 G3 G#3 A3 A#3 B3
    DW 01ACh, 0194h, 017Dh, 0168h, 0153h, 0140h, 012Eh, 011Dh, 010Dh, 00FEh, 00F0h, 00E2h  ; C4 C#4 D4 D#4 E4 F4 F#4 G4 G#4 A4 A#4 B4
    DW 00D6h, 00CAh, 00BEh, 00B4h, 00AAh, 00A0h, 0097h, 008Fh, 0087h, 007Fh, 0078h, 0071h  ; C5 C#5 D5 D#5 E5 F5 F#5 G5 G#5 A5 A#5 B5
    DW 006Bh, 0065h, 005Fh, 005Ah, 0055h, 0050h, 004Ch, 0047h, 0043h, 0040h, 003Ch, 0039h  ; C6 C#6 D6 D#6 E6 F6 F#6 G6 G#6 A6 A#6 B6

; -----------------------------------------------------------------------------
; PSG_PlayNoteIndexed: Toca uma nota por índice na PSG_NoteTable em vez de um
; período calculado à mão
; Entrada: A = canal (0=A, 1=B, 2=C), D = índice da nota (0..59, ver
;          PSG_NoteTable), E = volume (0..15)
; -----------------------------------------------------------------------------
PSG_PlayNoteIndexed:
    PUSH AF
    PUSH DE

    LD H, 0
    LD L, D
    ADD HL, HL ; índice*2 (tabela de words)
    LD DE, PSG_NoteTable
    ADD HL, DE
    LD E, (HL)
    INC HL
    LD D, (HL)
    EX DE, HL  ; HL = período

    POP DE     ; E = volume (D descartado)
    POP AF     ; A = canal
    CALL PSG_PlayTone
    RET

ENDMOD

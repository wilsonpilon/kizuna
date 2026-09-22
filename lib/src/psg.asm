; =============================================================================
; KIZUNA MSXLIB - PSG.ASM
; Rotinas de controle do gerador de som PSG (AY-3-8910 / YM2149)
; =============================================================================

MODULE PSG
BANK 0

PUBLIC PSG_Write, PSG_Read, PSG_MuteAll, PSG_PlayTone
PUBLIC PSG_PlayNoteIndexed, PSG_PlaySequence

PSG_REG_SEL EQU 00A0h
PSG_DATA_WR EQU 00A1h
PSG_DATA_RD EQU 00A2h

; -----------------------------------------------------------------------------
; PSG_Write: Escreve um valor em um registrador do PSG
; Entrada: A = número do registrador (0..15), E = valor (8 bits)
; -----------------------------------------------------------------------------
PSG_Write:
    OUT (PSG_REG_SEL), A
    LD A, E
    OUT (PSG_DATA_WR), A
    RET

; -----------------------------------------------------------------------------
; PSG_Read: Lê o valor atual de um registrador do PSG
; Entrada: A = número do registrador (0..15)
; Saída: A = valor lido
; -----------------------------------------------------------------------------
PSG_Read:
    OUT (PSG_REG_SEL), A
    IN A, (PSG_DATA_RD)
    RET

; -----------------------------------------------------------------------------
; PSG_MuteAll: Zera o volume dos 3 canais de som e silencia o PSG
; -----------------------------------------------------------------------------
PSG_MuteAll:
    PUSH AF
    PUSH DE
    LD E, 00h
    LD A, 08h ; Canal A volume
    CALL PSG_Write
    LD A, 09h ; Canal B volume
    CALL PSG_Write
    LD A, 0Ah ; Canal C volume
    CALL PSG_Write
    ; Desativa tons no mixer de forma segura para portas de I/O do MSX (0xBF)
    LD A, 07h
    LD E, 0BFh
    CALL PSG_Write
    POP DE
    POP AF
    RET

; -----------------------------------------------------------------------------
; PSG_PlayTone: Configura frequência e toca tom em um canal (0, 1 ou 2)
; Entrada: A = canal (0=A, 1=B, 2=C)
;          HL = período da nota (12 bits: menor valor = frequência mais alta)
;          E = volume (0..15)
; -----------------------------------------------------------------------------
PSG_PlayTone:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL

    LD B, A ; B = canal (0, 1 ou 2)
    ADD A, A
    LD C, A ; C = reg fino (0, 2 ou 4)

    ; 1. Escrever parte baixa do período (8 bits)
    LD A, C
    OUT (PSG_REG_SEL), A
    LD A, L
    OUT (PSG_DATA_WR), A

    ; 2. Escrever parte alta do período (4 bits)
    INC C
    LD A, C
    OUT (PSG_REG_SEL), A
    LD A, H
    AND 0Fh
    OUT (PSG_DATA_WR), A

    ; 3. Escrever volume no canal correspondente (8 + canal)
    LD A, 08h
    ADD A, B
    OUT (PSG_REG_SEL), A
    LD A, E
    AND 0Fh
    OUT (PSG_DATA_WR), A

    ; 4. Ativar tom no misturador (Reg 7):
    ; Canal 0 -> 0xBE (Tom A), Canal 1 -> 0xBD (Tom B), Canal 2 -> 0xBB (Tom C)
    LD A, B
    LD E, 0BEh
    OR A
    JR Z, PSG_SetMixer
    LD E, 0BDh
    DEC A
    JR Z, PSG_SetMixer
    LD E, 0BBh
PSG_SetMixer:
    LD A, 07h
    CALL PSG_Write

    POP HL
    POP DE
    POP BC
    POP AF
    RET

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

; -----------------------------------------------------------------------------
; PSG_PlaySequence: Toca uma sequência de notas em RAM/ROM, uma após a outra
; Entrada: HL = ponteiro para a sequência: 4 bytes por evento (canal,
;          índice-de-nota, volume, duração), terminada por um byte de canal
;          = 0FFh. Duração é um contador de laço, não milissegundos -- use
;          valores maiores para notas mais longas; 0 se comporta como 256
;          (laço de 8 bits decrementando até dar a volta).
; -----------------------------------------------------------------------------
PSG_PlaySequence:
PSG_Seq_Loop:
    LD A, (HL)
    CP 0FFh
    RET Z

    LD A, (HL) ; canal
    INC HL
    LD D, (HL) ; índice da nota
    INC HL
    LD E, (HL) ; volume
    INC HL
    LD C, (HL) ; duração
    INC HL
    PUSH HL    ; salva o ponteiro da sequência (única PUSH desta iteração)

    CALL PSG_PlayNoteIndexed

PSG_Seq_DelayOuter:
    LD B, 00h
PSG_Seq_DelayInner:
    DJNZ PSG_Seq_DelayInner
    DEC C
    JR NZ, PSG_Seq_DelayOuter

    POP HL     ; restaura o ponteiro da sequência
    JR PSG_Seq_Loop

ENDMOD

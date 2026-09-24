; =============================================================================
; KIZUNA MSXLIB - vdp/clearvram
; zera a VRAM
; =============================================================================

MODULE vdp_clearvram
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_ClearVRAM
EXTERN VDP_VramSetWrite, VDP_VramFillStream

; VDP_ClearVRAM: zera A blocos de 16 KB a partir do endereco 0 (1 = os primeiros
; 16 KB ... 8 = os 128 KB).
; Entrada: A = quantidade de blocos de 16 KB (0..8)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_ClearVRAM:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD D, A             ; D = blocos que faltam
    LD E, 00h           ; E = indice do bloco
VDP_ClearVRAM_Loop:
    LD A, D
    OR A
    JR Z, VDP_ClearVRAM_Done
    LD A, E
    AND 03h
    RRCA
    RRCA                ; (E & 3) << 6 = byte alto do endereco de 16 bits
    LD H, A
    LD L, 00h
    LD A, E
    RRA
    RRA
    AND 01h             ; bit 16
    CALL VDP_VramSetWrite
    LD BC, 4000h
    XOR A
    PUSH DE             ; VDP_VramFillStream destroi E
    CALL VDP_VramFillStream
    POP DE
    INC E
    DEC D
    JR VDP_ClearVRAM_Loop
VDP_ClearVRAM_Done:
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

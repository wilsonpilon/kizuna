; =============================================================================
; KIZUNA MSXLIB - vdp/blinkregs
; cores e tempos do piscar do texto de 80 colunas (V9938)
; =============================================================================

MODULE vdp_blinkregs
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetBlinkColor, VDP_SetBlinkTime, VDP_BlinkOff
EXTERN VDP_SetReg

; No texto de 80 colunas (VDP_SetMode 9) cada celula pode piscar: a tabela de piscar
; (VDP_BlinkFill/Line/Cell) diz quais, e R#12/R#13 dizem com que cores e por quanto tempo.

; VDP_SetBlinkColor: cores dos caracteres que estao na fase "piscando" (R#12)
; Entrada: A = cor do texto (0..15), B = cor do fundo (0..15)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetBlinkColor:
    PUSH AF
    PUSH BC
    AND 0Fh
    RLCA
    RLCA
    RLCA
    RLCA
    LD C, A
    LD A, B
    AND 0Fh
    OR C
    LD B, A
    LD C, VDP_REG_BLINKCOL
    CALL VDP_SetReg
    POP BC
    POP AF
    RET

; VDP_SetBlinkTime: duracao das duas fases do piscar (R#13), em unidades de 1/6 s (V9938)
; Entrada: A = tempo na fase "piscando" (0..15), B = tempo na fase normal (0..15)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetBlinkTime:
    PUSH AF
    PUSH BC
    AND 0Fh
    RLCA
    RLCA
    RLCA
    RLCA
    LD C, A
    LD A, B
    AND 0Fh
    OR C
    LD B, A
    LD C, VDP_REG_BLINKTIM
    CALL VDP_SetReg
    POP BC
    POP AF
    RET

; VDP_BlinkOff: R#13 = 0, o texto para de piscar
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_BlinkOff:
    PUSH AF
    PUSH BC
    LD B, 00h
    LD C, VDP_REG_BLINKTIM
    CALL VDP_SetReg
    POP BC
    POP AF
    RET

ENDMOD

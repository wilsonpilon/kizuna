; =============================================================================
; KIZUNA MSXLIB - vdp/setlines
; 192 ou 212 linhas, 50 ou 60 Hz
; =============================================================================

MODULE vdp_setlines
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetLines, VDP_SetRefresh
EXTERN VDP_UpdateReg

; VDP_SetLines: numero de linhas da tela (R#9, bit LN)
; Entrada: A = 192 ou 212 (qualquer outro valor conta como 192)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetLines:
    PUSH AF
    PUSH BC
    PUSH DE
    LD E, 00h
    CP 0D4h             ; 212
    JR NZ, VDP_SetLines_Go
    LD E, VDP_R9_LN
VDP_SetLines_Go:
    LD C, VDP_REG_MODE3
    LD D, VDP_R9_LN
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

; VDP_SetRefresh: frequencia de quadros (R#9, bit NT)
; Entrada: A = 50 ou 60 (qualquer outro valor conta como 60)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetRefresh:
    PUSH AF
    PUSH BC
    PUSH DE
    LD E, 00h
    CP 32h              ; 50
    JR NZ, VDP_SetRefresh_Go
    LD E, VDP_R9_NT
VDP_SetRefresh_Go:
    LD C, VDP_REG_MODE3
    LD D, VDP_R9_NT
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

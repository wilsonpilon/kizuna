; =============================================================================
; KIZUNA MSXLIB - vdp/hblankint
; interrupcao horizontal do V9938 (R#0, bit IE1)
; =============================================================================

MODULE vdp_hblankint
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_HBlankInt
EXTERN VDP_UpdateReg

; VDP_HBlankInt: interrupcao horizontal do V9938 (R#0, bit IE1)
; Entrada: A = 0 = desliga, qualquer outro valor = liga
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HBlankInt:
    PUSH AF
    PUSH BC
    PUSH DE
    LD E, 00h
    OR A
    JR Z, VDP_HBlankInt_Go
    LD E, 10h
VDP_HBlankInt_Go:
    LD C, VDP_REG_MODE0
    LD D, 10h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

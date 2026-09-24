; =============================================================================
; KIZUNA MSXLIB - vdp/spritemag
; sprites ampliados 2x (R#1, bit MAG)
; =============================================================================

MODULE vdp_spritemag
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SpriteMag
EXTERN VDP_UpdateReg

; VDP_SpriteMag: sprites ampliados 2x (R#1, bit MAG)
; Entrada: A = 0 = normal, qualquer outro valor = 2x
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SpriteMag:
    PUSH AF
    PUSH BC
    PUSH DE
    LD E, 00h
    OR A
    JR Z, VDP_SpriteMag_Go
    LD E, 01h
VDP_SpriteMag_Go:
    LD C, VDP_REG_MODE1
    LD D, 01h
    CALL VDP_UpdateReg
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

; =============================================================================
; KIZUNA MSXLIB - vdp/getreg
; le a copia sombra de um registrador do VDP
; =============================================================================

MODULE vdp_getreg
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_GetReg
EXTERN VDP_Shadow

; VDP_GetReg: ultimo valor escrito em R#reg pelas rotinas com copia sombra
; Entrada: A = numero do registrador (0..63)
; Saída: A = valor
; Preserva: BC, DE, HL. Destrói: flags.
VDP_GetReg:
    PUSH HL
    LD HL, VDP_Shadow
    AND 3Fh
    ADD A, L
    LD L, A
    JR NC, VDP_GetReg_Read
    INC H
VDP_GetReg_Read:
    LD A, (HL)
    POP HL
    RET

ENDMOD

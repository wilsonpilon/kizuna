; =============================================================================
; KIZUNA MSXLIB - vdp/setreg
; escreve um registrador do VDP e guarda a copia sombra
; =============================================================================

MODULE vdp_setreg
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetReg
EXTERN VDP_Shadow, VDP_WriteReg

; VDP_SetReg: R#reg = valor, atualizando a copia sombra
; Entrada: C = numero do registrador (0..63), B = valor
; Preserva: BC, DE, HL, A. Destrói: flags. Reabilita as interrupcoes (EI) ao terminar.
VDP_SetReg:
    PUSH AF
    PUSH HL
    LD HL, VDP_Shadow
    LD A, C
    AND 3Fh
    ADD A, L
    LD L, A
    JR NC, VDP_SetReg_Store
    INC H
VDP_SetReg_Store:
    LD (HL), B
    CALL VDP_WriteReg
    POP HL
    POP AF
    RET

ENDMOD

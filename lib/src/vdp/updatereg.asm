; =============================================================================
; KIZUNA MSXLIB - vdp/updatereg
; altera so alguns bits de um registrador do VDP
; =============================================================================

MODULE vdp_updatereg
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_UpdateReg
EXTERN VDP_Shadow, VDP_SetReg

; VDP_UpdateReg: R#reg = (copia sombra E NAO mascara) OU (valor E mascara) -- muda
; so os bits da mascara e mantem os outros como estavam.
; Entrada: C = numero do registrador (0..63), D = mascara, E = valor (so os bits da mascara valem)
; Preserva: BC, DE, HL, A. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_UpdateReg:
    PUSH AF
    PUSH BC
    PUSH HL
    LD HL, VDP_Shadow
    LD A, C
    AND 3Fh
    ADD A, L
    LD L, A
    JR NC, VDP_UpdateReg_Old
    INC H
VDP_UpdateReg_Old:
    LD A, D
    CPL
    AND (HL)            ; bits antigos fora da mascara
    LD B, A
    LD A, E
    AND D               ; bits novos dentro da mascara
    OR B
    LD B, A
    CALL VDP_SetReg
    POP HL
    POP BC
    POP AF
    RET

ENDMOD

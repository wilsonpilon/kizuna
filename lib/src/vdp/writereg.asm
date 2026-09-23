; =============================================================================
; KIZUNA MSXLIB - vdp/writereg
; escreve registrador do VDP (com e sem DI/EI)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_writereg
BANK 0

PUBLIC VDP_WriteReg, VDP_WriteReg_Raw
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_WriteReg: Escreve valor em um registrador do VDP (0..46 no V9938/V9958;
; 0..23 nos MSX1/TMS9918 originais)
; Entrada: C = número do registrador, B = valor a escrever
; -----------------------------------------------------------------------------
VDP_WriteReg:
    DI
    CALL VDP_WriteReg_Raw
    EI
    RET

; -----------------------------------------------------------------------------
; VDP_WriteReg_Raw: mesma logica do VDP_WriteReg, sem DI/EI proprios -- para
; ser chamada em sequencia (varios registradores) dentro de um unico DI/EI
; externo (ex: VDP_PSet_Raw configurando R#36-46 de uma vez).
; -----------------------------------------------------------------------------
VDP_WriteReg_Raw:
    LD A, B
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, C
    OR 80h
    OUT (VDP_CMD), A
    RET

ENDMOD

; =============================================================================
; KIZUNA MSXLIB - vdp/setcolor
; cores de texto (R#7)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_setcolor
BANK 0

PUBLIC VDP_SetColor
EXTERN VDP_WriteReg

; -----------------------------------------------------------------------------
; VDP_SetColor: Define as cores de texto (frente) e fundo (Reg 7)
; Entrada: H = cor de frente (0..15), L = cor de fundo (0..15)
; -----------------------------------------------------------------------------
VDP_SetColor:
    LD A, H
    SLA A
    SLA A
    SLA A
    SLA A
    AND F0h
    LD B, A
    LD A, L
    AND 0Fh
    OR B
    LD B, A
    LD C, 07h
    CALL VDP_WriteReg
    RET

ENDMOD

; =============================================================================
; KIZUNA MSXLIB - vdp/hwspan
; distancia com sinal entre duas coordenadas (auxiliar das rotinas de linha e caixa)
; =============================================================================

MODULE vdp_hwspan
BANK 0


PUBLIC VDP_HwSpan

; VDP_HwSpan: |b - a| e se b esta antes de a
; Entrada: HL = b, DE = a
; Saída: HL = |b - a|, carry = 1 se b < a (sentido negativo)
; Preserva: BC, DE. Destrói: A, flags.
VDP_HwSpan:
    OR A
    SBC HL, DE
    RET NC
    LD A, H
    CPL
    LD H, A
    LD A, L
    CPL
    LD L, A
    INC HL
    SCF
    RET

ENDMOD

; =============================================================================
; KIZUNA MSXLIB - math/abs16
; valor absoluto de 16 bits com sinal
; =============================================================================

MODULE math_abs16
BANK 0

PUBLIC MATH_Abs16
EXTERN MATH_Neg16

; MATH_Abs16: HL = |HL| (o valor -32768 continua 8000h, que lido sem sinal e 32768)
; Entrada: HL
; Saída: HL
; Preserva: A, BC, DE. Destrói: flags.
MATH_Abs16:
    BIT 7, H
    RET Z
    JP MATH_Neg16

ENDMOD

; =============================================================================
; KIZUNA MSXLIB - math/sign16
; sinal de um inteiro de 16 bits
; =============================================================================

MODULE math_sign16
BANK 0

PUBLIC MATH_Sign16

; MATH_Sign16: sinal de HL como inteiro com sinal
; Entrada: HL
; Saída: A = 0 (HL = 0), 1 (positivo) ou 0FFh (-1, negativo)
; Preserva: BC, DE, HL. Destrói: flags.
MATH_Sign16:
    LD A, H
    OR L
    RET Z
    BIT 7, H
    LD A, 01h
    RET Z
    LD A, 0FFh
    RET

ENDMOD

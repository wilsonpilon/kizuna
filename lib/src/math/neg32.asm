; =============================================================================
; KIZUNA MSXLIB - math/neg32
; negacao de 32 bits com sinal
; =============================================================================

MODULE math_neg32
BANK 0

PUBLIC MATH_Neg32

; MATH_Neg32: DE:HL = -DE:HL (complemento de dois de 32 bits)
; Entrada: DE:HL (DE = 16 bits altos)
; Saída: DE:HL
; Preserva: BC. Destrói: A, flags.
MATH_Neg32:
    LD A, L
    CPL
    LD L, A
    LD A, H
    CPL
    LD H, A
    LD A, E
    CPL
    LD E, A
    LD A, D
    CPL
    LD D, A
    INC HL
    LD A, H
    OR L
    RET NZ              ; nao houve vai-um para a palavra alta
    INC DE
    RET

ENDMOD

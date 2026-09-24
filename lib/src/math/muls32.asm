; =============================================================================
; KIZUNA MSXLIB - math/muls32
; multiplicacao 16x16 -> 32 bits, com sinal
; =============================================================================

MODULE math_muls32
BANK 0

PUBLIC MATH_MulS16x16
EXTERN MATH_Neg16, MATH_Neg32, MATH_MulU16x16

; MATH_MulS16x16: produto completo de 32 bits, com sinal
; Entrada: HL, DE (complemento de dois)
; Saída: DE:HL = HL * DE
; Preserva: BC. Destrói: A, flags.
MATH_MulS16x16:
    PUSH BC
    LD A, H
    XOR D
    LD B, A             ; bit 7 de B = sinal do resultado
    BIT 7, H
    CALL NZ, MATH_Neg16 ; HL = |HL|
    EX DE, HL
    BIT 7, H
    CALL NZ, MATH_Neg16 ; (ex-DE) = |DE|
    EX DE, HL
    CALL MATH_MulU16x16
    BIT 7, B
    CALL NZ, MATH_Neg32
    POP BC
    RET

ENDMOD

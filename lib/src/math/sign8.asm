; =============================================================================
; KIZUNA MSXLIB - math/sign8
; negacao e valor absoluto de 8 bits e de 32 bits
; =============================================================================

MODULE math_sign8
BANK 0

PUBLIC MATH_Neg8, MATH_Abs8, MATH_Abs32
EXTERN MATH_Neg32

; MATH_Neg8: A = -A (complemento de dois)
; Preserva: BC, DE, HL. Destrói: flags.
MATH_Neg8:
    NEG
    RET

; MATH_Abs8: A = |A| (-128 continua 80h, que sem sinal e 128)
; Preserva: BC, DE, HL. Destrói: flags.
MATH_Abs8:
    OR A
    RET P
    NEG
    RET

; MATH_Abs32: DE:HL = |DE:HL|
; Preserva: A, BC. Destrói: flags.
MATH_Abs32:
    BIT 7, D
    RET Z
    PUSH AF
    CALL MATH_Neg32
    POP AF
    RET

ENDMOD

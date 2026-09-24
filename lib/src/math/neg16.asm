; =============================================================================
; KIZUNA MSXLIB - math/neg16
; negacao de 16 bits com sinal
; =============================================================================

MODULE math_neg16
BANK 0

PUBLIC MATH_Neg16

; MATH_Neg16: HL = -HL (complemento de dois; -32768 continua -32768)
; Entrada: HL
; Saída: HL
; Preserva: A, BC, DE. Destrói: flags.
MATH_Neg16:
    PUSH DE
    EX DE, HL
    LD HL, 0000h
    OR A
    SBC HL, DE
    POP DE
    RET

ENDMOD

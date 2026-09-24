; =============================================================================
; KIZUNA MSXLIB - math/mul8
; multiplicacao 8x8 -> 16 bits, sem sinal
; =============================================================================

MODULE math_mul8
BANK 0

PUBLIC MATH_Mul8

; MATH_Mul8: HL = A * E
; Entrada: A, E (sem sinal)
; Saída: HL (0..65025)
; Preserva: BC, DE. Destrói: A, flags.
MATH_Mul8:
    PUSH BC
    PUSH DE
    LD D, 00h           ; DE = multiplicando
    LD HL, 0000h
    LD B, 08h
MATH_Mul8_Loop:
    ADD HL, HL
    RLA                 ; proximo bit do multiplicador (do mais alto) vai ao carry
    JR NC, MATH_Mul8_Skip
    ADD HL, DE
MATH_Mul8_Skip:
    DJNZ MATH_Mul8_Loop
    POP DE
    POP BC
    RET

ENDMOD

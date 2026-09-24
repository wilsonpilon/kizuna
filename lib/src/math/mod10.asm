; =============================================================================
; KIZUNA MSXLIB - math/mod10
; resto da divisao por 10 (sem sinal) e divisao por 10 com sinal
; =============================================================================

MODULE math_mod10
BANK 0

PUBLIC MATH_Mod10, MATH_DivS10
EXTERN MATH_DivMod10, MATH_Neg16

; MATH_Mod10: A = HL mod 10 (sem sinal)
; Entrada: HL
; Saída: A = 0..9. HL e o resto dos registradores ficam como estavam.
; Preserva: BC, DE, HL. Destrói: flags.
MATH_Mod10:
    PUSH HL
    CALL MATH_DivMod10
    POP HL
    RET

; MATH_DivS10: divide HL (com sinal) por 10, truncando para zero
; Entrada: HL
; Saída: HL = quociente, A = resto (o sinal do dividendo: -9..9 em complemento de dois)
; Preserva: BC, DE. Destrói: flags.
MATH_DivS10:
    BIT 7, H
    JR NZ, MATH_DivS10_Neg
    JP MATH_DivMod10
MATH_DivS10_Neg:
    CALL MATH_Neg16     ; |HL| (-32768 fica 8000h = 32768 sem sinal, correto aqui)
    CALL MATH_DivMod10  ; HL = quociente, A = resto (positivos)
    CALL MATH_Neg16     ; quociente com sinal
    NEG                 ; resto com o sinal do dividendo
    RET

ENDMOD

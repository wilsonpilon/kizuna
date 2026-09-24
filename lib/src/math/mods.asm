; =============================================================================
; KIZUNA MSXLIB - math/mods
; resto da divisao de 16 bits (sem sinal e com sinal), devolvido em HL
; =============================================================================

MODULE math_mods
BANK 0

PUBLIC MATH_Mod16, MATH_ModS16
EXTERN Div16, MATH_DivS16

; MATH_Mod16: HL = HL mod DE, sem sinal. Divisor zero: HL = 0.
; Preserva: BC. Destrói: A, DE, flags.
MATH_Mod16:
    CALL Div16
    EX DE, HL
    RET

; MATH_ModS16: HL = resto da divisao com sinal (o sinal do dividendo, como em
; C e Pascal). Divisor zero: HL = 0.
; Preserva: BC. Destrói: A, DE, flags.
MATH_ModS16:
    CALL MATH_DivS16
    EX DE, HL
    RET

ENDMOD

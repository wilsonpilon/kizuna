; =============================================================================
; KIZUNA MSXLIB - math/divmod10
; divisao por 10 (rapida, sem sinal) -- a base da conversao para decimal
; =============================================================================

MODULE math_divmod10
BANK 0

PUBLIC MATH_DivMod10

; MATH_DivMod10: divide HL (sem sinal) por 10
; Entrada: HL
; Saída: HL = quociente, A = resto (0..9)
; Preserva: BC, DE. Destrói: flags.
MATH_DivMod10:
    PUSH BC
    XOR A
    LD B, 10h
MATH_DivMod10_Loop:
    ADD HL, HL          ; proximo bit do dividendo sai no carry
    RLA                 ; e entra no resto (A <= 19 aqui)
    CP 0Ah
    JR C, MATH_DivMod10_Next
    SUB 0Ah
    INC L               ; bit 0 do quociente
MATH_DivMod10_Next:
    DJNZ MATH_DivMod10_Loop
    POP BC
    RET

ENDMOD

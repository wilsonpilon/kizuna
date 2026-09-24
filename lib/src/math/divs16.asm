; =============================================================================
; KIZUNA MSXLIB - math/divs16
; divisao inteira de 16 bits com sinal
; =============================================================================

MODULE math_divs16
BANK 0

PUBLIC MATH_DivS16
EXTERN MATH_Neg16, Div16

; MATH_DivS16: divisao com sinal, truncada em direcao a zero (como C e Pascal)
; Entrada: HL = dividendo, DE = divisor
; Saída: HL = quociente, DE = resto (o resto tem o sinal do dividendo)
; Divisor zero: HL = 7FFFh (dividendo >= 0) ou 8000h (dividendo < 0), DE = 0.
; -32768 / -1 estoura e devolve 8000h.
; Preserva: BC. Destrói: A, flags.
MATH_DivS16:
    PUSH BC
    LD A, H
    LD B, A             ; bit 7 de B = sinal do dividendo (sinal do resto)
    XOR D
    LD C, A             ; bit 7 de C = sinal do quociente
    BIT 7, H
    CALL NZ, MATH_Neg16 ; HL = |dividendo|
    EX DE, HL
    BIT 7, H
    CALL NZ, MATH_Neg16 ; (ex-DE) = |divisor|
    EX DE, HL           ; HL = |dividendo|, DE = |divisor|
    LD A, D
    OR E
    JR Z, MATH_DivS16_Zero
    CALL Div16          ; HL = quociente, DE = resto (sem sinal)
    BIT 7, C
    CALL NZ, MATH_Neg16 ; quociente com sinal
    BIT 7, B
    JR Z, MATH_DivS16_Done
    EX DE, HL
    CALL MATH_Neg16     ; resto com o sinal do dividendo
    EX DE, HL
MATH_DivS16_Done:
    POP BC
    RET

MATH_DivS16_Zero:
    LD HL, 7FFFh
    BIT 7, B
    JR Z, MATH_DivS16_ZeroOut
    LD HL, 8000h
MATH_DivS16_ZeroOut:
    LD DE, 0000h
    POP BC
    RET

ENDMOD

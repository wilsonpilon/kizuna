; =============================================================================
; KIZUNA MSXLIB - math/randbetween
; numeros aleatorios num intervalo [minimo, maximo)
; =============================================================================

MODULE math_randbetween
BANK 0

PUBLIC MATH_RandBetween16, MATH_RandBetween8
EXTERN MATH_RandRange16

; MATH_RandBetween16: numero em [lo, hi)
; Entrada: HL = lo, DE = hi (sem sinal). Se hi <= lo, devolve lo.
; Saída: HL
; Preserva: BC, DE. Destrói: A, flags.
MATH_RandBetween16:
    PUSH DE
    PUSH HL             ; guarda lo
    EX DE, HL           ; HL = hi, DE = lo
    OR A
    SBC HL, DE          ; HL = hi - lo
    JR C, MATH_RandBetween16_Lo
    CALL MATH_RandRange16   ; HL = 0..(hi-lo)-1  (0 se hi = lo)
    POP DE              ; DE = lo
    ADD HL, DE
    POP DE              ; restaura o DE do chamador (hi)
    RET
MATH_RandBetween16_Lo:
    POP HL              ; HL = lo
    POP DE
    RET

; MATH_RandBetween8: numero em [lo, hi)
; Entrada: A = lo, E = hi (sem sinal). Se hi <= lo, devolve lo.
; Saída: A
; Preserva: BC, DE, HL. Destrói: flags.
MATH_RandBetween8:
    PUSH HL
    PUSH DE
    LD L, A
    LD H, 00h
    LD D, 00h
    CALL MATH_RandBetween16
    LD A, L
    POP DE
    POP HL
    RET

ENDMOD

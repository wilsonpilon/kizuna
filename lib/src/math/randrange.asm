; =============================================================================
; KIZUNA MSXLIB - math/randrange
; numeros aleatorios em faixas (8 bits, 16 bits e intervalo)
; =============================================================================

MODULE math_randrange
BANK 0

PUBLIC MATH_Rand8, MATH_RandRange16, MATH_RandRange8
EXTERN MATH_Rand16, Div16

; MATH_Rand8: um byte pseudoaleatorio
; Saída: A = 0..255
; Preserva: BC, DE, HL. Destrói: flags.
MATH_Rand8:
    PUSH HL
    CALL MATH_Rand16
    LD A, H
    XOR L               ; mistura os dois bytes
    POP HL
    RET

; MATH_RandRange16: numero em [0, max)
; Entrada: HL = max. max = 0 devolve 0.
; Saída: HL = 0..max-1
; Usa "resto da divisao", que tem um leve vies quando max nao divide 65536
; (irrelevante para jogos; para max potencia de 2 nao ha vies).
; Preserva: BC, DE. Destrói: A, flags.
MATH_RandRange16:
    PUSH DE
    LD A, H
    OR L
    JR Z, MATH_RandRange16_Done
    LD D, H
    LD E, L             ; DE = max
    CALL MATH_Rand16    ; HL = aleatorio
    CALL Div16          ; DE = aleatorio mod max
    EX DE, HL
MATH_RandRange16_Done:
    POP DE
    RET

; MATH_RandRange8: numero em [0, max) com max de 8 bits
; Entrada: A = max. max = 0 devolve 0.
; Saída: A = 0..max-1
; Preserva: BC, DE, HL. Destrói: flags.
MATH_RandRange8:
    PUSH HL
    LD L, A
    LD H, 00h
    CALL MATH_RandRange16
    LD A, L
    POP HL
    RET

ENDMOD

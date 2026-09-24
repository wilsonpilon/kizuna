; =============================================================================
; KIZUNA MSXLIB - math/cmp16
; comparacao de 16 bits (sem sinal e com sinal)
; =============================================================================

MODULE math_cmp16
BANK 0

PUBLIC MATH_Cmp16U, MATH_Cmp16S

; MATH_Cmp16U: compara HL com DE, sem sinal
; Saída: flag C = 1 se HL < DE; flag Z = 1 se HL = DE (C = 0 e Z = 0: HL > DE)
; Preserva: A, BC, DE, HL.
MATH_Cmp16U:
    PUSH HL
    OR A
    SBC HL, DE
    POP HL
    RET

; MATH_Cmp16S: compara HL com DE, com sinal (complemento de dois)
; Saída: flag C = 1 se HL < DE; flag Z = 1 se HL = DE
; Preserva: A, BC, DE, HL.
; Truque: inverter o bit de sinal dos dois lados transforma a comparação com
; sinal numa comparação sem sinal.
MATH_Cmp16S:
    PUSH HL
    PUSH DE
    PUSH AF
    LD A, H
    XOR 80h
    LD H, A
    LD A, D
    XOR 80h
    LD D, A
    POP AF
    OR A
    SBC HL, DE
    POP DE
    POP HL
    RET

ENDMOD

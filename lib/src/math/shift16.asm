; =============================================================================
; KIZUNA MSXLIB - math/shift16
; deslocamentos de 16 bits por uma contagem variavel
; =============================================================================

MODULE math_shift16
BANK 0

PUBLIC MATH_Shl16, MATH_Shr16, MATH_Sar16

; Todas: Entrada HL = valor, B = quantidade de bits (0..255)
; Saída: HL. Contagem >= 16 zera o resultado (Sar16: preenche com o sinal).
; Preserva: A, BC, DE. Destrói: flags.

; MATH_Shl16: deslocamento para a esquerda
MATH_Shl16:
    PUSH AF
    PUSH BC
    LD A, B
    OR A
    JR Z, MATH_Shl16_Done
    CP 10h
    JR C, MATH_Shl16_Loop
    LD HL, 0000h
    JR MATH_Shl16_Done
MATH_Shl16_Loop:
    ADD HL, HL
    DJNZ MATH_Shl16_Loop
MATH_Shl16_Done:
    POP BC
    POP AF
    RET

; MATH_Shr16: deslocamento logico para a direita (entra zero)
MATH_Shr16:
    PUSH AF
    PUSH BC
    LD A, B
    OR A
    JR Z, MATH_Shr16_Done
    CP 10h
    JR C, MATH_Shr16_Loop
    LD HL, 0000h
    JR MATH_Shr16_Done
MATH_Shr16_Loop:
    SRL H
    RR L
    DJNZ MATH_Shr16_Loop
MATH_Shr16_Done:
    POP BC
    POP AF
    RET

; MATH_Sar16: deslocamento aritmetico para a direita (o bit de sinal se repete)
MATH_Sar16:
    PUSH AF
    PUSH BC
    LD A, B
    OR A
    JR Z, MATH_Sar16_Done
    CP 10h
    JR C, MATH_Sar16_Loop
    LD A, H
    RLA
    SBC A, A            ; A = 00h ou 0FFh conforme o sinal
    LD H, A
    LD L, A
    JR MATH_Sar16_Done
MATH_Sar16_Loop:
    SRA H
    RR L
    DJNZ MATH_Sar16_Loop
MATH_Sar16_Done:
    POP BC
    POP AF
    RET

ENDMOD

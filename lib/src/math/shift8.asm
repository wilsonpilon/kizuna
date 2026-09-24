; =============================================================================
; KIZUNA MSXLIB - math/shift8
; deslocamentos de 8 bits por uma contagem variavel
; =============================================================================

MODULE math_shift8
BANK 0

PUBLIC MATH_Shl8, MATH_Shr8, MATH_Sar8

; Todas: Entrada A = valor, B = quantidade de bits (0..255)
; Saída: A. Contagem >= 8 zera o resultado (Sar8: preenche com o sinal).
; Preserva: BC, DE, HL. Destrói: flags.

MATH_Shl8:
    PUSH BC
    INC B
MATH_Shl8_Loop:
    DEC B
    JR Z, MATH_Shl8_Done
    ADD A, A            ; deslocamento de 1 bit; uma contagem grande vira 0 sozinha
    JR MATH_Shl8_Loop
MATH_Shl8_Done:
    POP BC
    RET

MATH_Shr8:
    PUSH BC
    INC B
MATH_Shr8_Loop:
    DEC B
    JR Z, MATH_Shr8_Done
    SRL A
    JR MATH_Shr8_Loop
MATH_Shr8_Done:
    POP BC
    RET

MATH_Sar8:
    PUSH BC
    INC B
MATH_Sar8_Loop:
    DEC B
    JR Z, MATH_Sar8_Done
    SRA A               ; o bit de sinal se repete; contagem grande vira 00h ou FFh
    JR MATH_Sar8_Loop
MATH_Sar8_Done:
    POP BC
    RET

ENDMOD

; =============================================================================
; KIZUNA MSXLIB - math/sqrt16
; raiz quadrada inteira de 16 bits
; =============================================================================

MODULE math_sqrt16
BANK 0

PUBLIC MATH_Sqrt16

; MATH_Sqrt16: piso da raiz quadrada de HL (sem sinal)
; Entrada: HL (0..65535)
; Saída: A = raiz (0..255)
; Preserva: BC, DE, HL. Destrói: flags.
; Metodo: subtrai os impares 1, 3, 5, ... enquanto couber -- a soma dos k
; primeiros impares e k*k. No pior caso (65535) sao 255 passos (~4 ms num Z80
; de 3,58 MHz); a API permite trocar por algo mais rapido sem mudar nada.
MATH_Sqrt16:
    PUSH HL
    PUSH DE
    LD DE, 0001h        ; proximo impar
    XOR A               ; contador = raiz
MATH_Sqrt16_Loop:
    OR A
    SBC HL, DE
    JR C, MATH_Sqrt16_Done
    INC A
    INC DE
    INC DE
    JR MATH_Sqrt16_Loop
MATH_Sqrt16_Done:
    POP DE
    POP HL
    RET

ENDMOD

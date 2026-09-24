; =============================================================================
; KIZUNA MSXLIB - math/clamp16
; limita um inteiro de 16 bits com sinal a um intervalo
; =============================================================================

MODULE math_clamp16
BANK 0

PUBLIC MATH_Clamp16S
EXTERN MATH_Cmp16S

; MATH_Clamp16S: HL = min(max(HL, DE), BC), tudo com sinal
; Entrada: HL = valor, DE = minimo, BC = maximo
; Saída: HL. Se minimo > maximo, o resultado e o maximo.
; Preserva: A, BC, DE. Destrói: flags.
MATH_Clamp16S:
    CALL MATH_Cmp16S    ; valor < minimo ?
    JR NC, MATH_Clamp16S_Hi
    LD H, D
    LD L, E             ; valor = minimo
MATH_Clamp16S_Hi:
    PUSH DE
    LD D, B
    LD E, C             ; DE = maximo
    CALL MATH_Cmp16S    ; valor <= maximo ?
    POP DE
    RET C
    RET Z
    LD H, B
    LD L, C             ; valor = maximo
    RET

ENDMOD

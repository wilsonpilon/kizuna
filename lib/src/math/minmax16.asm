; =============================================================================
; KIZUNA MSXLIB - math/minmax16
; minimo e maximo de dois inteiros de 16 bits (sem sinal e com sinal)
; =============================================================================

MODULE math_minmax16
BANK 0

PUBLIC MATH_Min16U, MATH_Max16U, MATH_Min16S, MATH_Max16S
EXTERN MATH_Cmp16U, MATH_Cmp16S

; Todas: Entrada HL, DE. Saída HL = o menor (Min) ou o maior (Max).
; Preserva: A, BC, DE. Destrói: flags.

MATH_Min16U:
    CALL MATH_Cmp16U
    RET C               ; HL < DE: HL ja e o menor
    LD H, D
    LD L, E
    RET

MATH_Max16U:
    CALL MATH_Cmp16U
    RET NC              ; HL >= DE: HL ja e o maior
    LD H, D
    LD L, E
    RET

MATH_Min16S:
    CALL MATH_Cmp16S
    RET C
    LD H, D
    LD L, E
    RET

MATH_Max16S:
    CALL MATH_Cmp16S
    RET NC
    LD H, D
    LD L, E
    RET

ENDMOD

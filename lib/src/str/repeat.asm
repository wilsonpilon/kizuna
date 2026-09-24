; =============================================================================
; KIZUNA MSXLIB - str/repeat
; STRING$: um caractere repetido N vezes
; =============================================================================

MODULE str_repeat
BANK 0

PUBLIC STR_Repeat

; STR_Repeat: monta em DE uma string com o caractere A repetido B vezes
; Entrada: DE = destino (B + 1 bytes), A = caractere, B = quantidade (0 = vazia)
; Preserva: DE, HL. Destrói: A, BC, flags.
STR_Repeat:
    PUSH DE
    LD C, A             ; C = caractere
    LD A, B
    LD (DE), A          ; tamanho
    INC DE
    OR A
    JR Z, STR_Repeat_Done
STR_Repeat_Loop:
    LD A, C
    LD (DE), A
    INC DE
    DJNZ STR_Repeat_Loop
STR_Repeat_Done:
    POP DE
    RET

ENDMOD

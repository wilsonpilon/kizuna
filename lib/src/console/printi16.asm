; =============================================================================
; KIZUNA MSXLIB - console/printi16
; imprime um inteiro com sinal no console
; =============================================================================

MODULE console_printi16
BANK 0

PUBLIC CON_PrintI16
EXTERN NUM_I16ToCStr, CON_PrintCStr

; CON_PrintI16: imprime HL em decimal com sinal (a versao sem sinal e PrintDec16)
; Preserva: todos os registradores. Nao reentrante (usa um buffer proprio).
CON_PrintI16:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD DE, CON_PrintI16_Buf
    CALL NUM_I16ToCStr
    LD HL, CON_PrintI16_Buf
    CALL CON_PrintCStr
    POP HL
    POP DE
    POP BC
    POP AF
    RET

CON_PrintI16_Buf:
    DS 8

ENDMOD

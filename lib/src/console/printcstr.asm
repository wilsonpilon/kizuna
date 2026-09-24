; =============================================================================
; KIZUNA MSXLIB - console/printcstr
; imprime uma string terminada em zero no console
; =============================================================================

MODULE console_printcstr
BANK 0

PUBLIC CON_PrintCStr, CON_NewLine, CON_PrintLine
EXTERN BDOS_PrintChar

; CON_PrintCStr: imprime a string de HL no console (sem mudar de linha)
; Preserva: todos os registradores.
CON_PrintCStr:
    PUSH AF
    PUSH DE
    PUSH HL
CON_PrintCStr_Loop:
    LD A, (HL)
    OR A
    JR Z, CON_PrintCStr_Done
    LD E, A
    CALL BDOS_PrintChar
    INC HL
    JR CON_PrintCStr_Loop
CON_PrintCStr_Done:
    POP HL
    POP DE
    POP AF
    RET

; CON_NewLine: CR + LF
; Preserva: todos os registradores.
CON_NewLine:
    PUSH DE
    LD E, 0Dh
    CALL BDOS_PrintChar
    LD E, 0Ah
    CALL BDOS_PrintChar
    POP DE
    RET

; CON_PrintLine: imprime a string de HL e muda de linha
; Preserva: todos os registradores.
CON_PrintLine:
    CALL CON_PrintCStr
    JP CON_NewLine

ENDMOD

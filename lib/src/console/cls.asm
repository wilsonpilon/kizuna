; =============================================================================
; KIZUNA MSXLIB - console/cls
; limpa a tela e posiciona o cursor no console de texto
; =============================================================================

MODULE console_cls
BANK 0

PUBLIC CON_Cls, CON_Locate
EXTERN BDOS_PrintChar

; CON_Cls: limpa a tela de texto (codigo 0Ch) e leva o cursor ao canto superior esquerdo
; Preserva: todos os registradores.
CON_Cls:
    PUSH DE
    LD E, 0Ch
    CALL BDOS_PrintChar
    POP DE
    RET

; CON_Locate: posiciona o cursor (sequencia ESC Y linha+32 coluna+32; a primeira
; coluna e a primeira linha sao 0)
; Entrada: A = coluna, B = linha
; Preserva: todos os registradores.
CON_Locate:
    PUSH AF
    PUSH DE
    PUSH BC
    LD E, 1Bh
    CALL BDOS_PrintChar
    LD E, 59h           ; 'Y'
    CALL BDOS_PrintChar
    LD A, B
    ADD A, 20h
    LD E, A
    CALL BDOS_PrintChar
    POP BC
    POP DE
    POP AF
    PUSH AF
    PUSH DE
    PUSH BC
    ADD A, 20h
    LD E, A
    CALL BDOS_PrintChar
    POP BC
    POP DE
    POP AF
    RET

ENDMOD

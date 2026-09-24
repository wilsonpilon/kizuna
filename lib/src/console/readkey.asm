; =============================================================================
; KIZUNA MSXLIB - console/readkey
; leitura de tecla e teste de tecla pressionada
; =============================================================================

MODULE console_readkey
BANK 0

PUBLIC CON_ReadKey, CON_KeyPressed
EXTERN BDOS_Call

; CON_ReadKey: espera uma tecla e devolve o codigo, sem ecoar (funcao 08h)
; Saída: A = caractere
; Preserva: BC, DE, HL. Destrói: flags.
CON_ReadKey:
    PUSH BC
    PUSH DE
    PUSH HL
    LD C, 08h
    CALL BDOS_Call
    POP HL
    POP DE
    POP BC
    RET

; CON_KeyPressed: ha uma tecla esperando no buffer? (funcao 0Bh), sem consumi-la
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando nao
; Preserva: BC, DE, HL.
CON_KeyPressed:
    PUSH BC
    PUSH DE
    PUSH HL
    LD C, 0Bh
    CALL BDOS_Call
    POP HL
    POP DE
    POP BC
    OR A
    LD A, 00h
    RET Z
    INC A
    RET

ENDMOD

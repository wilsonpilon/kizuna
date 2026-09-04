; =============================================================================
; KIZUNA MSXLIB - BDOS.ASM
; Rotinas de suporte para chamadas ao kernel do MSX-DOS (BDOS 0x0005)
; =============================================================================

MODULE BDOS
BANK 0

PUBLIC BDOS_Call, BDOS_PrintChar, BDOS_PrintString, BDOS_ReadChar, BDOS_Exit

BDOS_ENTRY EQU 0005h

; -----------------------------------------------------------------------------
; BDOS_Call: Executa chamada direta ao BDOS com função em C
; Entrada: C = número da função BDOS, registradores conforme a função
; -----------------------------------------------------------------------------
BDOS_Call:
    CALL BDOS_ENTRY
    RET

; -----------------------------------------------------------------------------
; BDOS_PrintChar: Imprime um único caractere no console (Função 02h)
; Entrada: E = código ASCII do caractere
; -----------------------------------------------------------------------------
BDOS_PrintChar:
    PUSH BC
    LD C, 02h
    CALL BDOS_ENTRY
    POP BC
    RET

; -----------------------------------------------------------------------------
; BDOS_PrintString: Imprime string terminada em '$' no console (Função 09h)
; Entrada: DE = ponteiro para a string terminada em '$'
; -----------------------------------------------------------------------------
BDOS_PrintString:
    PUSH BC
    LD C, 09h
    CALL BDOS_ENTRY
    POP BC
    RET

; -----------------------------------------------------------------------------
; BDOS_ReadChar: Lê um caractere do teclado com eco (Função 01h)
; Saída: A = caractere lido
; -----------------------------------------------------------------------------
BDOS_ReadChar:
    PUSH BC
    LD C, 01h
    CALL BDOS_ENTRY
    POP BC
    RET

; -----------------------------------------------------------------------------
; BDOS_Exit: Termina o programa e retorna ao MSX-DOS (Função 00h)
; -----------------------------------------------------------------------------
BDOS_Exit:
    LD C, 00h
    CALL BDOS_ENTRY
    RET

ENDMOD

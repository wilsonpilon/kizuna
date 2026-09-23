; =============================================================================
; KIZUNA MSXLIB - bdos/readchar
; le um caractere do teclado
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_readchar
BANK 0

PUBLIC BDOS_ReadChar
INCLUDE "../../inc/bdos.inc"

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

ENDMOD

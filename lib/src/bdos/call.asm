; =============================================================================
; KIZUNA MSXLIB - bdos/call
; chamada direta ao BDOS
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_call
BANK 0

PUBLIC BDOS_Call
INCLUDE "../../inc/bdos.inc"

; -----------------------------------------------------------------------------
; BDOS_Call: Executa chamada direta ao BDOS com função em C
; Entrada: C = número da função BDOS, registradores conforme a função
; -----------------------------------------------------------------------------
BDOS_Call:
    CALL BDOS_ENTRY
    RET

ENDMOD

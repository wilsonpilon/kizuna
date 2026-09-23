; =============================================================================
; KIZUNA MSXLIB - bdos/exit
; termina o programa
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_exit
BANK 0

PUBLIC BDOS_Exit
INCLUDE "../../inc/bdos.inc"

; -----------------------------------------------------------------------------
; BDOS_Exit: Termina o programa e retorna ao MSX-DOS (Função 00h)
; -----------------------------------------------------------------------------
BDOS_Exit:
    LD C, 00h
    CALL BDOS_ENTRY
    RET

ENDMOD

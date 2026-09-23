; =============================================================================
; KIZUNA MSXLIB - bdos/printstring
; imprime string terminada em '$'
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_printstring
BANK 0

PUBLIC BDOS_PrintString
INCLUDE "../../inc/bdos.inc"

; -----------------------------------------------------------------------------
; BDOS_PrintString: Imprime string terminada em '$' no console (Função 09h)
; Entrada: DE = ponteiro para a string terminada em '$'
; Preserva: todos os registradores
; -----------------------------------------------------------------------------
BDOS_PrintString:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    PUSH IX
    PUSH IY
    LD C, 09h
    CALL BDOS_ENTRY
    POP IY
    POP IX
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

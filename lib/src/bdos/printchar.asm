; =============================================================================
; KIZUNA MSXLIB - bdos/printchar
; imprime um caractere
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_printchar
BANK 0

PUBLIC BDOS_PrintChar
INCLUDE "../../inc/bdos.inc"

; -----------------------------------------------------------------------------
; BDOS_PrintChar: Imprime um único caractere no console (Função 02h)
; Entrada: E = código ASCII do caractere
; Preserva: todos os registradores
; -----------------------------------------------------------------------------
BDOS_PrintChar:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    PUSH IX
    PUSH IY
    LD C, 02h
    CALL BDOS_ENTRY
    POP IY
    POP IX
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

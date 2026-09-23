; =============================================================================
; KIZUNA MSXLIB - bios/chput
; imprime caractere via BDOS
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bios_chput
BANK 0

PUBLIC BIOS_CHPUT
INCLUDE "../../inc/bdos.inc"

; -----------------------------------------------------------------------------
; BIOS_CHPUT: Imprime caractere no console via BDOS (Função 02h)
; Entrada: A = código ASCII do caractere
; -----------------------------------------------------------------------------
BIOS_CHPUT:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD E, A
    LD C, 02h
    CALL BDOS_ENTRY
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

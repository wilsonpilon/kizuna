; =============================================================================
; KIZUNA MSXLIB - bios/beep
; beep padrao do MSX
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bios_beep
BANK 0

PUBLIC BIOS_BEEP
INCLUDE "../../inc/bdos.inc"

; -----------------------------------------------------------------------------
; BIOS_BEEP: Emite som de aviso padrão do MSX (BEL ASCII 07h via BDOS)
; -----------------------------------------------------------------------------
BIOS_BEEP:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD E, 07h
    LD C, 02h
    CALL BDOS_ENTRY
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

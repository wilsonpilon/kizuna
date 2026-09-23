; =============================================================================
; KIZUNA MSXLIB - bios/call
; chamada inter-slot a Main-ROM (BIOS_Call)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bios_call
BANK 0

PUBLIC BIOS_Call
INCLUDE "../../inc/bios.inc"

; -----------------------------------------------------------------------------
; BIOS_Call: Executa inter-slot call para a Main-ROM BIOS (Slot primário lido de EXPTBL)
; Entrada: IX = endereço da rotina na Main-ROM
; -----------------------------------------------------------------------------
BIOS_Call:
    PUSH AF
    PUSH HL
    ; CALSLT espera em IY o slot-id armazenado em EXPTBL-1 (FCC0h).
    ; O KAJI80 ainda não aceita LD IY,(nn), então copiamos o word por HL.
    LD HL, 0FCC0h
    LD E, (HL)
    INC HL
    LD D, (HL)
    PUSH DE
    POP IY
    POP HL
    POP AF
    CALL CALSLT
    EI
    RET

ENDMOD

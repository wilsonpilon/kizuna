; =============================================================================
; KIZUNA MSXLIB - string/printhex
; imprime hexadecimal (8 e 16 bits)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE string_printhex
BANK 0

PUBLIC PrintHex16, PrintHex8
EXTERN BDOS_PrintChar

; -----------------------------------------------------------------------------
; PrintHex8: Imprime byte em A como 2 dígitos hexadecimais no console
; Entrada: A = byte
; -----------------------------------------------------------------------------
PrintHex8:
    PUSH AF
    RRA
    RRA
    RRA
    RRA
    CALL PrintNibble
    POP AF
    CALL PrintNibble
    RET

PrintNibble:
    AND 0Fh
    CP 0Ah
    JR C, NibbleDigit
    ADD A, 07h
NibbleDigit:
    ADD A, 30h
    PUSH BC
    LD E, A
    CALL BDOS_PrintChar
    POP BC
    RET

; -----------------------------------------------------------------------------
; PrintHex16: Imprime palavra de 16 bits em HL como 4 dígitos hexadecimais
; Entrada: HL = palavra de 16 bits
; -----------------------------------------------------------------------------
PrintHex16:
    PUSH AF
    LD A, H
    CALL PrintHex8
    LD A, L
    CALL PrintHex8
    POP AF
    RET

ENDMOD

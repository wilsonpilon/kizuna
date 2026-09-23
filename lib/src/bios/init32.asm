; =============================================================================
; KIZUNA MSXLIB - bios/init32
; inicializa SCREEN 1
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bios_init32
BANK 0

PUBLIC BIOS_INIT32
EXTERN BIOS_CHGMOD

; -----------------------------------------------------------------------------
; BIOS_INIT32: Inicializa modo SCREEN 1 (32 colunas)
; -----------------------------------------------------------------------------
BIOS_INIT32:
    PUSH AF
    LD A, 01h
    CALL BIOS_CHGMOD
    POP AF
    RET

ENDMOD

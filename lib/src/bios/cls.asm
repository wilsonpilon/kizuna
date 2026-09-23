; =============================================================================
; KIZUNA MSXLIB - bios/cls
; limpa a tela
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bios_cls
BANK 0

PUBLIC BIOS_CLS
EXTERN BIOS_CHGMOD

; -----------------------------------------------------------------------------
; BIOS_CLS: Limpa a tela no modo atual
; -----------------------------------------------------------------------------
BIOS_CLS:
    PUSH AF
    XOR A
    CALL BIOS_CHGMOD
    POP AF
    RET

ENDMOD

; =============================================================================
; KIZUNA MSXLIB - bios/posit
; posiciona o cursor (reservado)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bios_posit
BANK 0

PUBLIC BIOS_POSIT

; -----------------------------------------------------------------------------
; BIOS_POSIT: Posiciona o cursor de texto (reservado)
; Entrada: H = coluna (1..X), L = linha (1..Y)
; -----------------------------------------------------------------------------
BIOS_POSIT:
    RET

ENDMOD

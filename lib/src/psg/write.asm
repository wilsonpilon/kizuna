; =============================================================================
; KIZUNA MSXLIB - psg/write
; escreve registrador do PSG
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE psg_write
BANK 0

PUBLIC PSG_Write
INCLUDE "../../inc/psg.inc"

; -----------------------------------------------------------------------------
; PSG_Write: Escreve um valor em um registrador do PSG
; Entrada: A = número do registrador (0..15), E = valor (8 bits)
; -----------------------------------------------------------------------------
PSG_Write:
    OUT (PSG_REG_SEL), A
    LD A, E
    OUT (PSG_DATA_WR), A
    RET

ENDMOD

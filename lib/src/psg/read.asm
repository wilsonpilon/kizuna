; =============================================================================
; KIZUNA MSXLIB - psg/read
; le registrador do PSG
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE psg_read
BANK 0

PUBLIC PSG_Read
INCLUDE "../../inc/psg.inc"

; -----------------------------------------------------------------------------
; PSG_Read: Lê o valor atual de um registrador do PSG
; Entrada: A = número do registrador (0..15)
; Saída: A = valor lido
; -----------------------------------------------------------------------------
PSG_Read:
    OUT (PSG_REG_SEL), A
    IN A, (PSG_DATA_RD)
    RET

ENDMOD

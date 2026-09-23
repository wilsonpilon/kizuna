; =============================================================================
; KIZUNA MSXLIB - vdp/setwriteaddr
; ponteiro de escrita da VRAM
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_setwriteaddr
BANK 0

PUBLIC VDP_SetWriteAddr
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_SetWriteAddr: Configura ponteiro da VRAM para escrita sequencial
; Entrada: HL = endereço de 14 bits da VRAM (0x0000..0x3FFF)
; -----------------------------------------------------------------------------
VDP_SetWriteAddr:
    DI
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OR 40h
    OUT (VDP_CMD), A
    EI
    RET

ENDMOD

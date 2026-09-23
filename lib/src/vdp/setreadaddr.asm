; =============================================================================
; KIZUNA MSXLIB - vdp/setreadaddr
; ponteiro de leitura da VRAM
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_setreadaddr
BANK 0

PUBLIC VDP_SetReadAddr
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_SetReadAddr: Configura ponteiro da VRAM para leitura sequencial
; Entrada: HL = endereço de 14 bits da VRAM (0x0000..0x3FFF)
; -----------------------------------------------------------------------------
VDP_SetReadAddr:
    DI
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OUT (VDP_CMD), A
    EI
    RET

ENDMOD

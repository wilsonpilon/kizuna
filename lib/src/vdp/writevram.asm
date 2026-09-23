; =============================================================================
; KIZUNA MSXLIB - vdp/writevram
; RAM -> VRAM
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_writevram
BANK 0

PUBLIC VDP_WriteVRAM
EXTERN VDP_SetWriteAddr
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_WriteVRAM: Copia bloco de dados da RAM para a VRAM
; Entrada: HL = origem na RAM, DE = destino na VRAM, BC = tamanho em bytes
; -----------------------------------------------------------------------------
VDP_WriteVRAM:
    EX DE, HL
    CALL VDP_SetWriteAddr
    EX DE, HL
VDP_Write_Loop:
    LD A, (HL)
    OUT (VDP_DATA), A
    INC HL
    NOP
    NOP
    NOP
    NOP
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_Write_Loop
    RET

ENDMOD

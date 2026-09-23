; =============================================================================
; KIZUNA MSXLIB - vdp/readvram
; VRAM -> RAM
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_readvram
BANK 0

PUBLIC VDP_ReadVRAM
EXTERN VDP_SetReadAddr
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_ReadVRAM: Copia bloco de dados da VRAM para a RAM
; Entrada: HL = origem na VRAM, DE = destino na RAM, BC = tamanho em bytes
; -----------------------------------------------------------------------------
VDP_ReadVRAM:
    CALL VDP_SetReadAddr
    NOP
    NOP
    NOP
    NOP
VDP_Read_Loop:
    IN A, (VDP_DATA)
    LD (DE), A
    INC DE
    NOP
    NOP
    NOP
    NOP
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_Read_Loop
    RET

ENDMOD

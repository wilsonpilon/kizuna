; =============================================================================
; KIZUNA MSXLIB - vdp/copyvram
; VRAM -> VRAM
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_copyvram
BANK 0

PUBLIC VDP_CopyVRAM
EXTERN VDP_SetReadAddr, VDP_SetWriteAddr
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_CopyVRAM: Copia bloco de VRAM para VRAM diretamente
; Entrada: HL = origem VRAM, DE = destino VRAM, BC = tamanho em bytes
; -----------------------------------------------------------------------------
VDP_CopyVRAM:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
VDP_Copy_Loop:
    CALL VDP_SetReadAddr
    NOP
    NOP
    IN A, (VDP_DATA)
    EX DE, HL
    CALL VDP_SetWriteAddr
    EX DE, HL
    OUT (VDP_DATA), A
    INC HL
    INC DE
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_Copy_Loop
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

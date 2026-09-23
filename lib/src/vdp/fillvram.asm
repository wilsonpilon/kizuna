; =============================================================================
; KIZUNA MSXLIB - vdp/fillvram
; preenche VRAM
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_fillvram
BANK 0

PUBLIC VDP_FillVRAM
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_FillVRAM: Preenche área da VRAM com um byte repetido
; Entrada: HL = endereço inicial VRAM, BC = quantidade de bytes, A = valor
; -----------------------------------------------------------------------------
VDP_FillVRAM:
    PUSH AF
    PUSH DE
    DI
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OR 40h
    OUT (VDP_CMD), A
    POP DE
    POP AF
VDP_Fill_Loop:
    OUT (VDP_DATA), A
    DEC BC
    LD D, A
    LD A, B
    OR C
    LD A, D
    JR NZ, VDP_Fill_Loop
    EI
    RET

ENDMOD

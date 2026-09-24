; =============================================================================
; KIZUNA MSXLIB - vdp/vpoke
; VPOKE e VPEEK de 17 bits
; =============================================================================

MODULE vdp_vpoke
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_VPoke, VDP_VPeek
EXTERN VDP_VramSetWrite, VDP_VramSetRead

; VDP_VPoke: grava um byte na VRAM
; Entrada: D = bit 16 do endereco (0 ou 1), HL = os 16 bits baixos, A = valor
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_VPoke:
    PUSH AF
    LD A, D
    CALL VDP_VramSetWrite
    POP AF
    OUT (PORT_VDP_DATA), A
    RET

; VDP_VPeek: le um byte da VRAM
; Entrada: D = bit 16 do endereco (0 ou 1), HL = os 16 bits baixos
; Saída: A = valor
; Preserva: BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_VPeek:
    LD A, D
    CALL VDP_VramSetRead
    NOP
    NOP
    IN A, (PORT_VDP_DATA)
    RET

ENDMOD

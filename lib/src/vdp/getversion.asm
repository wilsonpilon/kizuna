; =============================================================================
; KIZUNA MSXLIB - vdp/getversion
; identifica o chip de video (V9938, V9958...)
; =============================================================================

MODULE vdp_getversion
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_GetVersion
EXTERN VDP_ReadStatus

; VDP_GetVersion: le o codigo do chip em S#1 (bits 5-1): 0 = V9938, 2 = V9958. Em
; um MSX1 (TMS9918) o status S#1 nao existe e o resultado nao vale.
; Saída: A = codigo do chip
; Preserva: BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_GetVersion:
    LD A, 01h
    CALL VDP_ReadStatus
    AND 3Eh
    RRCA
    RET

ENDMOD

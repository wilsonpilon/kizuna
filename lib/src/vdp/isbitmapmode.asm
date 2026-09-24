; =============================================================================
; KIZUNA MSXLIB - vdp/isbitmapmode
; diz se o modo ajustado e bitmap (SCREEN 5 a 8)
; =============================================================================

MODULE vdp_isbitmapmode
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_IsBitmapMode
EXTERN VDP_Mode

; VDP_IsBitmapMode: 1 se o ultimo modo de VDP_SetMode foi SCREEN 5, 6, 7 ou 8, senao 0
; Saída: A. Preserva: BC, DE, HL. Destrói: flags.
VDP_IsBitmapMode:
    LD A, (VDP_Mode)
    CP 05h
    JR C, VDP_IsBitmapMode_No
    CP 09h
    JR NC, VDP_IsBitmapMode_No
    LD A, 01h
    RET
VDP_IsBitmapMode_No:
    XOR A
    RET

ENDMOD

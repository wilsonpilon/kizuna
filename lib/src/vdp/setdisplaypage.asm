; =============================================================================
; KIZUNA MSXLIB - vdp/setdisplaypage
; escolhe a pagina exibida nos modos bitmap
; =============================================================================

MODULE vdp_setdisplaypage
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetDisplayPage
EXTERN VDP_SetReg

; VDP_SetDisplayPage: qual pagina de VRAM aparece na tela nos modos bitmap
; (SCREEN 5 a 8): R#2 = pagina * 32 + 1Fh. Em SCREEN 5 e 6 ha 4 paginas (0..3), em
; SCREEN 7 e 8, 2 (0..1).
; Entrada: A = pagina
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetDisplayPage:
    PUSH AF
    PUSH BC
    RRCA
    RRCA
    RRCA                ; A << 5
    AND 60h
    OR 1Fh
    LD B, A
    LD C, VDP_REG_NAME
    CALL VDP_SetReg
    POP BC
    POP AF
    RET

ENDMOD

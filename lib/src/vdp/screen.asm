; =============================================================================
; KIZUNA MSXLIB - vdp/screen
; troca de modo de tela (via BIOS_CHGMOD)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_screen
BANK 0

PUBLIC VDP_InitScreen0, VDP_InitScreen1, VDP_InitScreen2, VDP_SetScreen
EXTERN BIOS_CHGMOD

; -----------------------------------------------------------------------------
; VDP_SetScreen: Altera modo de vídeo chamando a rotina oficial CHGMOD da BIOS
; Entrada: A = modo (0 = SCREEN 0, 1 = SCREEN 1, 2 = SCREEN 2)
; -----------------------------------------------------------------------------
VDP_SetScreen:
    JP BIOS_CHGMOD

VDP_InitScreen2:
    LD A, 02h
    JP BIOS_CHGMOD

VDP_InitScreen0:
    XOR A
    JP BIOS_CHGMOD

VDP_InitScreen1:
    LD A, 01h
    JP BIOS_CHGMOD

VDP_CurScreen:
    DB 00h

ENDMOD

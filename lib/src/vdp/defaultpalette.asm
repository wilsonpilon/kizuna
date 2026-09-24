; =============================================================================
; KIZUNA MSXLIB - vdp/defaultpalette
; paleta padrao do MSX2 e paleta do MSX1
; =============================================================================

MODULE vdp_defaultpalette
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetDefaultPalette, VDP_SetMSX1Palette
EXTERN VDP_SetPaletteBlock

; VDP_SetDefaultPalette: as 16 cores padrao do MSX2 (as mesmas do BASIC)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetDefaultPalette:
    PUSH HL
    LD HL, VDP_Pal_MSX2
    CALL VDP_SetPaletteBlock
    POP HL
    RET

; VDP_SetMSX1Palette: as cores do TMS9918 (MSX1), aproximadas em 3 bits por componente
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetMSX1Palette:
    PUSH HL
    LD HL, VDP_Pal_MSX1
    CALL VDP_SetPaletteBlock
    POP HL
    RET

VDP_Pal_MSX2:
    DB 00h, 00h          ;  0: R=0 G=0 B=0
    DB 00h, 00h          ;  1: R=0 G=0 B=0
    DB 11h, 06h          ;  2: R=1 G=6 B=1
    DB 33h, 07h          ;  3: R=3 G=7 B=3
    DB 17h, 01h          ;  4: R=1 G=1 B=7
    DB 27h, 03h          ;  5: R=2 G=3 B=7
    DB 51h, 01h          ;  6: R=5 G=1 B=1
    DB 27h, 06h          ;  7: R=2 G=6 B=7
    DB 71h, 01h          ;  8: R=7 G=1 B=1
    DB 73h, 03h          ;  9: R=7 G=3 B=3
    DB 61h, 06h          ; 10: R=6 G=6 B=1
    DB 64h, 06h          ; 11: R=6 G=6 B=4
    DB 11h, 04h          ; 12: R=1 G=4 B=1
    DB 65h, 02h          ; 13: R=6 G=2 B=5
    DB 55h, 05h          ; 14: R=5 G=5 B=5
    DB 77h, 07h          ; 15: R=7 G=7 B=7
VDP_Pal_MSX1:
    DB 00h, 00h          ;  0: R=0 G=0 B=0
    DB 00h, 00h          ;  1: R=0 G=0 B=0
    DB 11h, 05h          ;  2: R=1 G=5 B=1
    DB 33h, 06h          ;  3: R=3 G=6 B=3
    DB 26h, 02h          ;  4: R=2 G=2 B=6
    DB 37h, 03h          ;  5: R=3 G=3 B=7
    DB 52h, 02h          ;  6: R=5 G=2 B=2
    DB 27h, 06h          ;  7: R=2 G=6 B=7
    DB 62h, 02h          ;  8: R=6 G=2 B=2
    DB 63h, 03h          ;  9: R=6 G=3 B=3
    DB 52h, 05h          ; 10: R=5 G=5 B=2
    DB 63h, 06h          ; 11: R=6 G=6 B=3
    DB 11h, 04h          ; 12: R=1 G=4 B=1
    DB 55h, 02h          ; 13: R=5 G=2 B=5
    DB 55h, 05h          ; 14: R=5 G=5 B=5
    DB 77h, 07h          ; 15: R=7 G=7 B=7

ENDMOD

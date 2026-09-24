; =============================================================================
; KIZUNA MSXLIB - vdp/setpaletteentry
; define uma cor da paleta (V9938/V9958)
; =============================================================================

MODULE vdp_setpaletteentry
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetPaletteEntry
EXTERN VDP_PalShadow

; VDP_SetPaletteEntry: cor `indice` = (R, G, B), cada componente de 0 a 7
; Entrada: A = indice (0..15), B = vermelho, C = verde, D = azul
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetPaletteEntry:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    AND 0Fh
    LD E, A             ; E = indice
    LD A, B
    AND 07h
    ADD A, A
    ADD A, A
    ADD A, A
    ADD A, A            ; R << 4
    LD B, A
    LD A, D
    AND 07h
    OR B
    LD B, A             ; B = (R << 4) | B
    LD A, C
    AND 07h
    LD C, A             ; C = G
    LD HL, VDP_PalShadow
    LD A, E
    ADD A, A
    ADD A, L
    LD L, A
    JR NC, VDP_SetPaletteEntry_Save
    INC H
VDP_SetPaletteEntry_Save:
    LD (HL), B
    INC HL
    LD (HL), C
    DI
    LD A, E
    OUT (PORT_VDP_CMD), A       ; R#16 = indice
    LD A, VDP_REG_PALPTR + 80h
    OUT (PORT_VDP_CMD), A
    LD A, B
    OUT (PORT_VDP_PAL), A
    LD A, C
    OUT (PORT_VDP_PAL), A
    EI
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

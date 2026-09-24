; =============================================================================
; KIZUNA MSXLIB - vdp/setpaletteblock
; define as 16 cores da paleta de uma vez
; =============================================================================

MODULE vdp_setpaletteblock
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetPaletteBlock
EXTERN VDP_PalShadow

; VDP_SetPaletteBlock: carrega a paleta inteira. Cada cor ocupa 2 bytes: primeiro
; (vermelho << 4) | azul, depois o verde (so os 3 bits baixos).
; Entrada: HL = 32 bytes (16 cores)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetPaletteBlock:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD DE, VDP_PalShadow
    LD BC, 0020h
    LDIR                ; guarda a copia sombra
    POP HL
    PUSH HL
    DI
    XOR A
    OUT (PORT_VDP_CMD), A       ; R#16 = 0
    LD A, VDP_REG_PALPTR + 80h
    OUT (PORT_VDP_CMD), A
    LD C, PORT_VDP_PAL
    LD B, 20h
    OTIR
    EI
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

; =============================================================================
; KIZUNA MSXLIB - vdp/spritehide
; esconde um sprite
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_spritehide
BANK 0

PUBLIC VDP_SpriteHide
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_SpriteHide: Oculta um único sprite, movendo seu Y para fora da área
; visível (0E0h), sem afetar os demais sprites
; Entrada: A = índice do sprite (0..31)
; -----------------------------------------------------------------------------
VDP_SpriteHide:
    LD L, A
    LD H, 0
    ADD HL, HL ; x2
    ADD HL, HL ; x4
    LD DE, VDP_SPRITE_ATTR_TABLE
    ADD HL, DE
    DI
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OR 40h
    OUT (VDP_CMD), A
    LD A, 0E0h
    OUT (VDP_DATA), A
    EI
    RET

ENDMOD

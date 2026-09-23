; =============================================================================
; KIZUNA MSXLIB - vdp/spriteset
; posiciona/configura sprite
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_spriteset
BANK 0

PUBLIC VDP_SpriteSet
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_SpriteSet: Posiciona/configura um sprite na Sprite Attribute Table
; Entrada: A = índice do sprite (0..31), H = Y, L = X, D = número do padrão,
;          E = cor (bits 0-3; bit 7 = EC/early-clock, desloca 32 pixels para
;          a esquerda -- usado para X negativo)
; -----------------------------------------------------------------------------
VDP_SpriteSet:
    ; Guarda os 5 valores de entrada em células de rascunho antes de
    ; calcular o endereço -- evita malabarismo de registradores, mesmo
    ; estilo de VDP_PSet_ColorArg. LD (nn),HL grava L em nn e H em nn+1,
    ; por isso VDP_Sprite_X (L=X) e VDP_Sprite_Y (H=Y) estão declarados
    ; nessa ordem, contíguos, logo abaixo.
    LD (VDP_Sprite_Idx), A
    LD (VDP_Sprite_X), HL
    LD A, D
    LD (VDP_Sprite_Pat), A
    LD A, E
    LD (VDP_Sprite_Color), A

    ; endereço = VDP_SPRITE_ATTR_TABLE + índice*4
    LD A, (VDP_Sprite_Idx)
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

    LD A, (VDP_Sprite_Y)
    OUT (VDP_DATA), A
    LD A, (VDP_Sprite_X)
    OUT (VDP_DATA), A
    LD A, (VDP_Sprite_Pat)
    OUT (VDP_DATA), A
    LD A, (VDP_Sprite_Color)
    OUT (VDP_DATA), A
    EI
    RET

VDP_Sprite_Idx:   DB 00h
VDP_Sprite_X:     DB 00h
VDP_Sprite_Y:     DB 00h
VDP_Sprite_Pat:   DB 00h
VDP_Sprite_Color: DB 00h

ENDMOD

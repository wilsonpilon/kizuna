; =============================================================================
; KIZUNA MSXLIB - vdp/spritedefine
; define padrao de sprite
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_spritedefine
BANK 0

PUBLIC VDP_SpriteDefine
EXTERN VDP_WriteVRAM
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_SpriteDefine: Grava o padrão gráfico de um sprite na Sprite Pattern
; Generator Table
; Entrada: A = número do padrão (0..255 para 8x8; para 16x16 o padrão ocupa
;          4 números consecutivos, um por quadrante), HL = ponteiro RAM com
;          os bytes do padrão, BC = tamanho em bytes (8 para 8x8, 32 para
;          16x16)
; -----------------------------------------------------------------------------
VDP_SpriteDefine:
    PUSH HL
    LD H, 0
    LD L, A
    ADD HL, HL ; x2
    ADD HL, HL ; x4
    ADD HL, HL ; x8
    LD DE, VDP_SPRITE_PATTERN_TABLE
    ADD HL, DE
    EX DE, HL  ; DE = endereço VRAM de destino
    POP HL     ; HL = ponteiro RAM de origem (restaurado)
    CALL VDP_WriteVRAM
    RET

ENDMOD

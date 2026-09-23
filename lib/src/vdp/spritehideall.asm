; =============================================================================
; KIZUNA MSXLIB - vdp/spritehideall
; esconde todos os sprites
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_spritehideall
BANK 0

PUBLIC VDP_SpriteHideAll
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_SpriteHideAll: Oculta todos os sprites de uma vez, escrevendo o valor
; terminador (0D0h/208) no Y do sprite de índice 0 -- truque padrão do VDP
; que interrompe o processamento da lista de sprites ali
; -----------------------------------------------------------------------------
VDP_SpriteHideAll:
    PUSH AF
    DI
    LD A, 00h ; byte baixo de VDP_SPRITE_ATTR_TABLE (1B00h)
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, 5Bh ; byte alto (1Bh) OR 40h (comando de escrita em VRAM)
    OUT (VDP_CMD), A
    LD A, 0D0h
    OUT (VDP_DATA), A
    EI
    POP AF
    RET

ENDMOD

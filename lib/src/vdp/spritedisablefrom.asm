; =============================================================================
; KIZUNA MSXLIB - vdp/spritedisablefrom
; marca o fim da lista de sprites
; =============================================================================

MODULE vdp_spritedisablefrom
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SpriteDisableFrom
EXTERN VDP_SpriteAttrAddr, VDP_SpriteModeIs2, VDP_VramSetWrite, VDP_VramPut

; VDP_SpriteDisableFrom: grava o Y "fim da lista" no sprite dado: o VDP nao mostra esse
; sprite nem os seguintes (Y = 0D0h no modo 1, 0D8h no modo 2). Indice 32 ou maior
; nao faz nada.
; Entrada: A = indice do primeiro sprite a desligar
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SpriteDisableFrom:
    CP 20h
    RET NC
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    CALL VDP_SpriteAttrAddr
    PUSH AF
    CALL VDP_SpriteModeIs2
    OR A
    LD B, 0D0h
    JR Z, VDP_SpriteDisableFrom_Go
    LD B, 0D8h
VDP_SpriteDisableFrom_Go:
    POP AF
    CALL VDP_VramSetWrite
    LD A, B
    CALL VDP_VramPut
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

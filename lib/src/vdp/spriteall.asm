; =============================================================================
; KIZUNA MSXLIB - vdp/spriteall
; define de uma vez posicao, padrao e cor de um sprite
; =============================================================================

MODULE vdp_spriteall
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SpriteSetAll
EXTERN VDP_SpriteAttrAddr, VDP_SpriteModeIs2, VDP_SpriteSetColor
EXTERN VDP_VramSetWrite, VDP_VramPut

; VDP_SpriteSetAll: escreve Y, X, padrao e cor. Modo 1: o 4o byte do atributo e a cor;
; modo 2: o 4o byte fica 0 e a cor vai para as 16 linhas da tabela de cores.
; Entrada: A = indice (0..31), H = Y, L = X, D = padrao, E = cor
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SpriteSetAll:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD B, H
    LD C, L
    PUSH AF             ; indice
    PUSH DE
    CALL VDP_SpriteAttrAddr
    CALL VDP_VramSetWrite
    LD A, B
    CALL VDP_VramPut
    LD A, C
    CALL VDP_VramPut
    POP DE
    LD A, D
    CALL VDP_VramPut
    CALL VDP_SpriteModeIs2
    POP HL              ; H = indice
    OR A
    JR NZ, VDP_SpriteSetAll_M2
    LD A, E
    CALL VDP_VramPut
    JR VDP_SpriteSetAll_Done
VDP_SpriteSetAll_M2:
    XOR A
    CALL VDP_VramPut
    LD A, H
    LD B, E
    CALL VDP_SpriteSetColor
VDP_SpriteSetAll_Done:
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

; =============================================================================
; KIZUNA MSXLIB - vdp/spritecolor
; cor de um sprite: modo 1 = 1 cor (atributo), modo 2 = cor por linha (tabela de cores)
; =============================================================================

MODULE vdp_spritecolor
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SpriteSetColor, VDP_SpriteSetLineColors
EXTERN VDP_SpriteAttrAddr, VDP_SpriteColorAddr, VDP_SpriteModeIs2
EXTERN VDP_VramSetWrite, VDP_VramPut, VDP_VramWriteStream

; VDP_SpriteSetColor: pinta o sprite inteiro com uma cor
;   modo 1 (SCREEN 1-3): grava o 4o byte do atributo (bits 0-3 = cor, bit 7 = EC:
;     desloca o sprite 32 pixels para a esquerda)
;   modo 2 (SCREEN 4-8): grava o mesmo byte nas 16 linhas da tabela de cores (bits
;     0-3 = cor, bit 5 = IC, bit 6 = CC, bit 7 = EC)
; Entrada: A = indice (0..31), B = cor
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SpriteSetColor:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    PUSH AF
    CALL VDP_SpriteModeIs2
    OR A
    JR NZ, VDP_SpriteSetColor_M2
    POP AF
    CALL VDP_SpriteAttrAddr
    INC HL
    INC HL
    INC HL
    CALL VDP_VramSetWrite
    LD A, B
    CALL VDP_VramPut
    JR VDP_SpriteSetColor_Done
VDP_SpriteSetColor_M2:
    POP AF
    CALL VDP_SpriteColorAddr
    CALL VDP_VramSetWrite
    LD C, 16
VDP_SpriteSetColor_Loop:
    LD A, B
    CALL VDP_VramPut
    DEC C
    JR NZ, VDP_SpriteSetColor_Loop
VDP_SpriteSetColor_Done:
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_SpriteSetLineColors: as 16 cores de linha de um sprite (so no modo 2; em outros
; modos nao faz nada)
; Entrada: A = indice (0..31), HL = 16 bytes (um por linha, mesmo formato de SetColor)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SpriteSetLineColors:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD D, H
    LD E, L
    PUSH AF
    CALL VDP_SpriteModeIs2
    OR A
    JR Z, VDP_SpriteSetLineColors_Skip
    POP AF
    CALL VDP_SpriteColorAddr
    CALL VDP_VramSetWrite
    EX DE, HL
    LD BC, 0010h
    CALL VDP_VramWriteStream
    JR VDP_SpriteSetLineColors_Done
VDP_SpriteSetLineColors_Skip:
    POP AF
VDP_SpriteSetLineColors_Done:
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

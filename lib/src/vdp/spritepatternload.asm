; =============================================================================
; KIZUNA MSXLIB - vdp/spritepatternload
; carrega padroes de sprite na tabela de padroes atual
; =============================================================================

MODULE vdp_spritepatternload
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SpritePatternLoad
EXTERN VDP_GetSpritePatternTable, VDP_VramSetWrite, VDP_VramWriteStream

; VDP_SpritePatternLoad: copia bytes da RAM para a tabela de padroes de sprite
; (R#6), comecando no padrao dado (cada padrao 8x8 ocupa 8 bytes; um sprite 16x16
; ocupa 32 bytes = 4 numeros seguidos)
; Entrada: A = numero do primeiro padrao, HL = origem na RAM, BC = quantidade de bytes
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SpritePatternLoad:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    PUSH HL
    PUSH BC
    LD L, A
    LD H, 00h
    ADD HL, HL
    ADD HL, HL
    ADD HL, HL          ; HL = padrao * 8
    EX DE, HL
    CALL VDP_GetSpritePatternTable
    ADD HL, DE
    ADC A, 00h
    AND 01h
    CALL VDP_VramSetWrite
    POP BC
    POP HL
    CALL VDP_VramWriteStream
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

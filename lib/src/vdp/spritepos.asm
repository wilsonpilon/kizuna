; =============================================================================
; KIZUNA MSXLIB - vdp/spritepos
; posicao e padrao de um sprite (atributos lidos da tabela de atributos atual)
; =============================================================================

MODULE vdp_spritepos
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SpriteSetPos, VDP_SpriteSetY, VDP_SpriteSetX, VDP_SpriteSetPattern
EXTERN VDP_SpriteAttrAddr, VDP_VramSetWrite, VDP_VramPut

; Os quatro bytes de cada sprite ficam na tabela de atributos (R#5/R#11): Y, X,
; numero do padrao, cor. Valem no modo 1 (SCREEN 1-3) e no modo 2 (SCREEN 4-8).
; Sem VDP_SetMode antes, a tabela e a que estiver nos registradores.
; Preservam: A, BC, DE, HL. Destroem: flags. Reabilitam as interrupcoes (EI).

; VDP_SpriteSetPos: Y e X de um sprite
; Entrada: A = indice (0..31), H = Y, L = X
VDP_SpriteSetPos:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD B, H
    LD C, L
    CALL VDP_SpriteAttrAddr
    CALL VDP_VramSetWrite
    LD A, B
    CALL VDP_VramPut
    LD A, C
    CALL VDP_VramPut
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_SpriteSetY: so o Y
; Entrada: A = indice (0..31), B = Y
VDP_SpriteSetY:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    CALL VDP_SpriteAttrAddr
    CALL VDP_VramSetWrite
    LD A, B
    CALL VDP_VramPut
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_SpriteSetX: so o X
; Entrada: A = indice (0..31), B = X
VDP_SpriteSetX:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    CALL VDP_SpriteAttrAddr
    INC HL
    CALL VDP_VramSetWrite
    LD A, B
    CALL VDP_VramPut
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_SpriteSetPattern: numero do padrao (com sprites 16x16 use multiplos de 4)
; Entrada: A = indice (0..31), B = numero do padrao
VDP_SpriteSetPattern:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    CALL VDP_SpriteAttrAddr
    INC HL
    INC HL
    CALL VDP_VramSetWrite
    LD A, B
    CALL VDP_VramPut
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

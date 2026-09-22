; ==============================================================================
; KIZUNA sample -- Sprites (VDP_SpriteDefine/VDP_SpriteSet/VDP_SpriteHideAll)
; Compilador: KAJI80
; Define um padrão 16x16 (uma bolinha), posiciona o sprite 0 e o move da
; esquerda para a direita da tela em SCREEN 2.
; ==============================================================================

MODULE SPRITEDEMO
BANK 0

PUBLIC Start
EXTERN BIOS_CHGMOD, BIOS_CHGET
EXTERN VDP_SpriteDefine, VDP_SpriteSet, VDP_SpriteHideAll, VDP_SpriteSetSize

BDOS    EQU 0005h
C_WRITE EQU 09h

Start:
    LD DE, MsgIntro
    LD C, C_WRITE
    CALL BDOS

    LD A, 2
    CALL BIOS_CHGMOD ; SCREEN 2

    CALL VDP_SpriteHideAll ; garante que só o nosso sprite fica visível

    LD A, 0E2h ; R#1: 16x16, sem zoom (mesmo default de SCREEN 2)
    CALL VDP_SpriteSetSize

    LD A, 0 ; padrão 0 (ocupa os padrões 0..3, um por quadrante 8x8)
    LD HL, SpritePattern
    LD BC, 32
    CALL VDP_SpriteDefine

    LD B, 20 ; 20 passos de movimento
    LD C, 20 ; X inicial
Sprite_MoveLoop:
    LD A, 0   ; índice do sprite 0
    LD H, 80  ; Y fixo
    LD L, C   ; X = posição atual
    LD D, 0   ; padrão 0
    LD E, 0Fh ; cor 15 (branco)
    CALL VDP_SpriteSet

    PUSH BC
    LD B, 30
Sprite_DelayOuter:
    PUSH BC
    LD B, 00h
Sprite_DelayInner:
    DJNZ Sprite_DelayInner
    POP BC
    DJNZ Sprite_DelayOuter
    POP BC

    LD A, C
    ADD A, 8
    LD C, A
    DJNZ Sprite_MoveLoop

    CALL BIOS_CHGET

    XOR A
    CALL BIOS_CHGMOD ; volta para SCREEN 0

    LD DE, MsgDone
    LD C, C_WRITE
    CALL BDOS

    RET

MsgIntro:
    DB 0Dh, 0Ah
    DB "KIZUNA sample -- Sprites (VDP_SpriteDefine/Set)", 0Dh, 0Ah
    DB "$"

MsgDone:
    DB 0Dh, 0Ah
    DB "Sprite demo concluida.", 0Dh, 0Ah
    DB "$"

; Bolinha 16x16, dividida nos 4 quadrantes 8x8 na ordem que a Sprite
; Pattern Generator Table espera (superior-esquerdo, inferior-esquerdo,
; superior-direito, inferior-direito).
SpritePattern:
    DB 03h, 0Fh, 1Fh, 3Fh, 7Fh, 7Fh, 0FFh, 0FFh
    DB 0FFh, 0FFh, 7Fh, 7Fh, 3Fh, 1Fh, 0Fh, 03h
    DB 0C0h, 0F0h, 0F8h, 0FCh, 0FEh, 0FEh, 0FFh, 0FFh
    DB 0FFh, 0FFh, 0FEh, 0FEh, 0FCh, 0F8h, 0F0h, 0C0h

ENDMOD

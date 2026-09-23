; =============================================================================
; KIZUNA MSXLIB - vdp/spritesetsize
; tamanho dos sprites
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_spritesetsize
BANK 0

PUBLIC VDP_SpriteSetSize
EXTERN VDP_WriteReg

; -----------------------------------------------------------------------------
; VDP_SpriteSetSize: Define tamanho/zoom dos sprites via R#1. Não faz
; leitura-modificação-escrita (o VDP não permite reler um registrador de
; forma simples) -- o chamador fornece o byte COMPLETO do R#1, não só os
; bits de tamanho.
; Entrada: A = byte completo do R#1. Valores prontos para SCREEN 2 (mesma
;          base já usada no projeto, R1=E2h): 0E0h=8x8, 0E1h=8x8+zoom,
;          0E2h=16x16 (padrão), 0E3h=16x16+zoom
; -----------------------------------------------------------------------------
VDP_SpriteSetSize:
    LD B, A
    LD C, 01h
    CALL VDP_WriteReg
    RET

ENDMOD

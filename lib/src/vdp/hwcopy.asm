; =============================================================================
; KIZUNA MSXLIB - vdp/hwcopy
; copia retangulos e linhas dentro da VRAM (LMMM, HMMM, YMMM)
; =============================================================================

MODULE vdp_hwcopy
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_HwCopyRect, VDP_HwMoveRect, VDP_HwCopyLines
EXTERN VDP_CmdRun
EXTERN VDP_Cmd_SX, VDP_Cmd_SY, VDP_Cmd_DX, VDP_Cmd_NY, VDP_Cmd_ARG, VDP_Cmd_CMD

; O motor copia ponto a ponto, comecando no canto (SX,SY) -> (DX,DY) e andando no sentido
; escolhido em ARG. Regioes que se sobrepoem: escolha o sentido de modo que a origem nao
; seja sobrescrita antes de ser lida (VDP_ARG_DIX = 04h anda para a esquerda, VDP_ARG_DIY
; = 08h para cima; nesses casos SX/SY/DX/DY sao o canto de PARTIDA, o mais a direita/baixo).

; VDP_HwCopyRect: copia um retangulo, pixel a pixel (LMMM), com operacao logica
; Entrada: HL = bloco em RAM: SX(2) SY(2) DX(2) DY(2) largura(2) altura(2), A = argumentos
;          (0, VDP_ARG_DIX, VDP_ARG_DIY), B = operacao logica
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HwCopyRect:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_Cmd_ARG), A
    LD A, B
    AND 0Fh
    OR VDP_CMD_LMMM
    LD (VDP_Cmd_CMD), A
    LD DE, VDP_Cmd_SX
    LD BC, 000Ch
    LDIR
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_HwMoveRect: copia por BYTES (HMMM), bem mais rapido; X e largura precisam ser
; multiplos de pixels-por-byte (SCREEN 5: 2, SCREEN 6: 4, SCREEN 7: 2, SCREEN 8: 1)
; Entrada: HL = bloco em RAM (igual ao de VDP_HwCopyRect), A = argumentos
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HwMoveRect:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_Cmd_ARG), A
    LD A, VDP_CMD_HMMM
    LD (VDP_Cmd_CMD), A
    LD DE, VDP_Cmd_SX
    LD BC, 000Ch
    LDIR
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_HwCopyLines: copia linhas inteiras da tela, da coluna DX ate a borda (YMMM)
; Entrada: HL = bloco em RAM: SY(2) DX(2) DY(2) quantidade de linhas(2), A = argumentos
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HwCopyLines:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_Cmd_ARG), A
    LD A, VDP_CMD_YMMM
    LD (VDP_Cmd_CMD), A
    LD DE, VDP_Cmd_SY
    LD BC, 0002h
    LDIR
    LD DE, VDP_Cmd_DX
    LD C, 04h
    LDIR
    LD DE, VDP_Cmd_NY
    LD C, 02h
    LDIR
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

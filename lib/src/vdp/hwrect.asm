; =============================================================================
; KIZUNA MSXLIB - vdp/hwrect
; preenche retangulo pelo motor de comandos (LMMV e HMMV)
; =============================================================================

MODULE vdp_hwrect
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_HwFillRect, VDP_HwFillRectFast
EXTERN VDP_CmdRun
EXTERN VDP_Cmd_SX, VDP_Cmd_DX, VDP_Cmd_CLR, VDP_Cmd_ARG, VDP_Cmd_CMD

; VDP_HwFillRect: preenche um retangulo com uma cor, pixel a pixel (LMMV), com operacao logica
; Entrada: HL = retangulo em RAM: X(2) Y(2) largura(2) altura(2), A = cor, B = operacao logica
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HwFillRect:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_Cmd_CLR), A
    LD A, B
    AND 0Fh
    OR VDP_CMD_LMMV
    LD (VDP_Cmd_CMD), A
    XOR A
    LD (VDP_Cmd_ARG), A
    LD DE, VDP_Cmd_DX
    LD BC, 0008h
    LDIR
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_HwFillRectFast: preenche um retangulo por BYTES (HMMV): muito mais rapido, mas X e a
; largura contam em pixels e viram bytes (SCREEN 5: 2 pixels por byte; SCREEN 6: 4;
; SCREEN 7: 2; SCREEN 8: 1) -- use X e largura multiplos disso. A cor e o byte inteiro
; (SCREEN 5: 0x11 * cor pinta os dois pixels com a mesma cor). Sem operacao logica.
; Entrada: HL = retangulo em RAM: X(2) Y(2) largura(2) altura(2), A = byte de cor
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HwFillRectFast:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_Cmd_CLR), A
    LD A, VDP_CMD_HMMV
    LD (VDP_Cmd_CMD), A
    XOR A
    LD (VDP_Cmd_ARG), A
    LD DE, VDP_Cmd_DX
    LD BC, 0008h
    LDIR
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD

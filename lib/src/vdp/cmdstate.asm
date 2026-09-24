; =============================================================================
; KIZUNA MSXLIB - vdp/cmdstate
; bloco de parametros do motor de comandos (espelho de R#32 a R#46)
; =============================================================================

MODULE vdp_cmdstate
BANK 0


PUBLIC VDP_Cmd_SX, VDP_Cmd_SY, VDP_Cmd_DX, VDP_Cmd_DY, VDP_Cmd_NX, VDP_Cmd_NY
PUBLIC VDP_Cmd_CLR, VDP_Cmd_ARG, VDP_Cmd_CMD

; Os 15 bytes abaixo ficam na mesma ordem dos registradores R#32..R#46 do V9938; as
; rotinas VDP_Hw* preenchem o que o comando usa e chamam VDP_CmdRun com HL = VDP_Cmd_SX.
; Voce tambem pode montar um bloco proprio (15 bytes nessa ordem) e chamar VDP_CmdRun.
VDP_Cmd_SX:
    DS 2                ; R#32-33 origem X
VDP_Cmd_SY:
    DS 2                ; R#34-35 origem Y
VDP_Cmd_DX:
    DS 2                ; R#36-37 destino X
VDP_Cmd_DY:
    DS 2                ; R#38-39 destino Y
VDP_Cmd_NX:
    DS 2                ; R#40-41 largura
VDP_Cmd_NY:
    DS 2                ; R#42-43 altura
VDP_Cmd_CLR:
    DS 1                ; R#44 cor / dado
VDP_Cmd_ARG:
    DS 1                ; R#45 argumentos (direcao, memoria expandida)
VDP_Cmd_CMD:
    DS 1                ; R#46 comando (nibble alto) e operacao logica (baixo)

ENDMOD

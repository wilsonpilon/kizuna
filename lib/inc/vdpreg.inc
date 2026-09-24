; =============================================================================
; KIZUNA MSXLIB - registradores e comandos do VDP (TMS9918 / V9938 / V9958)
; Nomes VDP_REG_<n> (numero do registrador), VDP_R<n>_<bit> (mascara de bit),
; VDP_CMD_* (comandos do motor V9938) e VDP_LOP_* (operacoes logicas).
; Conferido contra a folha de dados do V9938 e contra o guia vdp_reg.h do MSXgl.
; =============================================================================

; --- Numero de cada registrador ---------------------------------------------
VDP_REG_MODE0    EQU 0      ; bits de modo (M3-M5), IE1, IE2, DG
VDP_REG_MODE1    EQU 1      ; MAG, ST, M1, M2, IE0, BL, K16
VDP_REG_NAME     EQU 2      ; tabela de nomes (layout)  = endereco >> 10
VDP_REG_COLOR    EQU 3      ; tabela de cores, byte baixo
VDP_REG_PATTERN  EQU 4      ; tabela de padroes         = endereco >> 11
VDP_REG_SPRATTR  EQU 5      ; tabela de atributos de sprite, byte baixo (>> 7)
VDP_REG_SPRPAT   EQU 6      ; tabela de padroes de sprite = endereco >> 11
VDP_REG_COLORS   EQU 7      ; cor de texto (nibble alto) e de fundo/borda (baixo)
VDP_REG_MODE2    EQU 8      ; BW, SPD, VR, CB, TP, PL, MS
VDP_REG_MODE3    EQU 9      ; DC, NT, EO, IL, S0, S1, LN
VDP_REG_COLORH   EQU 10     ; tabela de cores, bits altos
VDP_REG_SPRATTRH EQU 11     ; tabela de atributos de sprite, bits altos
VDP_REG_BLINKCOL EQU 12     ; cores de piscar (modo texto 2)
VDP_REG_BLINKTIM EQU 13     ; tempos de piscar
VDP_REG_VRAMHI   EQU 14     ; bits A16-A14 do endereco de VRAM
VDP_REG_STATSEL  EQU 15     ; seleciona qual registrador de status (S#n) sera lido
VDP_REG_PALPTR   EQU 16     ; ponteiro de paleta
VDP_REG_INDIRECT EQU 17     ; ponteiro de escrita indireta (bit 7 = sem auto-incremento)
VDP_REG_ADJUST   EQU 18     ; ajuste de posicao da tela (X e Y)
VDP_REG_HINTLINE EQU 19     ; linha da interrupcao horizontal
VDP_REG_VOFFSET  EQU 23     ; deslocamento vertical (scroll de hardware)
VDP_REG_MODE4    EQU 25     ; V9958: CMD, VDS, YAE, YJK, WTE, MSK, SP2
VDP_REG_HSCROLLH EQU 26     ; V9958: scroll horizontal, parte alta
VDP_REG_HSCROLLL EQU 27     ; V9958: scroll horizontal, parte baixa (0-7)
; Motor de comandos (V9938/V9958): pares X/Y de 16 bits, byte baixo primeiro
VDP_REG_SXL      EQU 32     ; origem X (32-33)
VDP_REG_SYL      EQU 34     ; origem Y (34-35)
VDP_REG_DXL      EQU 36     ; destino X (36-37)
VDP_REG_DYL      EQU 38     ; destino Y (38-39)
VDP_REG_NXL      EQU 40     ; largura (40-41)
VDP_REG_NYL      EQU 42     ; altura (42-43)
VDP_REG_CLR      EQU 44     ; cor / dado
VDP_REG_ARG      EQU 45     ; argumentos (direcao, memoria expandida)
VDP_REG_CMD      EQU 46     ; comando (nibble alto) + operacao logica (baixo)

; --- R#0 ---------------------------------------------------------------------
VDP_R0_EV   EQU 01h         ; entrada de video externa
VDP_R0_M3   EQU 02h
VDP_R0_M4   EQU 04h
VDP_R0_M5   EQU 08h
VDP_R0_IE1  EQU 10h         ; interrupcao horizontal
VDP_R0_IE2  EQU 20h         ; light pen
VDP_R0_DG   EQU 40h         ; digitalizacao

; --- R#1 ---------------------------------------------------------------------
VDP_R1_MAG  EQU 01h         ; sprites ampliados (2x)
VDP_R1_ST   EQU 02h         ; sprites 16x16
VDP_R1_M1   EQU 08h
VDP_R1_M2   EQU 10h
VDP_R1_IE0  EQU 20h         ; interrupcao vertical
VDP_R1_BL   EQU 40h         ; tela habilitada
VDP_R1_K16  EQU 80h         ; 16 KB de VRAM (MSX1)

; --- R#8 ---------------------------------------------------------------------
VDP_R8_BW   EQU 01h         ; preto e branco
VDP_R8_SPD  EQU 02h         ; sprites desabilitados
VDP_R8_VR   EQU 08h         ; VRAM de 64 Kbit (chips 4464)
VDP_R8_CB   EQU 10h         ; entrada de cor
VDP_R8_TP   EQU 20h         ; cor 0 opaca (nao transparente)
VDP_R8_PL   EQU 40h         ; light pen
VDP_R8_MS   EQU 80h         ; mouse

; --- R#9 ---------------------------------------------------------------------
VDP_R9_DC   EQU 01h
VDP_R9_NT   EQU 02h         ; 1 = 50 Hz (PAL)
VDP_R9_EO   EQU 04h         ; alterna paginas par/impar
VDP_R9_IL   EQU 08h         ; entrelacado
VDP_R9_S0   EQU 10h
VDP_R9_S1   EQU 20h
VDP_R9_LN   EQU 80h         ; 212 linhas (0 = 192)

; --- R#25 (V9958) ------------------------------------------------------------
VDP_R25_SP2 EQU 01h
VDP_R25_MSK EQU 02h         ; esconde a coluna mais a esquerda (8 pixels)
VDP_R25_WTE EQU 04h
VDP_R25_YJK EQU 08h
VDP_R25_YAE EQU 10h
VDP_R25_VDS EQU 20h
VDP_R25_CMD EQU 40h         ; comandos de logica tambem em SCREEN 1-4

; --- Comandos do motor do V9938 (valor do nibble alto de R#46) --------------
VDP_CMD_STOP  EQU 00h
VDP_CMD_POINT EQU 40h
VDP_CMD_PSET  EQU 50h
VDP_CMD_SRCH  EQU 60h
VDP_CMD_LINE  EQU 70h
VDP_CMD_LMMV  EQU 80h
VDP_CMD_LMMM  EQU 90h
VDP_CMD_LMCM  EQU 0A0h
VDP_CMD_LMMC  EQU 0B0h
VDP_CMD_HMMV  EQU 0C0h
VDP_CMD_HMMM  EQU 0D0h
VDP_CMD_YMMM  EQU 0E0h
VDP_CMD_HMMC  EQU 0F0h

; --- Operacoes logicas (nibble baixo de R#46) -------------------------------
VDP_LOP_IMP   EQU 00h       ; copia
VDP_LOP_AND   EQU 01h
VDP_LOP_OR    EQU 02h
VDP_LOP_XOR   EQU 03h
VDP_LOP_NOT   EQU 04h
VDP_LOP_TIMP  EQU 08h       ; as variantes T ignoram a cor 0 (transparente)
VDP_LOP_TAND  EQU 09h
VDP_LOP_TOR   EQU 0Ah
VDP_LOP_TXOR  EQU 0Bh
VDP_LOP_TNOT  EQU 0Ch

; --- R#45: argumentos dos comandos ------------------------------------------
VDP_ARG_DIX     EQU 04h     ; 1 = X decrescente
VDP_ARG_DIY     EQU 08h     ; 1 = Y decrescente
VDP_ARG_MXS     EQU 10h     ; origem na memoria expandida
VDP_ARG_MXD     EQU 20h     ; destino na memoria expandida

; --- Registradores de status -------------------------------------------------
; S#0
VDP_S0_F    EQU 80h         ; quadro (interrupcao vertical pendente)
VDP_S0_5S   EQU 40h         ; quinto sprite na mesma linha
VDP_S0_C    EQU 20h         ; colisao de sprites
; S#1
VDP_S1_FH   EQU 01h         ; interrupcao horizontal pendente
; S#2
VDP_S2_CE   EQU 01h         ; comando em execucao
VDP_S2_EO   EQU 02h         ; campo par/impar
VDP_S2_BD   EQU 10h         ; borda detectada (SRCH)
VDP_S2_HR   EQU 20h         ; retraco horizontal
VDP_S2_VR   EQU 40h         ; retraco vertical
VDP_S2_TR   EQU 80h         ; pronto para transferir um byte (LMMC/HMMC/LMCM)

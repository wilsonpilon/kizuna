; ==============================================================================
; KIZUNA sample -- Motor de comandos do VDP (VDP_Hw*) em SCREEN 5
; Compilador: KAJI80
; Teste PARA HARDWARE / openMSX (MSX2 ou superior). Desenha na tela e confere a
; transferencia RAM <-> VRAM. O que olhar na tela:
;   1. moldura branca e duas diagonais (vermelha e ciano) de canto a canto
;   2. LINHA HORIZONTAL verde de (120,60) a (140,60) com um ponto branco encostado
;      em cada ponta (X=119 e X=141): NAO pode haver vao nem sobreposicao (a linha
;      tem que ter exatamente 21 pontos, do X=120 ao X=140)
;   3. o mesmo na vertical, em X=60, de Y=100 a Y=120, pontos brancos em Y=99 e Y=121
;   4. um retangulo cheio azul e uma copia dele ao lado
;   5. um quadradinho 8x8 de degrade (16 cores) no canto inferior direito
; Ao sair da SCREEN 5, o texto diz se a leitura de volta da VRAM bateu com o que
; foi escrito. Confirme no relatorio 1. e 2. e 3. (sao os pontos nao verificados).
; ==============================================================================

MODULE VDPCMDDEMO
BANK 0

PUBLIC Start
EXTERN BIOS_CHGMOD, BIOS_CHGET
EXTERN VDP_HwFillRect, VDP_HwBox, VDP_HwLine, VDP_HwPlot, VDP_HwBoxFill
EXTERN VDP_HwCopyRect, VDP_HwLoadRect, VDP_HwReadRect, VDP_CmdWait

BDOS    EQU 0005h
C_WRITE EQU 09h

Start:
    LD DE, MsgIntro
    LD C, C_WRITE
    CALL BDOS

    LD A, 5
    CALL BIOS_CHGMOD        ; SCREEN 5

    ; fundo
    LD HL, RectAll
    LD A, 1
    LD B, 0
    CALL VDP_HwFillRect

    ; moldura e diagonais
    LD HL, BoxFrame
    LD A, 15
    LD B, 0
    CALL VDP_HwBox
    LD HL, DiagA
    LD A, 8
    LD B, 0
    CALL VDP_HwLine
    LD HL, DiagB
    LD A, 7
    LD B, 0
    CALL VDP_HwLine

    ; linha horizontal de 21 pontos com marcas nas pontas
    LD HL, LineH
    LD A, 12
    LD B, 0
    CALL VDP_HwLine
    LD BC, 119
    LD DE, 60
    LD A, 15
    LD H, 0
    CALL VDP_HwPlot
    LD BC, 141
    LD DE, 60
    LD A, 15
    LD H, 0
    CALL VDP_HwPlot

    ; linha vertical de 21 pontos com marcas nas pontas
    LD HL, LineV
    LD A, 12
    LD B, 0
    CALL VDP_HwLine
    LD BC, 60
    LD DE, 99
    LD A, 15
    LD H, 0
    CALL VDP_HwPlot
    LD BC, 60
    LD DE, 121
    LD A, 15
    LD H, 0
    CALL VDP_HwPlot

    ; retangulo cheio e copia
    LD HL, BoxBlue
    LD A, 4
    LD B, 0
    CALL VDP_HwBoxFill
    LD HL, CopyBlock
    LD A, 0
    LD B, 0
    CALL VDP_HwCopyRect

    ; degrade 8x8: RAM -> VRAM e de volta para RAM
    LD HL, RectGrad
    LD DE, Pixels
    LD BC, 64
    LD A, 0
    CALL VDP_HwLoadRect
    LD HL, RectGrad
    LD DE, ReadBuf
    LD BC, 64
    CALL VDP_HwReadRect
    CALL VDP_CmdWait

    ; compara
    LD HL, Pixels
    LD DE, ReadBuf
    LD B, 64
    LD C, 0                 ; C = 0: bateu
Cmp_Loop:
    LD A, (DE)
    CP (HL)
    JR Z, Cmp_Same
    LD C, 1
Cmp_Same:
    INC HL
    INC DE
    DJNZ Cmp_Loop
    LD A, C
    LD (Result), A

    CALL BIOS_CHGET         ; espera uma tecla

    XOR A
    CALL BIOS_CHGMOD        ; volta para SCREEN 0

    LD DE, MsgOk
    LD A, (Result)
    OR A
    JR Z, Print_Result
    LD DE, MsgFail
Print_Result:
    LD C, C_WRITE
    CALL BDOS
    RET

Result:
    DB 0

MsgIntro:
    DB 0Dh, 0Ah
    DB "KIZUNA sample -- motor de comandos do VDP (SCREEN 5)", 0Dh, 0Ah
    DB "Confira a tela e aperte uma tecla.", 0Dh, 0Ah
    DB "$"

MsgOk:
    DB 0Dh, 0Ah
    DB "Leitura de volta (LMMC -> LMCM): OK.", 0Dh, 0Ah
    DB "$"

MsgFail:
    DB 0Dh, 0Ah
    DB "Leitura de volta (LMMC -> LMCM): DIFERENTE!", 0Dh, 0Ah
    DB "$"

; retangulos e linhas: x, y, largura, altura / x1, y1, x2, y2
RectAll:
    DW 0, 0, 256, 212
BoxFrame:
    DW 4, 4, 251, 207
DiagA:
    DW 4, 4, 251, 207
DiagB:
    DW 251, 4, 4, 207
LineH:
    DW 120, 60, 140, 60
LineV:
    DW 60, 100, 60, 120
BoxBlue:
    DW 30, 130, 90, 170
; SX, SY, DX, DY, largura, altura (copia do retangulo azul: 61 x 41)
CopyBlock:
    DW 30, 130, 110, 130, 61, 41
RectGrad:
    DW 230, 190, 8, 8

Pixels:
    DB 0, 1, 2, 3, 4, 5, 6, 7
    DB 1, 2, 3, 4, 5, 6, 7, 8
    DB 2, 3, 4, 5, 6, 7, 8, 9
    DB 3, 4, 5, 6, 7, 8, 9, 10
    DB 4, 5, 6, 7, 8, 9, 10, 11
    DB 5, 6, 7, 8, 9, 10, 11, 12
    DB 6, 7, 8, 9, 10, 11, 12, 13
    DB 7, 8, 9, 10, 11, 12, 13, 14

ReadBuf:
    DS 64

ENDMOD

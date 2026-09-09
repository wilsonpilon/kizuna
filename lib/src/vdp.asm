; =============================================================================
; KIZUNA MSXLIB - VDP.ASM
; Rotinas de controle do processador de vídeo TMS9918 / V9938 / V9958
; Compatível com MSX-DOS 1 / 2 / Nextor sem dependência de ROM BIOS
; =============================================================================

MODULE VDP
BANK 0

PUBLIC VDP_WriteReg, VDP_SetWriteAddr, VDP_SetReadAddr
PUBLIC VDP_FillVRAM, VDP_WriteVRAM, VDP_ReadVRAM, VDP_CopyVRAM, VDP_SetColor
PUBLIC VDP_SetScreen, VDP_InitScreen2, VDP_InitScreen1, VDP_InitScreen0, VDP_InitScreen2_Tables
PUBLIC VDP_PSet, VDP_Line, VDP_BoxFill
EXTERN BIOS_CHGMOD

VDP_DATA EQU 0098h
VDP_CMD  EQU 0099h

; -----------------------------------------------------------------------------
; VDP_WriteReg: Escreve valor em um registrador do VDP
; Entrada: C = número do registrador (0..23), B = valor a escrever
; -----------------------------------------------------------------------------
VDP_WriteReg:
    DI
    LD A, B
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, C
    OR 80h
    OUT (VDP_CMD), A
    EI
    RET

; -----------------------------------------------------------------------------
; VDP_SetWriteAddr: Configura ponteiro da VRAM para escrita sequencial
; Entrada: HL = endereço de 14 bits da VRAM (0x0000..0x3FFF)
; -----------------------------------------------------------------------------
VDP_SetWriteAddr:
    DI
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OR 40h
    OUT (VDP_CMD), A
    EI
    RET

; -----------------------------------------------------------------------------
; VDP_SetReadAddr: Configura ponteiro da VRAM para leitura sequencial
; Entrada: HL = endereço de 14 bits da VRAM (0x0000..0x3FFF)
; -----------------------------------------------------------------------------
VDP_SetReadAddr:
    DI
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OUT (VDP_CMD), A
    EI
    RET

; -----------------------------------------------------------------------------
; VDP_FillVRAM: Preenche área da VRAM com um byte repetido
; Entrada: HL = endereço inicial VRAM, BC = quantidade de bytes, A = valor
; -----------------------------------------------------------------------------
VDP_FillVRAM:
    PUSH AF
    PUSH DE
    DI
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OR 40h
    OUT (VDP_CMD), A
    POP DE
    POP AF
VDP_Fill_Loop:
    OUT (VDP_DATA), A
    DEC BC
    LD D, A
    LD A, B
    OR C
    LD A, D
    JR NZ, VDP_Fill_Loop
    EI
    RET

; -----------------------------------------------------------------------------
; VDP_WriteVRAM: Copia bloco de dados da RAM para a VRAM
; Entrada: HL = origem na RAM, DE = destino na VRAM, BC = tamanho em bytes
; -----------------------------------------------------------------------------
VDP_WriteVRAM:
    EX DE, HL
    CALL VDP_SetWriteAddr
    EX DE, HL
VDP_Write_Loop:
    LD A, (HL)
    OUT (VDP_DATA), A
    INC HL
    NOP
    NOP
    NOP
    NOP
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_Write_Loop
    RET

; -----------------------------------------------------------------------------
; VDP_ReadVRAM: Copia bloco de dados da VRAM para a RAM
; Entrada: HL = origem na VRAM, DE = destino na RAM, BC = tamanho em bytes
; -----------------------------------------------------------------------------
VDP_ReadVRAM:
    CALL VDP_SetReadAddr
    NOP
    NOP
    NOP
    NOP
VDP_Read_Loop:
    IN A, (VDP_DATA)
    LD (DE), A
    INC DE
    NOP
    NOP
    NOP
    NOP
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_Read_Loop
    RET

; -----------------------------------------------------------------------------
; VDP_CopyVRAM: Copia bloco de VRAM para VRAM diretamente
; Entrada: HL = origem VRAM, DE = destino VRAM, BC = tamanho em bytes
; -----------------------------------------------------------------------------
VDP_CopyVRAM:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
VDP_Copy_Loop:
    CALL VDP_SetReadAddr
    NOP
    NOP
    IN A, (VDP_DATA)
    EX DE, HL
    CALL VDP_SetWriteAddr
    EX DE, HL
    OUT (VDP_DATA), A
    INC HL
    INC DE
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_Copy_Loop
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; -----------------------------------------------------------------------------
; VDP_SetColor: Define as cores de texto (frente) e fundo (Reg 7)
; Entrada: H = cor de frente (0..15), L = cor de fundo (0..15)
; -----------------------------------------------------------------------------
VDP_SetColor:
    LD A, H
    SLA A
    SLA A
    SLA A
    SLA A
    AND F0h
    LD B, A
    LD A, L
    AND 0Fh
    OR B
    LD B, A
    LD C, 07h
    CALL VDP_WriteReg
    RET

; -----------------------------------------------------------------------------
; VDP_SetScreen: Altera modo de vídeo chamando a rotina oficial CHGMOD da BIOS
; Entrada: A = modo (0 = SCREEN 0, 1 = SCREEN 1, 2 = SCREEN 2)
; -----------------------------------------------------------------------------
VDP_SetScreen:
    JP BIOS_CHGMOD

VDP_InitScreen2:
    LD A, 02h
    JP BIOS_CHGMOD

VDP_InitScreen0:
    XOR A
    JP BIOS_CHGMOD

VDP_InitScreen1:
    LD A, 01h
    JP BIOS_CHGMOD

VDP_CurScreen:
    DB 00h

; -----------------------------------------------------------------------------
; VDP_InitScreen2_Tables: Inicializa tabelas VRAM essenciais para SCREEN 2:
; 1. Pattern Name Table (1800h..1AFFh): 3 blocos de 00h..FFh (768 bytes)
; 2. Pattern Generator Table (0000h..17FFh): limpa com 00h (6144 bytes)
; 3. Color Table (2000h..37FFh): preenche com 0F1h (Frente Branco, Fundo Preto)
; 4. Sprite Attribute Table (1B00h): desativa sprites colocando Y=208
; -----------------------------------------------------------------------------
VDP_InitScreen2_Tables:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL

    ; Reaplica a configuração de SCREEN 2 diretamente no VDP.
    ; A MSXgl usa estes mesmos endereços: NT=1800h, CT=2000h, PT=0000h.
    LD B, 02h
    LD C, 00h
    CALL VDP_WriteReg
    LD B, 0E0h
    LD C, 01h
    CALL VDP_WriteReg
    LD B, 06h
    LD C, 02h
    CALL VDP_WriteReg
    LD B, 0FFh
    LD C, 03h
    CALL VDP_WriteReg
    LD B, 03h
    LD C, 04h
    CALL VDP_WriteReg
    LD B, 36h
    LD C, 05h
    CALL VDP_WriteReg
    LD B, 07h
    LD C, 06h
    CALL VDP_WriteReg
    ; Desabilita sprites durante os testes de SCREEN 2.
    LD B, 02h
    LD C, 08h
    CALL VDP_WriteReg

    ; 1. Inicializa a Pattern Name Table com 00h..FFh em cada faixa.
    LD HL, 1800h
    DI
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OR 40h
    OUT (VDP_CMD), A
    LD D, 03h
VDP_Init2_BlockLoop:
    XOR A
VDP_Init2_ByteLoop:
    OUT (VDP_DATA), A
    NOP
    NOP
    NOP
    NOP
    INC A
    JR NZ, VDP_Init2_ByteLoop
    DEC D
    JR NZ, VDP_Init2_BlockLoop
    EI

    ; 2. Limpa Pattern Generator Table (0000h..17FFh, 6144 bytes) com 00h
    LD HL, 0000h
    LD BC, 1800h
    XOR A
    CALL VDP_FillVRAM

    ; 3. Preenche Color Table (2000h..37FFh, 6144 bytes) com 0F1h (Frente 15 Branco, Fundo 1 Preto)
    LD HL, 2000h
    LD BC, 1800h
    LD A, 0F1h
    CALL VDP_FillVRAM

    ; 4. Desativa todos os sprites (Y=208 em toda a SAT)
    LD HL, 1B00h
    LD BC, 0080h
    LD A, 0D0h
    CALL VDP_FillVRAM

    POP HL
    POP DE
    POP BC
    POP AF
    RET

; -----------------------------------------------------------------------------
; VDP_PSet: Plota um pixel no modo gráfico SCREEN 2 (256x192)
; Entrada: BC = X (0..255), DE = Y (0..191), A = Cor (0..15)
; Preserva: BC, DE, HL
; -----------------------------------------------------------------------------
VDP_PSet:
    PUSH BC
    PUSH DE
    PUSH HL

    ; HL = endereço do padrão selecionado pela Name Table.
    ; SCREEN 2 possui uma página de padrões para cada faixa vertical de 64 px.
    ; Índice = ((Y >> 3) * 32 + (X >> 3)) & 0xFF; endereço = página + índice * 8 + (Y & 7).
    LD A, E
    AND 0C0h
    RRCA
    RRCA
    RRCA
    LD D, A
    LD A, E
    RRCA
    RRCA
    RRCA
    AND 07h
    ADD A, A
    ADD A, A
    ADD A, A
    ADD A, A
    ADD A, A
    LD L, A
    LD A, C
    SRL A
    SRL A
    SRL A
    ADD A, L
    LD L, A
    LD H, 00h
    SLA L
    RL H
    SLA L
    RL H
    SLA L
    RL H
    LD A, E
    AND 07h
    ADD A, L
    LD L, A
    LD A, D
    ADD A, H
    LD H, A

    ; Bitmask: bit 7 - (X & 7)
    LD A, C
    AND 07h
    LD B, A
    LD A, 80h
VDP_PSet_MaskLoop:
    DEC B
    JP M, VDP_PSet_MaskDone
    RRCA
    JR VDP_PSet_MaskLoop
VDP_PSet_MaskDone:
    LD C, A

    ; Escreve diretamente a mascara no byte do padrao.
    ; Nesta etapa cada ponto ocupa uma celula 8x8 independente.
    CALL VDP_SetWriteAddr
    NOP
    NOP
    NOP
    NOP
    NOP
    NOP
    NOP
    LD A, C
    OUT (VDP_DATA), A

    ; A Color Table ja foi inicializada com F1h (branco sobre preto).
    ; A atualizacao de cor por pixel sera adicionada depois do PSET basico.

    POP HL
    POP DE
    POP BC
    RET

; -----------------------------------------------------------------------------
; VDP_BoxFill: Preenche uma região retangular na tela (SCREEN 2)
; Pilha (caller cleanup): X1, Y1, X2, Y2, Color (16-bit cada)
; -----------------------------------------------------------------------------
VDP_BoxFill:
    PUSH IX
    LD IX, 0000h
    ADD IX, SP
    PUSH BC
    PUSH DE
    PUSH HL

    ; Parâmetros relativos a IX:
    ; IX+4: Color
    ; IX+6: Y2
    ; IX+8: X2
    ; IX+10: Y1
    ; IX+12: X1

    ; Otimização: Se for tela cheia (0,0)-(255,191), usa VDP_FillVRAM
    LD A, (IX+12)
    OR A
    JR NZ, VDP_BoxFill_General
    LD A, (IX+10)
    OR A
    JR NZ, VDP_BoxFill_General
    LD A, (IX+8)
    CP 0FFh
    JR C, VDP_BoxFill_General
    LD A, (IX+6)
    CP 0BFh
    JR C, VDP_BoxFill_General

    ; Limpa tela inteira
    LD HL, 0000h
    LD BC, 1800h
    XOR A
    CALL VDP_FillVRAM

    LD A, (IX+4)
    AND 0Fh
    LD B, A
    SLA A
    SLA A
    SLA A
    SLA A
    OR B
    LD HL, 2000h
    LD BC, 1800h
    CALL VDP_FillVRAM
    JR VDP_BoxFill_Exit

VDP_BoxFill_General:
    ; Preenchimento por linhas e colunas
    LD E, (IX+10) ; Y = Y1
VDP_BoxFill_YLoop:
    LD C, (IX+12) ; X = X1
VDP_BoxFill_XLoop:
    LD B, 00h
    LD D, 00h
    LD A, (IX+4)
    CALL VDP_PSet
    INC C
    LD A, C
    CP (IX+8)
    JR C, VDP_BoxFill_XLoop
    JR Z, VDP_BoxFill_XLast
    JR VDP_BoxFill_YNext
VDP_BoxFill_XLast:
    LD B, 00h
    LD D, 00h
    LD A, (IX+4)
    CALL VDP_PSet
VDP_BoxFill_YNext:
    INC E
    LD A, E
    CP (IX+6)
    JR C, VDP_BoxFill_YLoop
    JR Z, VDP_BoxFill_YLast
    JR VDP_BoxFill_Exit
VDP_BoxFill_YLast:
    LD C, (IX+12)
VDP_BoxFill_XLastLoop:
    LD B, 00h
    LD D, 00h
    LD A, (IX+4)
    CALL VDP_PSet
    INC C
    LD A, C
    CP (IX+8)
    JR C, VDP_BoxFill_XLastLoop
    JR Z, VDP_BoxFill_XLastFinal
    JR VDP_BoxFill_Exit
VDP_BoxFill_XLastFinal:
    LD B, 00h
    LD D, 00h
    LD A, (IX+4)
    CALL VDP_PSet

VDP_BoxFill_Exit:
    POP HL
    POP DE
    POP BC
    POP IX
    RET

; -----------------------------------------------------------------------------
; VDP_Line: Desenha uma linha reta entre (X1, Y1) e (X2, Y2) via Bresenham
; Pilha (caller cleanup): X1, Y1, X2, Y2, Color (16-bit cada)
; -----------------------------------------------------------------------------
VDP_Line:
    PUSH IX
    LD IX, 0000h
    ADD IX, SP
    PUSH BC
    PUSH DE
    PUSH HL

    ; Se horizontal puro (Y1 == Y2): otimização rápida
    LD A, (IX+10)
    CP (IX+6)
    JR Z, VDP_Line_H_Start

    ; Se vertical puro (X1 == X2): otimização rápida
    LD A, (IX+12)
    CP (IX+8)
    JR Z, VDP_Line_V_Start

    JR VDP_Line_General

VDP_Line_H_Start:
    ; Linha horizontal de min(X1,X2) até max(X1,X2)
    LD C, (IX+12) ; X1
    LD A, (IX+8)  ; X2
    CP C
    JR NC, VDP_Line_H_X1_Le_X2
    LD C, (IX+8)
    LD A, (IX+12)
VDP_Line_H_X1_Le_X2:
    LD (VDP_Line_H_End), A
    LD E, (IX+10) ; Y
    LD D, 00h
VDP_Line_HLoop:
    LD B, 00h
    LD A, (IX+4)  ; Cor
    CALL VDP_PSet
    LD A, (VDP_Line_H_End)
    CP C
    JP Z, VDP_Line_Exit
    INC C
    JR VDP_Line_HLoop

VDP_Line_V_Start:
    ; Linha vertical de min(Y1,Y2) até max(Y1,Y2)
    LD E, (IX+10) ; Y1
    LD A, (IX+6)  ; Y2
    CP E
    JR NC, VDP_Line_V_Y1_Le_Y2
    LD E, (IX+6)
    LD A, (IX+10)
VDP_Line_V_Y1_Le_Y2:
    LD (VDP_Line_H_End), A
    LD C, (IX+12) ; X
    LD B, 00h
    LD D, 00h
VDP_Line_VLoop:
    LD A, (IX+4)  ; Cor
    CALL VDP_PSet
    LD A, (VDP_Line_H_End)
    CP E
    JP Z, VDP_Line_Exit
    INC E
    JR VDP_Line_VLoop

VDP_Line_General:
    ; Bresenham completo
    ; DX e SX
    LD A, (IX+8)  ; X2
    SUB (IX+12)   ; X2 - X1
    JR NC, VDP_Line_SX_Pos
    NEG
    LD (VDP_Line_DX), A
    LD A, 0FFh   ; SX = -1
    LD (VDP_Line_SX), A
    JR VDP_Line_CalcDY
VDP_Line_SX_Pos:
    LD (VDP_Line_DX), A
    LD A, 01h    ; SX = +1
    LD (VDP_Line_SX), A

VDP_Line_CalcDY:
    ; DY e SY: DY será armazenado como valor negativo (-abs(Y2 - Y1)) em 16-bit
    LD A, (IX+6)  ; Y2
    SUB (IX+10)   ; Y2 - Y1
    JR NC, VDP_Line_SY_Pos
    NEG           ; A = abs(Y2 - Y1)
    LD L, A
    LD H, 00h
    XOR A
    SUB L
    LD L, A
    SBC A, A
    LD H, A
    LD (VDP_Line_DY), HL
    LD A, 0FFh   ; SY = -1
    LD (VDP_Line_SY), A
    JR VDP_Line_CalcErr
VDP_Line_SY_Pos:
    LD L, A
    LD H, 00h
    XOR A
    SUB L
    LD L, A
    SBC A, A
    LD H, A
    LD (VDP_Line_DY), HL
    LD A, 01h    ; SY = +1
    LD (VDP_Line_SY), A

VDP_Line_CalcErr:
    ; Err = DX + DY
    LD HL, (VDP_Line_DY)
    EX DE, HL
    LD A, (VDP_Line_DX)
    LD L, A
    LD H, 00h
    ADD HL, DE
    LD (VDP_Line_Err), HL

    ; Ponto inicial (X, Y)
    LD A, (IX+12)
    LD (VDP_Line_CurX), A
    LD A, (IX+10)
    LD (VDP_Line_CurY), A

VDP_Line_Loop:
    ; Plota ponto atual
    LD A, (VDP_Line_CurX)
    LD C, A
    LD B, 00h
    LD A, (VDP_Line_CurY)
    LD E, A
    LD D, 00h
    LD A, (IX+4)
    CALL VDP_PSet

    ; Verifica se chegou ao fim: CurX == X2 && CurY == Y2
    LD A, (VDP_Line_CurX)
    CP (IX+8)
    JR NZ, VDP_Line_Step
    LD A, (VDP_Line_CurY)
    CP (IX+6)
    JP Z, VDP_Line_Exit

VDP_Line_Step:
    ; E2 = 2 * Err
    LD HL, (VDP_Line_Err)
    ADD HL, HL
    LD (VDP_Line_E2), HL

    ; Se E2 >= DY: Err += DY, CurX += SX
    ; HL = E2 - DY
    LD HL, (VDP_Line_DY)
    EX DE, HL
    LD HL, (VDP_Line_E2)
    OR A
    SBC HL, DE          ; HL = E2 - DY
    JP M, VDP_Line_CheckE2DX
    LD HL, (VDP_Line_Err)
    ADD HL, DE
    LD (VDP_Line_Err), HL
    LD A, (VDP_Line_SX)
    LD B, A
    LD A, (VDP_Line_CurX)
    ADD A, B
    LD (VDP_Line_CurX), A

VDP_Line_CheckE2DX:
    ; Se E2 <= DX: Err += DX, CurY += SY
    LD A, (VDP_Line_DX)
    LD E, A
    LD D, 00h
    LD HL, (VDP_Line_E2)
    EX DE, HL
    OR A
    SBC HL, DE          ; HL = DX - E2
    JP M, VDP_Line_Loop
    LD HL, (VDP_Line_Err)
    LD A, (VDP_Line_DX)
    LD E, A
    LD D, 00h
    ADD HL, DE
    LD (VDP_Line_Err), HL
    LD A, (VDP_Line_SY)
    LD B, A
    LD A, (VDP_Line_CurY)
    ADD A, B
    LD (VDP_Line_CurY), A
    JR VDP_Line_Loop

VDP_Line_Exit:
    POP HL
    POP DE
    POP BC
    POP IX
    RET

VDP_Line_H_End: DB 00h
VDP_Line_DX:    DB 00h
VDP_Line_SX:    DB 00h
VDP_Line_DY:    DW 0000h
VDP_Line_SY:    DB 00h
VDP_Line_Err:   DW 0000h
VDP_Line_E2:    DW 0000h
VDP_Line_CurX:  DB 00h
VDP_Line_CurY:  DB 00h

ENDMOD

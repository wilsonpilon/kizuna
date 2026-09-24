; =============================================================================
; KIZUNA MSXLIB - vdp/blinktable
; tabela de piscar do texto de 80 colunas: 1 bit por celula, 10 bytes por linha
; =============================================================================

MODULE vdp_blinktable
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_BlinkFill, VDP_BlinkLine, VDP_BlinkCell
EXTERN VDP_GetColorTable, VDP_VramSetWrite, VDP_VramSetRead, VDP_VramPut, VDP_VramGet, VDP_VramFillStream

; A tabela de piscar e a "tabela de cores" (VDP_SetColorTable) do texto de 80 colunas:
; cada linha de texto ocupa 10 bytes (80 celulas), o bit 7 de cada byte e a celula
; mais a esquerda do grupo de 8. Cabem 27 linhas (270 bytes).

; VDP_BlinkFill: preenche a tabela inteira (0 = nada pisca, 0FFh = tudo pisca)
; Entrada: A = valor
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_BlinkFill:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD E, A
    CALL VDP_GetColorTable
    CALL VDP_VramSetWrite
    LD A, E
    LD BC, 010Eh        ; 270 bytes
    CALL VDP_VramFillStream
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_BlinkLine: preenche os 10 bytes de uma linha de texto
; Entrada: A = linha (0..26), B = valor (0 = nada pisca, 0FFh = a linha toda pisca)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_BlinkLine:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD E, B
    LD L, A
    LD H, 00h
    ADD HL, HL          ; 2 * linha
    PUSH HL
    ADD HL, HL
    ADD HL, HL          ; 8 * linha
    POP BC
    ADD HL, BC          ; 10 * linha
    PUSH HL
    CALL VDP_GetColorTable
    POP BC
    ADD HL, BC
    ADC A, 00h
    AND 01h
    CALL VDP_VramSetWrite
    LD A, E
    LD BC, 000Ah
    CALL VDP_VramFillStream
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_BlinkCell: liga ou desliga o piscar de uma celula
; Entrada: A = coluna (0..79), B = linha (0..26), C = 0 desliga, outro valor liga
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_BlinkCell:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD D, A
    LD A, C
    LD (VDP_BlinkCell_On), A
    LD A, D
    AND 07h
    LD E, A             ; E = coluna & 7
    LD A, D
    RRCA
    RRCA
    RRCA
    AND 1Fh
    LD D, A             ; D = coluna / 8
    LD H, 00h
    LD L, B
    ADD HL, HL
    PUSH HL
    ADD HL, HL
    ADD HL, HL
    POP BC
    ADD HL, BC          ; HL = 10 * linha
    LD C, D
    LD B, 00h
    ADD HL, BC          ; HL = deslocamento do byte
    PUSH HL
    CALL VDP_GetColorTable
    POP BC
    ADD HL, BC
    ADC A, 00h
    AND 01h
    PUSH AF
    PUSH HL
    CALL VDP_VramSetRead
    CALL VDP_VramGet
    LD B, A             ; B = byte atual
    LD A, 80h
    INC E
    DEC E
    JR Z, VDP_BlinkCell_Mask
VDP_BlinkCell_Shift:
    SRL A
    DEC E
    JR NZ, VDP_BlinkCell_Shift
VDP_BlinkCell_Mask:
    LD C, A             ; C = mascara do bit
    LD A, (VDP_BlinkCell_On)
    OR A
    LD A, C
    JR Z, VDP_BlinkCell_Clear
    OR B
    JR VDP_BlinkCell_Write
VDP_BlinkCell_Clear:
    CPL
    AND B
VDP_BlinkCell_Write:
    LD B, A
    POP HL
    POP AF
    CALL VDP_VramSetWrite
    LD A, B
    CALL VDP_VramPut
    POP HL
    POP DE
    POP BC
    POP AF
    RET
VDP_BlinkCell_On:
    DB 00h

ENDMOD

; =============================================================================
; Código gerado pelo compilador MSX-BASIC Dignified DIGNAC - Kizuna Toolchain
; Módulo: Chart | Banco: 0
; =============================================================================

MODULE Chart
BANK 0

PUBLIC Main, Desenhar, Start
EXTERN BIOS_CHGET, BDOS_Exit, Mul16, Div16, VDP_PSet, VDP_Line, VDP_BoxFill, BIOS_CHGMOD

; --- Ponto de Entrada para Executável MSX-DOS 2 ---
Start:
    CALL Main
    CALL BDOS_Exit

; -----------------------------------------------------------------------------
; Procedimento: Main (Locais: 0 bytes, Parâmetros: 0)
; -----------------------------------------------------------------------------
Main:
    PUSH IX
    LD IX, 0000h
    ADD IX, SP
    LD HL, 0002h
    LD A, L
    CALL BIOS_CHGMOD
    LD HL, 000Ah
    PUSH HL
    CALL Desenhar
    LD HL, 2
    ADD HL, SP
    LD SP, HL
    CALL BIOS_CHGET
    LD HL, 0000h
    LD A, L
    CALL BIOS_CHGMOD
.Exit_Main:
    LD SP, IX
    POP IX
    RET

; -----------------------------------------------------------------------------
; Procedimento: Desenhar (Locais: 4 bytes, Parâmetros: 1)
; -----------------------------------------------------------------------------
Desenhar:
    PUSH IX
    LD IX, 0000h
    ADD IX, SP
    LD HL, -4
    ADD HL, SP
    LD SP, HL
    LD HL, 0000h
    PUSH HL
    LD HL, 0000h
    PUSH HL
    LD HL, 00FFh
    PUSH HL
    LD HL, 00BFh
    PUSH HL
    LD HL, 0001h
    PUSH HL
    CALL VDP_BoxFill
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD HL, 0008h
    PUSH HL
    LD HL, 0008h
    PUSH HL
    LD HL, 00F7h
    PUSH HL
    LD HL, 0008h
    PUSH HL
    LD HL, 000Fh
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD HL, 00F7h
    PUSH HL
    LD HL, 0008h
    PUSH HL
    LD HL, 00F7h
    PUSH HL
    LD HL, 00B7h
    PUSH HL
    LD HL, 000Fh
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD HL, 00F7h
    PUSH HL
    LD HL, 00B7h
    PUSH HL
    LD HL, 0008h
    PUSH HL
    LD HL, 00B7h
    PUSH HL
    LD HL, 000Fh
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD HL, 0008h
    PUSH HL
    LD HL, 00B7h
    PUSH HL
    LD HL, 0008h
    PUSH HL
    LD HL, 0008h
    PUSH HL
    LD HL, 000Fh
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD HL, 0018h
    PUSH HL
    LD HL, 0014h
    PUSH HL
    LD HL, 0018h
    PUSH HL
    LD HL, 00A5h
    PUSH HL
    LD HL, 0007h
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD HL, 0018h
    PUSH HL
    LD HL, 00A5h
    PUSH HL
    LD HL, 00ECh
    PUSH HL
    LD HL, 00A5h
    PUSH HL
    LD HL, 0007h
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD HL, 0018h
    PUSH HL
    LD HL, 0082h
    PUSH HL
    LD HL, 00ECh
    PUSH HL
    LD HL, 0082h
    PUSH HL
    LD HL, 000Eh
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD HL, 0018h
    PUSH HL
    LD HL, 005Fh
    PUSH HL
    LD HL, 00ECh
    PUSH HL
    LD HL, 005Fh
    PUSH HL
    LD HL, 000Eh
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD HL, 0018h
    PUSH HL
    LD HL, 003Ch
    PUSH HL
    LD HL, 00ECh
    PUSH HL
    LD HL, 003Ch
    PUSH HL
    LD HL, 000Eh
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD HL, 0020h
    LD (IX - 2), L
    LD (IX - 1), H
For_Start_1:
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD HL, 00E4h
    EX DE, HL
    POP HL
    OR A
    SBC HL, DE
    JP Z, For_Ok_3
    JP NC, For_End_2
For_Ok_3:
    LD HL, 00A0h
    PUSH HL
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD L, (IX + 4)
    LD H, (IX + 5)
    PUSH HL
    LD HL, 0001h
    EX DE, HL
    POP HL
    ADD HL, DE
    EX DE, HL
    POP HL
    CALL Div16
    EX DE, HL
    PUSH HL
    LD HL, 0009h
    EX DE, HL
    POP HL
    CALL Mul16
    EX DE, HL
    POP HL
    OR A
    SBC HL, DE
    LD (IX - 4), L
    LD (IX - 3), H
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD HL, 00A4h
    PUSH HL
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD L, (IX - 4)
    LD H, (IX - 3)
    PUSH HL
    LD HL, 000Ah
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD HL, 0001h
    EX DE, HL
    POP HL
    ADD HL, DE
    PUSH HL
    LD HL, 00A4h
    PUSH HL
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD HL, 0001h
    EX DE, HL
    POP HL
    ADD HL, DE
    PUSH HL
    LD L, (IX - 4)
    LD H, (IX - 3)
    PUSH HL
    LD HL, 000Ah
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD HL, 0002h
    EX DE, HL
    POP HL
    ADD HL, DE
    PUSH HL
    LD HL, 00A4h
    PUSH HL
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD HL, 0002h
    EX DE, HL
    POP HL
    ADD HL, DE
    PUSH HL
    LD L, (IX - 4)
    LD H, (IX - 3)
    PUSH HL
    LD HL, 000Ah
    PUSH HL
    CALL VDP_Line
    LD HL, 10
    ADD HL, SP
    LD SP, HL
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD L, (IX - 4)
    LD H, (IX - 3)
    PUSH HL
    LD HL, 000Fh
    LD A, L
    POP DE
    POP BC
    CALL VDP_PSet
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD HL, 0001h
    EX DE, HL
    POP HL
    ADD HL, DE
    PUSH HL
    LD L, (IX - 4)
    LD H, (IX - 3)
    PUSH HL
    LD HL, 000Fh
    LD A, L
    POP DE
    POP BC
    CALL VDP_PSet
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD HL, 0002h
    EX DE, HL
    POP HL
    ADD HL, DE
    PUSH HL
    LD L, (IX - 4)
    LD H, (IX - 3)
    PUSH HL
    LD HL, 000Fh
    LD A, L
    POP DE
    POP BC
    CALL VDP_PSet
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD HL, 0008h
    EX DE, HL
    POP HL
    ADD HL, DE
    LD (IX - 2), L
    LD (IX - 1), H
    JP For_Start_1
For_End_2:
    LD HL, 0019h
    LD (IX - 2), L
    LD (IX - 1), H
For_Start_4:
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD HL, 00EBh
    EX DE, HL
    POP HL
    OR A
    SBC HL, DE
    JP Z, For_Ok_6
    JP NC, For_End_5
For_Ok_6:
    LD HL, 00A0h
    PUSH HL
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD L, (IX + 4)
    LD H, (IX + 5)
    PUSH HL
    LD HL, 0001h
    EX DE, HL
    POP HL
    ADD HL, DE
    EX DE, HL
    POP HL
    CALL Div16
    EX DE, HL
    PUSH HL
    LD HL, 0009h
    EX DE, HL
    POP HL
    CALL Mul16
    EX DE, HL
    POP HL
    OR A
    SBC HL, DE
    LD (IX - 4), L
    LD (IX - 3), H
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD L, (IX - 4)
    LD H, (IX - 3)
    PUSH HL
    LD HL, 000Fh
    LD A, L
    POP DE
    POP BC
    CALL VDP_PSet
    LD L, (IX - 2)
    LD H, (IX - 1)
    INC HL
    LD (IX - 2), L
    LD (IX - 1), H
    JP For_Start_4
For_End_5:
.Exit_Desenhar:
    LD SP, IX
    POP IX
    RET

ENDMOD

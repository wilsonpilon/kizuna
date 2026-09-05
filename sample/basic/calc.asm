; =============================================================================
; Código gerado pelo compilador MSX-BASIC Dignified DIGNAC - Kizuna Toolchain
; Módulo: Calc | Banco: 0
; =============================================================================

MODULE Calc
BANK 0

PUBLIC Main, Start
EXTERN BDOS_Exit, Mul16, Div16, BDOS_PrintString, BDOS_PrintChar, PrintDec16

; --- Ponto de Entrada para Executável MSX-DOS 2 ---
Start:
    CALL Main
    CALL BDOS_Exit

; -----------------------------------------------------------------------------
; Procedimento: Main (Locais: 6 bytes, Parâmetros: 0)
; -----------------------------------------------------------------------------
Main:
    PUSH IX
    LD IX, 0000h
    ADD IX, SP
    LD HL, -6
    ADD HL, SP
    LD SP, HL
    LD HL, 007Bh
    LD (IX - 2), L
    LD (IX - 1), H
    LD HL, 002Dh
    LD (IX - 4), L
    LD (IX - 3), H
    LD L, (IX - 2)
    LD H, (IX - 1)
    PUSH HL
    LD L, (IX - 4)
    LD H, (IX - 3)
    EX DE, HL
    POP HL
    CALL Mul16
    LD (IX - 6), L
    LD (IX - 5), H
    LD DE, StrLit_1
    CALL BDOS_PrintString
    LD A, 0Dh
    CALL BDOS_PrintChar
    LD A, 0Ah
    CALL BDOS_PrintChar
    LD DE, StrLit_2
    CALL BDOS_PrintString
    LD A, 0Dh
    CALL BDOS_PrintChar
    LD A, 0Ah
    CALL BDOS_PrintChar
    LD DE, StrLit_1
    CALL BDOS_PrintString
    LD A, 0Dh
    CALL BDOS_PrintChar
    LD A, 0Ah
    CALL BDOS_PrintChar
    LD DE, StrLit_3
    CALL BDOS_PrintString
    LD L, (IX - 2)
    LD H, (IX - 1)
    CALL PrintDec16
    LD A, 0Dh
    CALL BDOS_PrintChar
    LD A, 0Ah
    CALL BDOS_PrintChar
    LD DE, StrLit_4
    CALL BDOS_PrintString
    LD L, (IX - 4)
    LD H, (IX - 3)
    CALL PrintDec16
    LD A, 0Dh
    CALL BDOS_PrintChar
    LD A, 0Ah
    CALL BDOS_PrintChar
    LD DE, StrLit_5
    CALL BDOS_PrintString
    LD L, (IX - 6)
    LD H, (IX - 5)
    CALL PrintDec16
    LD A, 0Dh
    CALL BDOS_PrintChar
    LD A, 0Ah
    CALL BDOS_PrintChar
    LD DE, StrLit_6
    CALL BDOS_PrintString
    LD L, (IX - 6)
    LD H, (IX - 5)
    PUSH HL
    LD HL, 000Ah
    EX DE, HL
    POP HL
    CALL Div16
    CALL PrintDec16
    LD A, 0Dh
    CALL BDOS_PrintChar
    LD A, 0Ah
    CALL BDOS_PrintChar
    LD DE, StrLit_7
    CALL BDOS_PrintString
    LD L, (IX - 6)
    LD H, (IX - 5)
    PUSH HL
    LD HL, 000Ah
    EX DE, HL
    POP HL
    CALL Div16
    EX DE, HL
    CALL PrintDec16
    LD A, 0Dh
    CALL BDOS_PrintChar
    LD A, 0Ah
    CALL BDOS_PrintChar
.Exit_Main:
    LD SP, IX
    POP IX
    RET

; --- Literais de String ---
StrLit_2:
    DB " DIGNAC Math & Logic Demo (16-bit)      $"
StrLit_3:
    DB "a% = $"
StrLit_4:
    DB "b% = $"
StrLit_5:
    DB "Multiplicacao (a% * b%): $"
StrLit_6:
    DB "Divisao inteira (res% / 10): $"
StrLit_7:
    DB "Resto / Modulo (res% MOD 10): $"
StrLit_1:
    DB "========================================$"

ENDMOD

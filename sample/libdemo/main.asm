; =============================================================================
; KIZUNA - Exemplo de Uso da Biblioteca Padrao MSXLIB com Smart-Linking
; =============================================================================
; Este programa demonstra a utilizacao da msxlib.hlib.
; Rotinas chamadas:
;   - BDOS_PrintString
;   - Mul16
;   - PrintDec16
;   - PSG_PlayTone / PSG_MuteAll
;   - BDOS_Exit
;
; O Linker MUSUBI extraira apenas os modulos necessarios (bdos, math, string, psg)
; e eliminara os modulos nao utilizados (bios, vdp) automaticamente.
; =============================================================================

MODULE Main
BANK 0

PUBLIC Start

EXTERN BDOS_PrintString, BDOS_PrintChar, BDOS_Exit
EXTERN Mul16, PrintDec16
EXTERN PSG_PlayTone, PSG_MuteAll

Start:
    ; 1. Imprimir mensagem de boas-vindas
    LD DE, MsgIntro
    CALL BDOS_PrintString

    ; 2. Demonstrar rotina aritmetica Mul16: 123 * 45 = 5535
    LD HL, 007Bh ; 123
    LD DE, 002Dh ; 45
    CALL Mul16   ; HL = 5535

    ; 3. Imprimir o resultado em decimal via PrintDec16
    CALL PrintDec16

    ; 4. Imprimir nova linha (CR, LF)
    LD E, 0Dh
    CALL BDOS_PrintChar
    LD E, 0Ah
    CALL BDOS_PrintChar

    ; 5. Tocar uma nota musical no PSG (Canal 0, Tom A440, Volume 15)
    LD A, 00h    ; Canal A
    LD HL, 00FEh ; Periodo aprox. 440 Hz no clock MSX
    LD E, 0Fh    ; Volume maximo (15)
    CALL PSG_PlayTone

    ; Pausa audivel de aproximadamente um segundo
    LD B, 02h
DelayOuter:
    LD HL, 0000h
DelayInner:
    DEC HL
    LD A, H
    OR L
    JR NZ, DelayInner
    DJNZ DelayOuter

    ; 6. Silenciar som
    CALL PSG_MuteAll

    ; 7. Mensagem final e retorno ao MSX-DOS
    LD DE, MsgDone
    CALL BDOS_PrintString
    CALL BDOS_Exit

MsgIntro:
    DB "Kizuna MSXLIB Demo - Calculando 123 x 45 = $"

MsgDone:
    DB "Som tocado com sucesso! Retornando ao DOS.", 0Dh, 0Ah, "$"

ENDMOD

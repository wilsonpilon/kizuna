; ==============================================================================
; KIZUNA sample -- Música PSG (PSG_PlaySequence/PSG_PlayNoteIndexed)
; Compilador: KAJI80
; Toca uma escala simples de Dó a Dó e de volta, usando a tabela de notas da
; MSXLIB em vez de períodos calculados à mão.
; ==============================================================================

MODULE MUSICDEMO
BANK 0

PUBLIC Start
EXTERN BIOS_CHGET
EXTERN PSG_PlaySequence, PSG_MuteAll

BDOS    EQU 0005h
C_WRITE EQU 09h

Start:
    LD DE, MsgIntro
    LD C, C_WRITE
    CALL BDOS

    LD HL, Tune
    CALL PSG_PlaySequence

    CALL PSG_MuteAll ; a ultima nota tocaria indefinidamente sem isso

    LD DE, MsgDone
    LD C, C_WRITE
    CALL BDOS

    CALL BIOS_CHGET

    RET

MsgIntro:
    DB 0Dh, 0Ah
    DB "KIZUNA sample -- Musica PSG (PSG_PlaySequence)", 0Dh, 0Ah
    DB "Tocando escala de Do a Do...", 0Dh, 0Ah
    DB "$"

MsgDone:
    DB 0Dh, 0Ah
    DB "Fim da sequencia. Pressione uma tecla para sair.", 0Dh, 0Ah
    DB "$"

; Sequência: canal, índice-de-nota (PSG_NoteTable), volume, duração;
; terminada por canal = 0FFh. Escala C4..C5 subindo, depois descendo de
; volta a C4.
Tune:
    DB 0, 24, 0Fh, 50 ; C4
    DB 0, 26, 0Fh, 50 ; D4
    DB 0, 28, 0Fh, 50 ; E4
    DB 0, 29, 0Fh, 50 ; F4
    DB 0, 31, 0Fh, 50 ; G4
    DB 0, 33, 0Fh, 50 ; A4
    DB 0, 35, 0Fh, 50 ; B4
    DB 0, 36, 0Fh, 60 ; C5
    DB 0, 35, 0Fh, 50 ; B4
    DB 0, 33, 0Fh, 50 ; A4
    DB 0, 31, 0Fh, 50 ; G4
    DB 0, 29, 0Fh, 50 ; F4
    DB 0, 28, 0Fh, 50 ; E4
    DB 0, 26, 0Fh, 50 ; D4
    DB 0, 24, 0Fh, 80 ; C4
    DB 0FFh

ENDMOD

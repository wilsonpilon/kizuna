; =============================================================================
; KIZUNA MSXLIB - BIOS.ASM
; Rotinas de suporte para chamadas ao sistema preservando o estado do MSX-DOS
; =============================================================================

MODULE BIOS
BANK 0

PUBLIC BIOS_Call, BIOS_CHPUT, BIOS_CHGET, BIOS_CLS, BIOS_POSIT, BIOS_BEEP, BIOS_INIT32, BIOS_CHGMOD
EXTERN VDP_SetScreen, VDP_InitScreen2_Tables

EXPTBL EQU 0FCC1h
CALSLT EQU 001Ch
BDOS_ENTRY EQU 0005h

; -----------------------------------------------------------------------------
; BIOS_Call: Executa inter-slot call para a Main-ROM BIOS (Slot primário lido de EXPTBL)
; Entrada: IX = endereço da rotina na Main-ROM
; -----------------------------------------------------------------------------
BIOS_Call:
    PUSH AF
    PUSH HL
    LD A, (EXPTBL)
    LD H, A
    LD L, 00h
    PUSH HL
    POP IY
    POP HL
    POP AF
    CALL CALSLT
    EI
    RET

; -----------------------------------------------------------------------------
; BIOS_CHPUT: Imprime caractere no console via BDOS (Função 02h)
; Entrada: A = código ASCII do caractere
; -----------------------------------------------------------------------------
BIOS_CHPUT:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD E, A
    LD C, 02h
    CALL BDOS_ENTRY
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; -----------------------------------------------------------------------------
; BIOS_CHGET: Aguarda uma tecla com temporizador de ~60 segundos (1 minuto)
; Esvazia buffer prévio (Enter do comando) e aguarda tecla ou 1 minuto
; Saída: A = código da tecla (ou 0 se deu timeout)
; -----------------------------------------------------------------------------
BIOS_CHGET:
    PUSH BC
    PUSH DE
    PUSH HL

    ; 1. Drena qualquer caractere residual (ex: Enter do prompt)
BIOS_CHGET_Flush:
    LD C, 06h
    LD E, 0FFh
    CALL BDOS_ENTRY
    OR A
    JR NZ, BIOS_CHGET_Flush

    ; 2. Loop de temporização (~60 segundos no Z80 a 3.58 MHz)
    ; A cada passo checa se o usuário pressionou alguma tecla
    LD HL, 5000         ; ~60 segundos totais
BIOS_CHGET_Outer:
    LD B, 00h           ; 256 iterações internas
BIOS_CHGET_Inner:
    LD C, 06h
    LD E, 0FFh
    CALL BDOS_ENTRY
    OR A
    JR Z, BIOS_CHGET_NoKey
    CP 1Bh               ; Verifica se é tecla ESC (ASCII 27 / 1Bh)
    JR Z, BIOS_CHGET_Key ; Se for ESC, sai imediatamente!
    ; Qualquer outra tecla é ignorada e a contagem continua

BIOS_CHGET_NoKey:
    DJNZ BIOS_CHGET_Inner

    DEC HL
    LD A, H
    OR L
    JR NZ, BIOS_CHGET_Outer

    ; Timeout de 60 segundos expirado
    XOR A
    JR BIOS_CHGET_Exit

BIOS_CHGET_Key:
    ; Usuário pressionou tecla

BIOS_CHGET_Exit:
    POP HL
    POP DE
    POP BC
    RET

; -----------------------------------------------------------------------------
; BIOS_CLS: Limpa a tela no modo atual
; -----------------------------------------------------------------------------
BIOS_CLS:
    PUSH AF
    XOR A
    CALL BIOS_CHGMOD
    POP AF
    RET

; -----------------------------------------------------------------------------
; BIOS_POSIT: Posiciona o cursor de texto (reservado)
; Entrada: H = coluna (1..X), L = linha (1..Y)
; -----------------------------------------------------------------------------
BIOS_POSIT:
    RET

; -----------------------------------------------------------------------------
; BIOS_BEEP: Emite som de aviso padrão do MSX (BEL ASCII 07h via BDOS)
; -----------------------------------------------------------------------------
BIOS_BEEP:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD E, 07h
    LD C, 02h
    CALL BDOS_ENTRY
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; -----------------------------------------------------------------------------
; BIOS_INIT32: Inicializa modo SCREEN 1 (32 colunas)
; -----------------------------------------------------------------------------
BIOS_INIT32:
    PUSH AF
    LD A, 01h
    CALL BIOS_CHGMOD
    POP AF
    RET

; -----------------------------------------------------------------------------
; BIOS_CHGMOD: Altera o modo de tela chamando rotinas oficiais da BIOS
; Entrada: A = modo de vídeo (0 = SCREEN 0, 1 = SCREEN 1, 2 = SCREEN 2)
; Compatível com MSX1, MSX2, MSX2+ e MSX Turbo R
; -----------------------------------------------------------------------------
BIOS_CHGMOD:
    PUSH BC
    PUSH DE
    PUSH HL
    PUSH IX
    PUSH IY
    PUSH AF

    OR A
    JR NZ, BIOS_CHGMOD_Not0

    ; Modo 0 (SCREEN 0): Chama INITXT (006Ch) para inicializar tela texto e restaurar fonte ROM
    LD IX, 006Ch
    CALL BIOS_Call

    ; Restaura cores padrão do texto (Branco sobre Preto)
    LD A, 0Fh
    LD (0F3E9h), A      ; FORGND = 15 (Branco)
    LD A, 01h
    LD (0F3EAh), A      ; BAKGDN = 1 (Preto)
    LD (0F3EBh), A      ; BDRCLR = 1 (Preto)
    LD IX, 0062h        ; CHGCLR
    CALL BIOS_Call
    JR BIOS_CHGMOD_Done

BIOS_CHGMOD_Not0:
    CP 02h
    JR NZ, BIOS_CHGMOD_Other

    ; Modo 2 (SCREEN 2): Chama INIGRP (0072h) da BIOS para configurar hardware e tabelas
    LD IX, 0072h
    CALL BIOS_Call

    ; Inicializa tabelas VRAM essenciais (Name Table 3x 0..255, Pattern, Color)
    CALL VDP_InitScreen2_Tables
    JR BIOS_CHGMOD_Done

BIOS_CHGMOD_Other:
    ; Demais modos de vídeo: CHGMOD (005Fh) padrão
    LD IX, 005Fh
    CALL BIOS_Call

BIOS_CHGMOD_Done:
    POP AF
    POP IY
    POP IX
    POP HL
    POP DE
    POP BC
    RET

ENDMOD

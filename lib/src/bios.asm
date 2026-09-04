; =============================================================================
; KIZUNA MSXLIB - BIOS.ASM
; Rotinas de suporte para chamadas à Main-ROM BIOS via Inter-Slot Call (CALSLT)
; =============================================================================

MODULE BIOS
BANK 0

PUBLIC BIOS_Call, BIOS_CHPUT, BIOS_CHGET, BIOS_CLS, BIOS_POSIT, BIOS_BEEP, BIOS_INIT32

EXPTBL EQU 0FCC1h
CALSLT EQU 0024h

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
; BIOS_CHPUT: Imprime caractere na tela via BIOS (0x00A2)
; Entrada: A = código ASCII do caractere
; -----------------------------------------------------------------------------
BIOS_CHPUT:
    PUSH IX
    PUSH IY
    LD IX, 00A2h
    CALL BIOS_Call
    POP IY
    POP IX
    RET

; -----------------------------------------------------------------------------
; BIOS_CHGET: Lê caractere do teclado via BIOS (0x009F)
; Saída: A = código da tecla pressionada
; -----------------------------------------------------------------------------
BIOS_CHGET:
    PUSH IX
    PUSH IY
    LD IX, 009Fh
    CALL BIOS_Call
    POP IY
    POP IX
    RET

; -----------------------------------------------------------------------------
; BIOS_CLS: Limpa a tela no modo de texto atual (0x00C3)
; -----------------------------------------------------------------------------
BIOS_CLS:
    PUSH IX
    PUSH IY
    PUSH AF
    LD IX, 00C3h
    CALL BIOS_Call
    POP AF
    POP IY
    POP IX
    RET

; -----------------------------------------------------------------------------
; BIOS_POSIT: Posiciona o cursor de texto (0x00C6)
; Entrada: H = coluna (1..X), L = linha (1..Y)
; -----------------------------------------------------------------------------
BIOS_POSIT:
    PUSH IX
    PUSH IY
    PUSH AF
    LD IX, 00C6h
    CALL BIOS_Call
    POP AF
    POP IY
    POP IX
    RET

; -----------------------------------------------------------------------------
; BIOS_BEEP: Emite som de aviso padrão do MSX (0x00C0)
; -----------------------------------------------------------------------------
BIOS_BEEP:
    PUSH IX
    PUSH IY
    PUSH AF
    LD IX, 00C0h
    CALL BIOS_Call
    POP AF
    POP IY
    POP IX
    RET

; -----------------------------------------------------------------------------
; BIOS_INIT32: Inicializa modo SCREEN 1 (32 colunas) (0x006F)
; -----------------------------------------------------------------------------
BIOS_INIT32:
    PUSH IX
    PUSH IY
    PUSH AF
    LD IX, 006Fh
    CALL BIOS_Call
    POP AF
    POP IY
    POP IX
    RET

ENDMOD

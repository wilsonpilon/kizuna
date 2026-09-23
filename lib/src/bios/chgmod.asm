; =============================================================================
; KIZUNA MSXLIB - bios/chgmod
; altera o modo de tela
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bios_chgmod
BANK 0

PUBLIC BIOS_CHGMOD
EXTERN BIOS_Call, VDP_InitScreen2_Tables

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

    ; Modo 2 (SCREEN 2): CHGMOD (005Fh) seleciona o modo indicado em A.
    ; INIGRP (0072h) inicializa apenas o modo gráfico padrão (SCREEN 1).
    LD IX, 005Fh
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

; =============================================================================
; KIZUNA MSXLIB - bios/chget
; aguarda tecla
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bios_chget
BANK 0

PUBLIC BIOS_CHGET
INCLUDE "../../inc/bdos.inc"

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
    JR NZ, BIOS_CHGET_Key ; Qualquer tecla pressionada sai imediatamente
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

ENDMOD

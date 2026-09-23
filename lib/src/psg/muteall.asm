; =============================================================================
; KIZUNA MSXLIB - psg/muteall
; silencia os 3 canais
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE psg_muteall
BANK 0

PUBLIC PSG_MuteAll
EXTERN PSG_Write

; -----------------------------------------------------------------------------
; PSG_MuteAll: Zera o volume dos 3 canais de som e silencia o PSG
; -----------------------------------------------------------------------------
PSG_MuteAll:
    PUSH AF
    PUSH DE
    LD E, 00h
    LD A, 08h ; Canal A volume
    CALL PSG_Write
    LD A, 09h ; Canal B volume
    CALL PSG_Write
    LD A, 0Ah ; Canal C volume
    CALL PSG_Write
    ; Desativa tons no mixer de forma segura para portas de I/O do MSX (0xBF)
    LD A, 07h
    LD E, 0BFh
    CALL PSG_Write
    POP DE
    POP AF
    RET

ENDMOD

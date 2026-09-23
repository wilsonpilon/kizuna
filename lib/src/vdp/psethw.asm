; =============================================================================
; KIZUNA MSXLIB - vdp/psethw
; PSET pelo motor de comando do V9938/V9958 (modos bitmap)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_psethw
BANK 0

PUBLIC VDP_PSet_HW
EXTERN VDP_CommandWait_Raw, VDP_WriteReg_Raw

; -----------------------------------------------------------------------------
; VDP_PSet_HW: motor de comando de hardware do V9938/V9958 (comando PSET,
; R#36-46) -- sem DI/EI proprios, deve ser chamada apenas com interrupcoes
; ja desabilitadas pelo chamador. Entrada/Saida: identicas ao VDP_PSet.
;
; ATENCAO: o motor de comando do V9938/V9958 SO funciona nos modos bitmap
; (Graphic 4-7 / SCREEN 5-8) -- confirmado via documentacao tecnica externa
; em 2026-09-10. Em SCREEN 2 (Graphic 2, modo baseado em Pattern/Name/Color
; Table como este), os comandos sao aceitos pelos registradores mas NAO tem
; efeito visivel algum. NAO USAR esta rotina para SCREEN 2 -- mantida aqui
; apenas para quando a MSXLIB ganhar suporte a SCREEN 5+ (Graphic 4-7).
; -----------------------------------------------------------------------------
VDP_PSet_HW:
    PUSH BC
    PUSH DE
    PUSH AF
    CALL VDP_CommandWait_Raw  ; garante que um comando anterior ja terminou

    LD B, C
    LD C, 24h                 ; R#36 = X (low)
    CALL VDP_WriteReg_Raw
    LD B, 00h
    LD C, 25h                 ; R#37 = X (high, sempre 0: X < 256)
    CALL VDP_WriteReg_Raw
    LD B, E
    LD C, 26h                 ; R#38 = Y (low)
    CALL VDP_WriteReg_Raw
    LD B, D
    LD C, 27h                 ; R#39 = Y (high, sempre 0: Y < 256)
    CALL VDP_WriteReg_Raw

    POP AF                    ; recupera a cor original
    PUSH AF
    LD B, A
    LD C, 2Ch                 ; R#44 = Cor
    CALL VDP_WriteReg_Raw
    LD B, 00h
    LD C, 2Dh                 ; R#45 = Argumento (0 = operacao normal)
    CALL VDP_WriteReg_Raw
    LD B, 50h                 ; 50h = comando PSET, operacao logica 0 (copia)
    LD C, 2Eh                 ; R#46 = Comando -- escrever aqui dispara
    CALL VDP_WriteReg_Raw

    CALL VDP_CommandWait_Raw  ; aguarda ESTE PSET terminar antes de retornar

    POP AF
    POP DE
    POP BC
    RET

ENDMOD

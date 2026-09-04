; ==============================================================================
; KIZUNA (絆) - Exemplo Multi-Banco: Módulo no Banco 1 (Paginável na Página 2)
; ==============================================================================

MODULE BANK1
BANK 1

PUBLIC PrintBank1

BDOS       EQU 0x0005
C_WRITE    EQU 0x09

PrintBank1:
    ; Imprime mensagem oficial do Banco 1 na Página 2
    ld   de, MsgBank1
    ld   c, C_WRITE
    call BDOS
    ret

MsgBank1:
    db 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db ">>> [BANCO 1] ROTINA DO BANCO 1 EXECUTADA!", 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "$"

ENDMOD

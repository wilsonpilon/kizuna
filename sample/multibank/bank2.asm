; ==============================================================================
; KIZUNA (絆) - Exemplo Multi-Banco: Módulo no Banco 2 (Paginável na Página 2)
; ==============================================================================

MODULE BANK2
BANK 2

PUBLIC PrintBank2

BDOS       EQU 0x0005
C_WRITE    EQU 0x09

PrintBank2:
    ; Imprime mensagem oficial do Banco 2 na Página 2
    ld   de, MsgBank2
    ld   c, C_WRITE
    call BDOS
    ret

MsgBank2:
    db 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db ">>> [BANCO 2] ROTINA DO BANCO 2 EXECUTADA!", 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "$"

ENDMOD

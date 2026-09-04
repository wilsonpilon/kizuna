; ==============================================================================
; KIZUNA (絆) - Exemplo Multi-Banco: Módulo Principal (Área Comum / Banco 0)
; ==============================================================================

MODULE MAIN
BANK 0

PUBLIC Start
EXTERN PrintBank1, PrintBank2

BDOS       EQU 0x0005
C_WRITE    EQU 0x09

Start:
    ; 1. Mensagem a partir do Banco 0 (Área Comum)
    ld   de, MsgCommon
    ld   c, C_WRITE
    call BDOS

    ; 2. Chamada Cross-Bank: Salta para rotina no Banco 1
    ; O MUSUBI intercepta esta chamada e gera o trampolim automático!
    call PrintBank1

    ; 3. Chamada Cross-Bank: Salta para rotina no Banco 2
    ; O MUSUBI intercepta e troca para o Banco 2 na Página 2!
    call PrintBank2

    ; 4. Mensagem final do Banco 0 e retorno limpo ao MSX-DOS
    ld   de, MsgDone
    ld   c, C_WRITE
    call BDOS

    ret

MsgCommon:
    db 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "[Banco 0 - Area Comum] KIZUNA Multi-Banco Iniciado", 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "$"

MsgDone:
    db 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "[Banco 0] Execucao Multi-Banco Concluida com Sucesso!", 0x0D, 0x0A
    db "==================================================", 0x0D, 0x0A
    db "$"

ENDMOD

; ==============================================================================
; KIZUNA (絆) - Exemplo Mínimo em Assembly Z80 para MSX-DOS 2
; Programa funcional que imprime mensagem na tela via chamada BDOS 09h
; ==============================================================================

MODULE HELLO
BANK 0

PUBLIC Start

; Constantes do MSX-DOS / CP/M
BDOS        EQU 0x0005      ; Ponto de entrada padrão da BDOS
C_WRITESTR  EQU 0x09        ; Função 09h: Imprimir string terminada em '$'

Start:
    ; Carrega o endereço da mensagem em DE
    ld   de, MsgHello

    ; Carrega a função 09h no registrador C
    ld   c, C_WRITESTR

    ; Chama a BDOS para imprimir
    call BDOS

    ; Retorna limpo ao MSX-DOS (mantém o sistema operacional intacto)
    ret

; Dados da mensagem (terminada com '$' conforme exigido pela função 09h da BDOS)
MsgHello:
    db 0x0D, 0x0A
    db "========================================", 0x0D, 0x0A
    db "  KIZUNA (絆) - MSX2+ Z80 Toolchain     ", 0x0D, 0x0A
    db "  Hello World via KAJI80 & MOB Format!  ", 0x0D, 0x0A
    db "========================================", 0x0D, 0x0A
    db "$"

ENDMOD

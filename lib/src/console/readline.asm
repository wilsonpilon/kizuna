; =============================================================================
; KIZUNA MSXLIB - console/readline
; leitura de uma linha do teclado
; =============================================================================

MODULE console_readline
BANK 0

PUBLIC CON_ReadLine
EXTERN BDOS_Call

; CON_ReadLine: le uma linha (ate ENTER) com edicao do MSX-DOS (funcao 0Ah).
; Entrada: DE = buffer de A + 2 bytes, A = maximo de caracteres (1..255)
; Saída: HL = string tamanho+dados DENTRO do buffer (em DE + 1): o byte de
;        tamanho que o MSX-DOS preenche vale como o de uma STR_
; Preserva: BC, DE. Destrói: A, flags.
CON_ReadLine:
    PUSH BC
    PUSH DE
    LD (DE), A
    LD C, 0Ah
    CALL BDOS_Call
    POP DE
    LD H, D
    LD L, E
    INC HL
    POP BC
    RET

ENDMOD

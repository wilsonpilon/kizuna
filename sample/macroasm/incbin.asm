; =============================================================================
; KIZUNA - Teste de hardware: INCBIN "arquivo", SKIP=x, SIZE=y (Fase 8, a
; ultima) do KAJI80.
;
; Inclui o conteudo bruto de greeting.dat (arquivo externo, ao lado deste
; .asm) diretamente no binario -- sem SKIP na primeira vez (imprime o
; cabecalho "HEADERXX" literal, prova que o arquivo inteiro foi incluido
; certo) e com SKIP=8 na segunda vez (pula o cabecalho de 8 bytes, imprime
; só a mensagem de verdade -- prova que SKIP funciona). Compilador: KAJI80
; puro.
; =============================================================================

MODULE IncbinTest
BANK 0
PUBLIC Start

Start:
    LD DE, MsgIntro
    CALLDOS F_STROUT

    LD DE, MsgSemSkip
    CALLDOS F_STROUT
    LD DE, DataSemSkip
    CALLDOS F_STROUT
    LD DE, MsgCRLF
    CALLDOS F_STROUT

    LD DE, MsgComSkip
    CALLDOS F_STROUT
    LD DE, DataComSkip
    CALLDOS F_STROUT
    LD DE, MsgCRLF
    CALLDOS F_STROUT

    LD DE, MsgFim
    CALLDOS F_STROUT
    RET

MsgIntro:    DB "=== KIZUNA: INCBIN \"arquivo\", SKIP=x, SIZE=y ===\r\n$"
MsgSemSkip:  DB "[1] INCBIN sem SKIP (arquivo inteiro, deve comecar com HEADERXX):\r\n$"
MsgComSkip:  DB "[2] INCBIN com SKIP=8 (pula o cabecalho, so a mensagem de verdade):\r\n$"
MsgCRLF:     DB "\r\n$"
MsgFim:      DB "=== fim do teste ===\r\n$"

; O proprio arquivo greeting.dat ja termina em '$' (BDOS_PrintString /
; CALLDOS F_STROUT param no primeiro '$' encontrado), entao cada bloco
; abaixo funciona como uma string normal pra impressao.
DataSemSkip:
    INCBIN "greeting.dat"

DataComSkip:
    INCBIN "greeting.dat", SKIP=8

; =============================================================================
; KIZUNA - Teste de hardware: avaliador de expressoes (Fase 1), rotulos
; locais (Fase 2), IF/ELSE/ENDIF (Fase 3), REPT/ENDR (Fase 4) e MACRO
; (Fase 5) do KAJI80, recursos ao estilo asMSX com sintaxe propria do
; KIZUNA.
;
; Cada verificacao e feita de verdade EM TEMPO DE EXECUCAO (nao so
; confiando no valor calculado em tempo de montagem) -- se o avaliador de
; expressoes tivesse calculado algo errado, o CP/SBC HL,DE abaixo
; detectaria e imprimiria FALHOU, nao OK. Compilador: KAJI80 puro.
; =============================================================================

MODULE ExprLabelsTest
BANK 0
PUBLIC Start

; --- Fase 1: avaliador de expressoes em tempo de montagem ---
VAL_EXPR EQU ((2*8)/(1+3))<<2  ; ((16)/(4))<<2 = 4<<2 = 16
VAL_FIX  EQU FIX(1.5)          ; ponto fixo 8.8: 1.5*256 = 384 (0180h)

TESTFLAG EQU 1

; --- Fase 5: macro simples -- imprime uma string terminada em '$' ---
m_PRINT: MACRO @MSG
    LD DE, @MSG
    CALLDOS F_STROUT
ENDM

Start:
    LD DE, MsgTitle
    CALLDOS F_STROUT

    ; --- Teste 1: EQU com expressao aritmetica + bits (Fase 1) ---
    LD A, VAL_EXPR
    CP 16
    JR NZ, .fail_expr
    m_PRINT MsgExprOK
    JR .done_expr
.fail_expr:
    m_PRINT MsgExprFail
.done_expr:

    ; --- Teste 2: EQU com FIX() -- ponto fixo 8.8 (Fase 1) ---
    LD HL, VAL_FIX
    LD DE, 0180h
    OR A
    SBC HL, DE
    JR NZ, .fail_fix
    m_PRINT MsgFixOK
    JR .done_fix
.fail_fix:
    m_PRINT MsgFixFail
.done_fix:

    ; --- Teste 3: IF/ELSE/ENDIF (Fase 3) -- selecao em tempo de montagem,
    ; so um dos dois ramos abaixo existe de verdade no binario final ---
IF TESTFLAG == 1
    m_PRINT MsgIfOK
ELSE
    m_PRINT MsgIfFail
ENDIF

    ; --- Teste 4: REPT/ENDR (Fase 4) -- deve imprimir exatamente 5 '*' ---
    LD DE, MsgReptIntro
    CALLDOS F_STROUT
REPT 5
    LD E, '*'
    CALLDOS F_CONOUT
ENDR
    LD DE, MsgCRLF
    CALLDOS F_STROUT

    LD DE, MsgFim
    CALLDOS F_STROUT
    RET

MsgTitle:      DB "=== KIZUNA: expressoes/rotulos locais/IF/REPT/MACRO ===\r\n$"
MsgExprOK:     DB "[OK]     ((2*8)/(1+3))<<2 = 16\r\n$"
MsgExprFail:   DB "[FALHOU] expressao aritmetica/bits\r\n$"
MsgFixOK:      DB "[OK]     FIX(1.5) = 0180h\r\n$"
MsgFixFail:    DB "[FALHOU] FIX() ponto fixo\r\n$"
MsgIfOK:       DB "[OK]     IF escolheu o ramo verdadeiro\r\n$"
MsgIfFail:     DB "[FALHOU] IF escolheu o ramo errado\r\n$"
MsgReptIntro:  DB "[REPT]   deve aparecer exatamente 5 asteriscos: $"
MsgCRLF:       DB "\r\n$"
MsgFim:        DB "=== fim do teste ===\r\n$"

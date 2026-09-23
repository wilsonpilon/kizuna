; =============================================================================
; KIZUNA - Teste de hardware: rotulos pre-definidos de BIOS/BDOS (Fase 6)
; e CALLBIOS/CALLDOS (Fase 7) do KAJI80.
;
; CALLBIOS CHGMOD troca o modo de tela de verdade via chamada inter-slot
; real (LD IX,CHGMOD / CALL BIOS_Call, reaproveitando a rotina ja
; hardware-testada de lib/src/bios.asm) -- a tela vai mudar visivelmente
; duas vezes: SCREEN 1 (fundo diferente) e de volta pra SCREEN 0. Compila
; SEM precisar de EXTERN nem INCLUDE pra CHGMOD/F_STROUT -- ambos vem das
; tabelas pre-definidas.
; =============================================================================

MODULE PredefinedTest
BANK 0
PUBLIC Start

Start:
    LD DE, MsgIntro
    CALLDOS F_STROUT

    LD DE, MsgAntes
    CALLDOS F_STROUT

    ; --- CALLBIOS com rotulo pre-definido de BIOS: troca pra SCREEN 1 ---
    LD A, 1
    CALLBIOS CHGMOD

    LD DE, MsgScreen1
    CALLDOS F_STROUT

    ; pequena pausa visual (loop vazio) antes de voltar pro modo original
    LD BC, 0FFFFh
.pausa1:
    DEC BC
    LD A, B
    OR C
    JR NZ, .pausa1

    ; --- de volta pro modo de texto padrao do MSX-DOS (SCREEN 0) ---
    LD A, 0
    CALLBIOS CHGMOD

    LD DE, MsgScreen0
    CALLDOS F_STROUT

    LD DE, MsgFim
    CALLDOS F_STROUT
    RET

MsgIntro:   DB "=== KIZUNA: rotulos pre-definidos + CALLBIOS/CALLDOS ===\r\n$"
MsgAntes:   DB "Trocando pra SCREEN 1 via CALLBIOS CHGMOD...\r\n$"
MsgScreen1: DB "[OK] Voce deveria estar vendo SCREEN 1 agora (40->32 colunas)\r\n$"
MsgScreen0: DB "[OK] De volta pro SCREEN 0 (modo padrao do MSX-DOS)\r\n$"
MsgFim:     DB "=== fim do teste -- se a tela mudou duas vezes, CALLBIOS funcionou ===\r\n$"

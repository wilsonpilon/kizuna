; =============================================================================
; KIZUNA MSXLIB - vdp/pset
; pixel em SCREEN 2 (wrapper atomico com DI/EI)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_pset
BANK 0

PUBLIC VDP_PSet
EXTERN VDP_PSet_Raw

; -----------------------------------------------------------------------------
; VDP_PSet: Plota um pixel no modo gráfico SCREEN 2 (256x192)
; Entrada: BC = X (0..255), DE = Y (0..191), A = Cor (0..15)
; Preserva: BC, DE, HL
; -----------------------------------------------------------------------------
VDP_PSet:
    ; Wrapper publico: torna UMA chamada isolada atomica. Rotinas que
    ; plotam MUITOS pontos em sequencia (VDP_Line, VDP_BoxFill) NAO devem
    ; usar este wrapper por ponto -- preferem chamar VDP_PSet_Raw
    ; diretamente dentro do PROPRIO DI/EI (unico, cobrindo o laco inteiro),
    ; para nao reabrir interrupcoes a cada pixel.
    ;
    ; Implementacao: calculo manual de endereco na Pattern/Name/Color
    ; Table (ver VDP_PSet_Raw). O motor de comando de hardware do
    ; V9938/V9958 (registradores 36-46) NAO funciona em SCREEN 2 -- so
    ; opera nos modos bitmap Graphic 4-7 (SCREEN 5-8) -- por isso nao e
    ; usado aqui; ficou implementado em VDP_PSet_HW para quando a MSXLIB
    ; ganhar suporte a esses modos.
    DI
    CALL VDP_PSet_Raw
    EI
    RET

ENDMOD

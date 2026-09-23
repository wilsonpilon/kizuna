; =============================================================================
; KIZUNA MSXLIB - vdp/psetraw
; pixel em SCREEN 2 sem DI/EI (calcula Pattern/Color na mao)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_psetraw
BANK 0

PUBLIC VDP_PSet_Raw
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_PSet_Raw: calcula a celula da Pattern/Name/Color Table na mao e
; escreve os bytes diretamente -- a UNICA tecnica que funciona em SCREEN 2
; (ver nota em VDP_PSet_HW acima sobre o motor de comando nao se aplicar
; aqui). Sem DI/EI proprios -- deve ser chamada apenas com interrupcoes ja
; desabilitadas pelo chamador (VDP_PSet acima, ou o DI unico que VDP_Line/
; VDP_BoxFill colocam em volta do laco inteiro de multiplos pontos).
; Entrada/Saida: identicas ao VDP_PSet.
; -----------------------------------------------------------------------------
VDP_PSet_Raw:
    ; A (Cor) e sobrescrito pelos calculos de endereco logo abaixo;
    ; precisa ser salvo antes de qualquer outra coisa.
    LD (VDP_PSet_ColorArg), A
    PUSH BC
    PUSH DE
    PUSH HL

    ; HL = endereço do padrão selecionado pela Name Table.
    ; SCREEN 2 possui uma página de padrões para cada faixa vertical de 64 px.
    ; Índice = ((Y >> 3) * 32 + (X >> 3)) & 0xFF; endereço = página + índice * 8 + (Y & 7).
    LD A, E
    AND 0C0h
    RRCA
    RRCA
    RRCA
    LD D, A
    LD A, E
    RRCA
    RRCA
    RRCA
    AND 07h
    ADD A, A
    ADD A, A
    ADD A, A
    ADD A, A
    ADD A, A
    LD L, A
    LD A, C
    SRL A
    SRL A
    SRL A
    ADD A, L
    LD L, A
    LD H, 00h
    SLA L
    RL H
    SLA L
    RL H
    SLA L
    RL H
    LD A, E
    AND 07h
    ADD A, L
    LD L, A
    LD A, D
    ADD A, H
    LD H, A

    ; Bitmask: bit 7 - (X & 7)
    LD A, C
    AND 07h
    LD B, A
    LD A, 80h
VDP_PSet_MaskLoop:
    DEC B
    JP M, VDP_PSet_MaskDone
    RRCA
    JR VDP_PSet_MaskLoop
VDP_PSet_MaskDone:
    LD C, A

    ; Le o byte atual do padrao e funde (OR) com a mascara do pixel antes
    ; de regravar -- cada byte do padrao cobre 8 pixels horizontais da
    ; mesma linha da celula 8x8; gravar so a mascara (como antes) apagava
    ; os outros 7 pixels dessa linha a cada chamada, sobrando 1 pixel a
    ; cada 8 em qualquer LINE/BOXFILL (a Color Table ja fazia esse mesmo
    ; read-modify-write logo abaixo -- faltava aqui tambem).
    ; Endereco de LEITURA (sem OR 40h) -- sequencia de VDP_SetReadAddr inlined.
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OUT (VDP_CMD), A
    NOP
    NOP
    NOP
    NOP
    NOP
    NOP
    NOP
    IN A, (VDP_DATA)
    OR C
    LD C, A

    ; Endereco de ESCRITA (com OR 40h) -- reaponta para o mesmo byte do padrao.
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OR 40h
    OUT (VDP_CMD), A
    NOP
    NOP
    NOP
    NOP
    NOP
    NOP
    NOP
    LD A, C
    OUT (VDP_DATA), A

    ; Atualiza o nibble de frente (foreground) do byte correspondente na
    ; Color Table, preservando o nibble de fundo (background) existente.
    ; A Color Table tem exatamente o mesmo layout/offset da Pattern
    ; Generator Table, só que baseada em 2000h em vez de 0000h — por isso
    ; HL (ainda válido aqui, intocado pela sequencia acima) só precisa
    ; somar 2000h para apontar para a célula de cor correspondente.
    LD DE, 2000h
    ADD HL, DE

    ; Endereco de LEITURA (sem OR 40h) -- sequencia de VDP_SetReadAddr inlined.
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OUT (VDP_CMD), A
    NOP
    NOP
    NOP
    NOP
    NOP
    NOP
    NOP
    IN A, (VDP_DATA)
    AND 0Fh
    LD B, A
    LD A, (VDP_PSet_ColorArg)
    AND 0Fh
    SLA A
    SLA A
    SLA A
    SLA A
    OR B
    LD C, A

    ; Endereco de ESCRITA (com OR 40h) -- reaponta para o mesmo byte de cor.
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OR 40h
    OUT (VDP_CMD), A
    NOP
    NOP
    NOP
    NOP
    NOP
    NOP
    NOP
    LD A, C
    OUT (VDP_DATA), A

    POP HL
    POP DE
    POP BC
    RET

VDP_PSet_ColorArg: DB 00h

ENDMOD

; =============================================================================
; KIZUNA MSXLIB - vdp/state
; estado do VDP: copia sombra dos registradores, modo atual e paleta
; =============================================================================

MODULE vdp_state
BANK 0


PUBLIC VDP_Shadow, VDP_Mode, VDP_PalShadow

; Os registradores do VDP so aceitam escrita: nao da para "ler R#1, mudar um bit e
; escrever de volta". A MSXLIB guarda uma COPIA de cada registrador escrito pelas
; rotinas VDP_SetReg / VDP_UpdateReg / VDP_SetMode... e altera bits a partir dela.
; (Escritas feitas por outros meios, como VDP_WriteReg ou pela BIOS, nao passam
; pela copia: use VDP_SetReg se depois quiser VDP_UpdateReg no mesmo registrador.)
VDP_Shadow:
    DS 64               ; R#0..R#63
VDP_Mode:
    DB 0FFh             ; ultimo modo ajustado por VDP_SetMode (0FFh = nenhum)
VDP_PalShadow:
    DS 32               ; 16 cores x 2 bytes: (R << 4) | B, depois G

ENDMOD

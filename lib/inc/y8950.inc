; =============================================================================
; KIZUNA MSXLIB - MSX-AUDIO (Y8950 / OPL com ADPCM)
; Portas: PORT_Y8950_ADDR (0C0h) e PORT_Y8950_DATA (0C1h), em ports.inc.
; =============================================================================

; --- Registradores gerais ----------------------------------------------------
Y8950_R_TEST      EQU 01h   ; bit 5 = habilita selecao de forma de onda (E0h-F5h)
Y8950_R_TIMER1    EQU 02h
Y8950_R_TIMER2    EQU 03h
Y8950_R_FLAGCTRL  EQU 04h   ; controle de flags/interrupcoes
Y8950_R_KBD_IN    EQU 05h   ; teclado (porta do cartucho)
Y8950_R_KBD_OUT   EQU 06h
Y8950_R_ADPCM_CTL EQU 07h
Y8950_R_MODE      EQU 08h   ; CSM / selecao de teclado
Y8950_R_START_LO  EQU 09h   ; endereco inicial do ADPCM
Y8950_R_START_HI  EQU 0Ah
Y8950_R_STOP_LO   EQU 0Bh
Y8950_R_STOP_HI   EQU 0Ch
Y8950_R_PRESC_LO  EQU 0Dh
Y8950_R_PRESC_HI  EQU 0Eh
Y8950_R_ADPCM_DAT EQU 0Fh
Y8950_R_DELTA_LO  EQU 10h
Y8950_R_DELTA_HI  EQU 11h
Y8950_R_ADPCM_LVL EQU 12h   ; nivel de saida do ADPCM
Y8950_R_DAC_LO    EQU 15h
Y8950_R_DAC_HI    EQU 16h
Y8950_R_DAC_SHIFT EQU 17h
Y8950_R_IO_CTRL   EQU 18h
Y8950_R_IO_DATA   EQU 19h
Y8950_R_PCM_DATA  EQU 1Ah

; --- Registradores por operador (18 operadores, 9 vozes x modulador/portadora)
; O registrador de um operador e BASE + DESLOCAMENTO, com os deslocamentos
;   op 1..6  = 00h..05h     op 7..12 = 08h..0Dh     op 13..18 = 10h..15h
; Voz 1 = op 1 (modulador) e op 4 (portadora), voz 2 = op 2 e 5, voz 3 = op 3 e 6;
; vozes 4-6 = ops 7-12, vozes 7-9 = ops 13-18, no mesmo padrao.
Y8950_R_OP_AMVIB   EQU 20h  ; AM / VIB / EG / KSR / multiplicador
Y8950_R_OP_LEVEL   EQU 40h  ; KSL / nivel total
Y8950_R_OP_ATKDEC  EQU 60h  ; ataque / decaimento
Y8950_R_OP_SUSREL  EQU 80h  ; sustentacao / liberacao
Y8950_R_OP_WAVE    EQU 0E0h ; forma de onda (com Y8950_R_TEST bit 5 ligado)

; --- Registradores por voz (9 vozes: base + 0..8) ---------------------------
Y8950_R_FNUM_LO    EQU 0A0h ; F-number, byte baixo
Y8950_R_KEY_BLOCK  EQU 0B0h ; bit5 = KEY ON, bits2-4 = oitava, bits0-1 = F-number alto
Y8950_R_FB_CONN    EQU 0C0h ; realimentacao / conexao
Y8950_R_RHYTHM     EQU 0BDh ; profundidade AM/VIB, modo ritmo, disparo das pecas

Y8950_KEY_ON       EQU 20h

; Ritmo (R#BDh)
Y8950_RHYTHM_ON    EQU 20h
Y8950_RHYTHM_BD    EQU 10h
Y8950_RHYTHM_SD    EQU 08h
Y8950_RHYTHM_TOM   EQU 04h
Y8950_RHYTHM_CYM   EQU 02h
Y8950_RHYTHM_HH    EQU 01h

; --- Bits do registrador de status (leitura da porta de indice) -------------
Y8950_S_PCM_BUSY   EQU 01h
Y8950_S_BUF_RDY    EQU 08h
Y8950_S_EOS        EQU 10h  ; fim da amostra
Y8950_S_TIMER2     EQU 20h
Y8950_S_TIMER1     EQU 40h
Y8950_S_IRQ        EQU 80h

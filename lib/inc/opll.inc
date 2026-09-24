; =============================================================================
; KIZUNA MSXLIB - MSX-MUSIC (YM2413 / OPLL)
; Portas: PORT_OPLL_ADDR (7Ch) e PORT_OPLL_DATA (7Dh), em ports.inc.
; Escrever no dado exige uma pausa depois do indice (~12 ciclos de chip; no Z80
; do MSX, ao menos 4 instrucoes NOP-equivalentes) e outra depois do dado (~84).
; =============================================================================
OPLL_R_USER0     EQU 00h    ; instrumento do usuario: 8 registradores (00h-07h)
OPLL_R_USER7     EQU 07h
OPLL_R_RHYTHM    EQU 0Eh    ; modo ritmo + disparo de bateria
OPLL_R_TEST      EQU 0Fh
OPLL_R_FNUM_LO   EQU 10h    ; F-number, byte baixo (10h-18h = canais 1-9)
OPLL_R_KEY_BLOCK EQU 20h    ; bit0 = F-number bit 8, bits1-3 = oitava, bit4 = KEY ON, bit5 = sustain (20h-28h)
OPLL_R_INST_VOL  EQU 30h    ; nibble alto = instrumento, baixo = volume (30h-38h)

OPLL_KEY_ON      EQU 10h
OPLL_SUSTAIN     EQU 20h

; Ritmo (R#0Eh)
OPLL_RHYTHM_ON   EQU 20h    ; canais 7-9 viram 5 pecas de percussao
OPLL_RHYTHM_BD   EQU 10h    ; bumbo
OPLL_RHYTHM_SD   EQU 08h    ; caixa
OPLL_RHYTHM_TOM  EQU 04h
OPLL_RHYTHM_CYM  EQU 02h    ; prato
OPLL_RHYTHM_HH   EQU 01h    ; chimbal

; Instrumentos de fabrica (nibble alto de 30h-38h); 0 = instrumento do usuario
OPLL_INST_USER      EQU 0
OPLL_INST_VIOLIN    EQU 1
OPLL_INST_GUITAR    EQU 2
OPLL_INST_PIANO     EQU 3
OPLL_INST_FLUTE     EQU 4
OPLL_INST_CLARINET  EQU 5
OPLL_INST_OBOE      EQU 6
OPLL_INST_TRUMPET   EQU 7
OPLL_INST_ORGAN     EQU 8
OPLL_INST_HORN      EQU 9
OPLL_INST_SYNTH     EQU 10
OPLL_INST_HARPSI    EQU 11
OPLL_INST_VIBRAPH   EQU 12
OPLL_INST_SYNBASS   EQU 13
OPLL_INST_ACBASS    EQU 14
OPLL_INST_ELGUITAR  EQU 15

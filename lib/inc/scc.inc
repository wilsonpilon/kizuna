; =============================================================================
; KIZUNA MSXLIB - Konami SCC (mapeado em memoria, pagina 2: 8000h-BFFFh)
; O SCC so aparece em 9800h-98FFh depois de selecionar o banco dele: gravar
; 3Fh em 9000h (SCC padrao). O SCC+ tem modo proprio (registro em 0BFFEh) e
; mapa estendido -- entra numa fase posterior.
; =============================================================================
SCC_BANK_SELECT  EQU 9000h  ; gravar SCC_BANK_VALUE aqui expoe o SCC em 9800h
SCC_BANK_VALUE   EQU 3Fh

SCC_BASE         EQU 9800h
SCC_WAVE1        EQU 9800h  ; 32 amostras de 8 bits com sinal, por canal
SCC_WAVE2        EQU 9820h
SCC_WAVE3        EQU 9840h
SCC_WAVE4        EQU 9860h  ; o canal 5 divide esta forma de onda no SCC padrao
SCC_FREQ1        EQU 9880h  ; periodo de 12 bits: byte baixo, depois nibble alto
SCC_FREQ2        EQU 9882h
SCC_FREQ3        EQU 9884h
SCC_FREQ4        EQU 9886h
SCC_FREQ5        EQU 9888h
SCC_VOL1         EQU 988Ah  ; volume 0-15
SCC_VOL2         EQU 988Bh
SCC_VOL3         EQU 988Ch
SCC_VOL4         EQU 988Dh
SCC_VOL5         EQU 988Eh
SCC_MIXER        EQU 988Fh  ; bit n liga o canal n+1
SCC_TEST         EQU 98C0h

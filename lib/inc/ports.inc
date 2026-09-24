; =============================================================================
; KIZUNA MSXLIB - portas de I/O do MSX (nomes PORT_*)
; Escrito a partir da lista de portas padrao do MSX; conferido contra o
; mapa publico de portas (map.grauw.nl / msx.org "I/O Ports List").
; As constantes especificas de cada chip ficam em vdpreg.inc, psgreg.inc, ...
; =============================================================================

; --- V9938/V9958 (VDP) -------------------------------------------------------
PORT_VDP_DATA   EQU 98h     ; leitura/escrita de dados da VRAM
PORT_VDP_CMD    EQU 99h     ; escrita: endereco/registrador; leitura: status
PORT_VDP_PAL    EQU 9Ah     ; escrita de paleta (2 bytes por cor)
PORT_VDP_IREG   EQU 9Bh     ; escrita indireta de registrador (R#17)

; --- PSG AY-3-8910 / YM2149 --------------------------------------------------
PORT_PSG_ADDR   EQU 0A0h    ; seleciona o registrador
PORT_PSG_WRITE  EQU 0A1h    ; escreve o valor
PORT_PSG_READ   EQU 0A2h    ; le o valor
PORT_PSG_EXT_ADDR  EQU 10h  ; PSG externo (cartuchos), mesmas funcoes
PORT_PSG_EXT_WRITE EQU 11h
PORT_PSG_EXT_READ  EQU 12h

; --- PPI 8255: slots, teclado, LEDs, som ------------------------------------
PORT_PPI_A      EQU 0A8h    ; selecao de slot primario das 4 paginas
PORT_PPI_B      EQU 0A9h    ; leitura da linha do teclado selecionada
PORT_PPI_C      EQU 0AAh    ; bits 0-3 = linha do teclado, 4 = motor, 5 = Caps, 6 = click, 7 = fita
PORT_PPI_CTRL   EQU 0ABh    ; modo/bit set-reset do 8255

; --- Relogio RP-5C01 ---------------------------------------------------------
PORT_RTC_ADDR   EQU 0B4h
PORT_RTC_DATA   EQU 0B5h

; --- Impressora --------------------------------------------------------------
PORT_LPT_STATUS EQU 90h     ; leitura: bit 1 = ocupada; escrita: bit 0 = strobe
PORT_LPT_DATA   EQU 91h

; --- Kanji ROM ---------------------------------------------------------------
PORT_KANJI1_LO  EQU 0D8h
PORT_KANJI1_HI  EQU 0D9h
PORT_KANJI2_LO  EQU 0DAh
PORT_KANJI2_HI  EQU 0DBh

; --- Memory mapper: segmento de cada pagina de 16 KB -------------------------
PORT_MAPPER_P0  EQU 0FCh
PORT_MAPPER_P1  EQU 0FDh
PORT_MAPPER_P2  EQU 0FEh
PORT_MAPPER_P3  EQU 0FFh

; --- MSX Turbo R -------------------------------------------------------------
PORT_TR_TIMER_LO EQU 0E6h   ; temporizador de 16 bits (escrita zera; leitura le)
PORT_TR_TIMER_HI EQU 0E7h

; --- Cartuchos de som --------------------------------------------------------
PORT_OPLL_ADDR  EQU 7Ch     ; MSX-MUSIC (YM2413): indice do registrador
PORT_OPLL_DATA  EQU 7Dh     ;                     dado
PORT_Y8950_ADDR EQU 0C0h    ; MSX-AUDIO (Y8950): indice do registrador
PORT_Y8950_DATA EQU 0C1h    ;                    dado
PORT_Y8950_ADDR2 EQU 0C2h   ; segunda interface (quando ha dois cartuchos)
PORT_Y8950_DATA2 EQU 0C3h

; --- Endereco (de memoria, nao de porta) do registrador de subslot ----------
SLOT_SUBSLOT_REG EQU 0FFFFh ; ao ler, devolve o complemento do valor gravado

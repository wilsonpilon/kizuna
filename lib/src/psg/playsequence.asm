; =============================================================================
; KIZUNA MSXLIB - psg/playsequence
; toca uma sequencia de notas
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE psg_playsequence
BANK 0

PUBLIC PSG_PlaySequence
EXTERN PSG_PlayNoteIndexed

; -----------------------------------------------------------------------------
; PSG_PlaySequence: Toca uma sequência de notas em RAM/ROM, uma após a outra
; Entrada: HL = ponteiro para a sequência: 4 bytes por evento (canal,
;          índice-de-nota, volume, duração), terminada por um byte de canal
;          = 0FFh. Duração é um contador de laço, não milissegundos -- use
;          valores maiores para notas mais longas; 0 se comporta como 256
;          (laço de 8 bits decrementando até dar a volta).
; -----------------------------------------------------------------------------
PSG_PlaySequence:
PSG_Seq_Loop:
    LD A, (HL)
    CP 0FFh
    RET Z

    LD A, (HL) ; canal
    INC HL
    LD D, (HL) ; índice da nota
    INC HL
    LD E, (HL) ; volume
    INC HL
    LD C, (HL) ; duração
    INC HL
    PUSH HL    ; salva o ponteiro da sequência (única PUSH desta iteração)

    CALL PSG_PlayNoteIndexed

PSG_Seq_DelayOuter:
    LD B, 00h
PSG_Seq_DelayInner:
    DJNZ PSG_Seq_DelayInner
    DEC C
    JR NZ, PSG_Seq_DelayOuter

    POP HL     ; restaura o ponteiro da sequência
    JR PSG_Seq_Loop

ENDMOD

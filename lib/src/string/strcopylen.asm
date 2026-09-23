; =============================================================================
; KIZUNA MSXLIB - string/strcopylen
; copia N bytes
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE string_strcopylen
BANK 0

PUBLIC StrCopyLen

; -----------------------------------------------------------------------------
; StrCopyLen: Copia uma string no formato "curto" (1 byte de tamanho + até 255
; bytes de dados, ver SPEC.md §7) da origem para o destino -- usado por DIGNAC
; para atribuição de variável STRING (s$ = ...), que não é um simples LD de
; 2 bytes como INTEGER.
; Entrada: HL = origem, DE = destino (ambos no formato tamanho+dados)
; -----------------------------------------------------------------------------
StrCopyLen:
    PUSH AF
    PUSH BC
    PUSH HL
    PUSH DE

    LD A, (HL)  ; byte de tamanho (0..255 bytes de dados)
    LD B, A     ; B = contador de bytes de DADOS a copiar
    LD (DE), A  ; copia o próprio byte de tamanho pro destino
    INC HL
    INC DE

    LD A, B
    OR A
    JR Z, StrCopyLen_Done ; tamanho 0: nenhum dado a copiar (B=0 não pode ir pro DJNZ, viraria 256 iterações)
StrCopyLen_Loop:
    LD A, (HL)
    LD (DE), A
    INC HL
    INC DE
    DJNZ StrCopyLen_Loop
StrCopyLen_Done:

    POP DE
    POP HL
    POP BC
    POP AF
    RET

ENDMOD

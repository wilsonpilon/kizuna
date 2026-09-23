; =============================================================================
; KIZUNA MSXLIB - string/printdec16buf
; decimal de 16 bits num buffer
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE string_printdec16buf
BANK 0

PUBLIC PrintDec16ToBuffer

; -----------------------------------------------------------------------------
; PrintDec16ToBuffer: Formata número não sinalizado de 16 bits em HL como
; decimal (0..65535) num buffer em RAM em vez do console -- usado por DIGNAC
; para "PRINT #n, expr" (BDOS_FileWrite espera um buffer de bytes, não uma
; chamada por caractere). Suprime zeros à esquerda; não escreve nenhum
; terminador (nem \0 nem $), só os dígitos.
; Entrada: HL = valor de 16 bits, DE = buffer de destino
; Saída: A = quantidade de dígitos escritos (1..5)
; Preserva: BC, DE, HL
; -----------------------------------------------------------------------------
PrintDec16ToBuffer:
    PUSH BC
    PUSH DE
    PUSH HL

    EX DE, HL ; HL = buffer, DE = valor (LD (nn),DE não existe no Z80; só (nn),HL)
    LD (PDTB_BufPtr), HL
    EX DE, HL ; HL = valor (de volta), DE = buffer (não precisa mais aqui)
    XOR A
    LD (PDTB_Count), A
    LD C, 00h ; C = 0 indica supressão de zeros à esquerda

    LD DE, 2710h
    CALL PrintDecDigitBuf
    LD DE, 03E8h
    CALL PrintDecDigitBuf
    LD DE, 0064h
    CALL PrintDecDigitBuf
    LD DE, 000Ah
    CALL PrintDecDigitBuf

    ; Dígito das unidades (1): escreve sempre (0..9)
    LD A, L
    ADD A, 30h
    CALL PDTB_WriteChar

    POP HL
    POP DE
    POP BC
    ; Lê o resultado de PDTB_Count SÓ DEPOIS de restaurar BC do chamador --
    ; guardar a contagem num registrador (ex: B) antes do POP BC seria
    ; destruído pelo próprio POP (B faz parte do par BC restaurado ali).
    ; Bug real desta forma já aconteceu aqui: a contagem escrita ficava
    ; certa no buffer, mas o valor devolvido em A era o B original do
    ; chamador (lixo), não a contagem -- fazendo BDOS_FileWrite escrever
    ; um tamanho errado (por coincidência às vezes "certo", às vezes não).
    LD A, (PDTB_Count)
    RET

; Subrotina interna: mesma lógica de PrintDecDigit, só que escrevendo no
; buffer (via PDTB_WriteChar) em vez de chamar BDOS_PrintChar
PrintDecDigitBuf:
    LD B, 2Fh ; '0' - 1
PDDB_Loop:
    INC B
    OR A
    SBC HL, DE
    JR NC, PDDB_Loop
    ADD HL, DE ; Restaura último excesso

    LD A, B
    CP 30h ; '0'
    JR NZ, PDDB_Print
    LD A, C
    OR A
    RET Z ; suprime o zero à esquerda

    LD A, 30h
PDDB_Print:
    LD C, 01h
    CALL PDTB_WriteChar
    RET

; Subrotina interna: escreve o caractere em A no buffer apontado por
; PDTB_BufPtr, avança o ponteiro e incrementa PDTB_Count
PDTB_WriteChar:
    PUSH HL
    LD HL, (PDTB_BufPtr)
    LD (HL), A
    INC HL
    LD (PDTB_BufPtr), HL
    POP HL

    PUSH AF
    LD A, (PDTB_Count)
    INC A
    LD (PDTB_Count), A
    POP AF
    RET

PDTB_BufPtr: DW 0000h
PDTB_Count:  DB 00h

ENDMOD

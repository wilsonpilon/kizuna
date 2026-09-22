; ==============================================================================
; KIZUNA sample -- I/O de arquivo (BDOS_FileCreate/Open/Write/Read/Close)
; Compilador: KAJI80
; Cria TEST.TXT, escreve um texto, fecha, reabre para leitura, lê de volta e
; imprime o conteúdo lido -- prova de ponta a ponta das funções de arquivo
; baseadas em handle do MSX-DOS 2.
; ==============================================================================

MODULE FILEIODEMO
BANK 0

PUBLIC Start
EXTERN BDOS_FileCreate, BDOS_FileOpen, BDOS_FileClose, BDOS_FileWrite, BDOS_FileRead

BDOS    EQU 0005h
C_WRITE EQU 09h

Start:
    LD DE, MsgIntro
    LD C, C_WRITE
    CALL BDOS

    ; 1. Cria (ou trunca) TEST.TXT em modo leitura+escrita
    LD DE, FileName
    LD A, 00h ; modo: leitura + escrita
    LD B, 00h ; atributo normal
    CALL BDOS_FileCreate
    OR A
    JR NZ, FileIO_Error
    LD A, B
    LD (Handle), A

    ; 2. Escreve o conteúdo
    LD A, (Handle)
    LD B, A
    LD DE, WriteBuf
    LD HL, 66
    CALL BDOS_FileWrite
    OR A
    JR NZ, FileIO_Error

    ; 3. Fecha
    LD A, (Handle)
    LD B, A
    CALL BDOS_FileClose
    OR A
    JR NZ, FileIO_Error

    ; 4. Reabre só para leitura
    LD DE, FileName
    LD A, 01h ; modo: sem escrita (bit0)
    CALL BDOS_FileOpen
    OR A
    JR NZ, FileIO_Error
    LD A, B
    LD (Handle), A

    ; 5. Lê de volta
    LD A, (Handle)
    LD B, A
    LD DE, ReadBuf
    LD HL, 80
    CALL BDOS_FileRead
    OR A
    JR NZ, FileIO_Error

    ; HL = bytes realmente lidos; usa para terminar a string em '$' (função
    ; 09h da BDOS exige terminador '$', o arquivo em si não tem um)
    LD DE, ReadBuf
    ADD HL, DE
    LD (HL), '$'

    ; 6. Fecha de novo
    LD A, (Handle)
    LD B, A
    CALL BDOS_FileClose

    ; 7. Imprime o que foi lido do arquivo
    LD DE, MsgRead
    LD C, C_WRITE
    CALL BDOS
    LD DE, ReadBuf
    LD C, C_WRITE
    CALL BDOS

    LD DE, MsgDone
    LD C, C_WRITE
    CALL BDOS
    RET

FileIO_Error:
    LD DE, MsgError
    LD C, C_WRITE
    CALL BDOS
    RET

MsgIntro:
    DB 0Dh, 0Ah
    DB "KIZUNA sample -- I/O de arquivo (BDOS_File*)", 0Dh, 0Ah
    DB "$"

MsgRead:
    DB 0Dh, 0Ah
    DB "Conteudo lido de volta de TEST.TXT:", 0Dh, 0Ah
    DB "$"

MsgDone:
    DB 0Dh, 0Ah
    DB "Concluido com sucesso.", 0Dh, 0Ah
    DB "$"

MsgError:
    DB 0Dh, 0Ah
    DB "Erro de I/O de arquivo!", 0Dh, 0Ah
    DB "$"

FileName: DB "TEST.TXT", 00h
WriteBuf: DB "Ola do KIZUNA! Este arquivo foi criado, escrito e lido de volta.", 0Dh, 0Ah

Handle:  DB 00h
ReadBuf: DS 81

ENDMOD

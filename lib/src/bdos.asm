; =============================================================================
; KIZUNA MSXLIB - BDOS.ASM
; Rotinas de suporte para chamadas ao kernel do MSX-DOS (BDOS 0x0005)
; =============================================================================

MODULE BDOS
BANK 0

PUBLIC BDOS_Call, BDOS_PrintChar, BDOS_PrintString, BDOS_ReadChar, BDOS_Exit
PUBLIC BDOS_FileOpen, BDOS_FileCreate, BDOS_FileClose, BDOS_FileRead
PUBLIC BDOS_FileWrite, BDOS_FileSeek

BDOS_ENTRY EQU 0005h

; Funções de I/O de arquivo baseado em handle do MSX-DOS 2 (não o FCB antigo
; de DOS 1) -- números confirmados contra o protocolo oficial de MSX-DOS 2
; (map.grauw.nl/resources/dos2_functioncalls.php), diferentes dos números
; genéricos de CP/M/MS-DOS.
BDOS_F_OPEN   EQU 43h
BDOS_F_CREATE EQU 44h
BDOS_F_CLOSE  EQU 45h
BDOS_F_READ   EQU 48h
BDOS_F_WRITE  EQU 49h
BDOS_F_SEEK   EQU 4Ah

; Modo de abertura (byte A de BDOS_FileOpen/BDOS_FileCreate)
BDOS_FMODE_RW     EQU 00h ; leitura e escrita
BDOS_FMODE_RDONLY EQU 01h ; bit0 = sem escrita
BDOS_FMODE_WRONLY EQU 02h ; bit1 = sem leitura

; Atributo de criação (byte B de BDOS_FileCreate)
BDOS_ATTR_NORMAL     EQU 00h ; cria ou trunca arquivo existente
BDOS_ATTR_CREATE_NEW EQU 80h ; bit7 = falha se o arquivo já existir

; -----------------------------------------------------------------------------
; BDOS_Call: Executa chamada direta ao BDOS com função em C
; Entrada: C = número da função BDOS, registradores conforme a função
; -----------------------------------------------------------------------------
BDOS_Call:
    CALL BDOS_ENTRY
    RET

; -----------------------------------------------------------------------------
; BDOS_PrintChar: Imprime um único caractere no console (Função 02h)
; Entrada: E = código ASCII do caractere
; Preserva: todos os registradores
; -----------------------------------------------------------------------------
BDOS_PrintChar:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    PUSH IX
    PUSH IY
    LD C, 02h
    CALL BDOS_ENTRY
    POP IY
    POP IX
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; -----------------------------------------------------------------------------
; BDOS_PrintString: Imprime string terminada em '$' no console (Função 09h)
; Entrada: DE = ponteiro para a string terminada em '$'
; Preserva: todos os registradores
; -----------------------------------------------------------------------------
BDOS_PrintString:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    PUSH IX
    PUSH IY
    LD C, 09h
    CALL BDOS_ENTRY
    POP IY
    POP IX
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; -----------------------------------------------------------------------------
; BDOS_ReadChar: Lê um caractere do teclado com eco (Função 01h)
; Saída: A = caractere lido
; -----------------------------------------------------------------------------
BDOS_ReadChar:
    PUSH BC
    LD C, 01h
    CALL BDOS_ENTRY
    POP BC
    RET

; -----------------------------------------------------------------------------
; BDOS_Exit: Termina o programa e retorna ao MSX-DOS (Função 00h)
; -----------------------------------------------------------------------------
BDOS_Exit:
    LD C, 00h
    CALL BDOS_ENTRY
    RET

; -----------------------------------------------------------------------------
; BDOS_FileOpen: Abre um arquivo existente (Função 43h, MSX-DOS 2)
; Entrada: DE = ponteiro para caminho ASCIIZ, A = modo (BDOS_FMODE_*)
; Saída: A = código de erro (0 = sucesso), B = handle do arquivo
; -----------------------------------------------------------------------------
BDOS_FileOpen:
    LD C, BDOS_F_OPEN
    CALL BDOS_ENTRY
    RET

; -----------------------------------------------------------------------------
; BDOS_FileCreate: Cria (ou trunca) um arquivo (Função 44h, MSX-DOS 2)
; Entrada: DE = ponteiro para caminho ASCIIZ, A = modo (BDOS_FMODE_*),
;          B = atributos (BDOS_ATTR_*)
; Saída: A = código de erro (0 = sucesso), B = handle do arquivo
; -----------------------------------------------------------------------------
BDOS_FileCreate:
    LD C, BDOS_F_CREATE
    CALL BDOS_ENTRY
    RET

; -----------------------------------------------------------------------------
; BDOS_FileClose: Fecha um handle de arquivo aberto (Função 45h, MSX-DOS 2)
; Entrada: B = handle do arquivo
; Saída: A = código de erro (0 = sucesso)
; -----------------------------------------------------------------------------
BDOS_FileClose:
    LD C, BDOS_F_CLOSE
    CALL BDOS_ENTRY
    RET

; -----------------------------------------------------------------------------
; BDOS_FileRead: Lê bytes de um handle de arquivo (Função 48h, MSX-DOS 2)
; Entrada: B = handle, DE = buffer de destino, HL = quantidade de bytes a ler
; Saída: A = código de erro (0 = sucesso), HL = bytes realmente lidos
;        (menor que o pedido, ou 0, indica fim de arquivo)
; -----------------------------------------------------------------------------
BDOS_FileRead:
    LD C, BDOS_F_READ
    CALL BDOS_ENTRY
    RET

; -----------------------------------------------------------------------------
; BDOS_FileWrite: Escreve bytes num handle de arquivo (Função 49h, MSX-DOS 2)
; Entrada: B = handle, DE = buffer de origem, HL = quantidade de bytes a
;          escrever
; Saída: A = código de erro (0 = sucesso), HL = bytes realmente escritos
; -----------------------------------------------------------------------------
BDOS_FileWrite:
    LD C, BDOS_F_WRITE
    CALL BDOS_ENTRY
    RET

; -----------------------------------------------------------------------------
; BDOS_FileSeek: Reposiciona o ponteiro de um handle de arquivo (Função 4Ah,
; MSX-DOS 2)
; Entrada: B = handle, A = método (0=início, 1=posição atual, 2=fim),
;          DE:HL = deslocamento com sinal (DE = word alto, HL = word baixo)
; Saída: A = código de erro (0 = sucesso), DE:HL = nova posição absoluta
; -----------------------------------------------------------------------------
BDOS_FileSeek:
    LD C, BDOS_F_SEEK
    CALL BDOS_ENTRY
    RET

ENDMOD

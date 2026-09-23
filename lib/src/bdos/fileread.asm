; =============================================================================
; KIZUNA MSXLIB - bdos/fileread
; le de um handle (MSX-DOS 2)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_fileread
BANK 0

PUBLIC BDOS_FileRead
INCLUDE "../../inc/bdos.inc"

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

ENDMOD

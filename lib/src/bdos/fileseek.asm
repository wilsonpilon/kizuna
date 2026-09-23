; =============================================================================
; KIZUNA MSXLIB - bdos/fileseek
; reposiciona um handle (MSX-DOS 2)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_fileseek
BANK 0

PUBLIC BDOS_FileSeek
INCLUDE "../../inc/bdos.inc"

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

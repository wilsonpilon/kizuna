; =============================================================================
; KIZUNA MSXLIB - mem/copyfast
; copia de bloco direta (LDIR / LDDR), sem checar sobreposicao
; =============================================================================

MODULE mem_copyfast
BANK 0

PUBLIC MEM_CopyFast, MEM_CopyRev

; MEM_CopyFast: copia BC bytes de HL para DE do inicio para o fim (um LDIR).
; SO e seguro se as regioes nao se sobrepoem, ou se o destino esta antes da
; origem. Para o caso geral use MEM_Copy.
; Entrada: HL = origem, DE = destino, BC = quantidade (0 nao faz nada)
; Destrói: A, BC, DE, HL, flags.
MEM_CopyFast:
    LD A, B
    OR C
    RET Z
    LDIR
    RET

; MEM_CopyRev: copia BC bytes de HL para DE do FIM para o inicio (um LDDR).
; Os ponteiros sao os do INICIO dos blocos. Seguro quando o destino esta
; depois da origem, mesmo com sobreposicao.
; Entrada: HL = origem, DE = destino, BC = quantidade (0 nao faz nada)
; Destrói: A, BC, DE, HL, flags.
MEM_CopyRev:
    LD A, B
    OR C
    RET Z
    ADD HL, BC
    DEC HL
    EX DE, HL
    ADD HL, BC
    DEC HL
    EX DE, HL
    LDDR
    RET

ENDMOD

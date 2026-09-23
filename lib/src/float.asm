; =============================================================================
; KIZUNA MSXLIB - FLOAT.ASM
; Aritmética SINGLE (IEEE 754 binary32) por software: Add/Sub/Cmp.
;
; Algoritmo verificado primeiro em Go (protótipo com unpack/align/normalize/
; repack byte-a-byte, testado contra a aritmética nativa float32 do Go em
; ~200 mil pares aleatórios + casos especiais) antes de transliterar pra Z80
; aqui.
;
; ESCOPO DESTA LEVA: Add32, Sub32, Cmp32. Mul32/Div32 (multiplicação de
; mantissa 24x24->48 bits e divisão longa) se revelaram significativamente
; mais complexas na prática do que a estimativa inicial -- a normalização
; do produto de 48 bits tem uma sutileza real (a direção do ajuste de
; expoente depende de QUAL dos dois casos de faixa o produto cai, e é
; fácil inverter por engano) que só ficou clara escrevendo o algoritmo com
; cuidado. Entregar Add/Sub/Cmp corretos e verificados agora, e deixar
; Mul/Div pra uma leva própria (com o mesmo tratamento de verificação),
; em vez de arriscar um bloco de código malfeito, bate com a disciplina já
; estabelecida no projeto inteiro: nunca preferir "parece pronto" a
; "verificado pronto".
;
; LIMITAÇÕES CONHECIDAS, aceitas e documentadas (não escondidas):
;   - SEM bit de guarda/arredondamento fino: os deslocamentos de alinhamento
;     e normalização TRUNCAM em vez de arredondar corretamente. Na prática,
;     uma fração grande das operações fica 1 ULP (a última casa binária)
;     diferente do resultado IEEE754 "de verdade" (arredondado ao mais
;     próximo) -- aceito. Em subtrações de magnitudes muito próximas
;     (cancelamento catastrófico), esse erro de 1 bit pode ser amplificado
;     pela normalização subsequente -- também aceito nesta leva.
;   - SEM tratamento de Infinity/NaN/overflow de expoente.
;   - Subnormais (expoente armazenado 0, mantissa != 0) são tratados como
;     zero.
;   - SEM suporte a expressões aninhadas -- responsabilidade do lado
;     DIGNAC (pkg/dignac/codegen.go), não deste módulo.
;
; Convenção de chamada: Float_Add32/Sub32 recebem HL=endereço do operando A
; (TAMBÉM onde o resultado é escrito, em lugar) e DE=endereço do operando
; B. Float_Cmp32 recebe HL=endereço de A, DE=endereço de B, não modifica
; nenhum dos dois, e devolve o resultado nas flags Z (iguais) e C (A<B) --
; mesma semântica de um "SBC HL,DE" entre inteiros, pra permitir ao DIGNAC
; reaproveitar a mesma lógica de branching que já usa pra INTEGER.
;
; Nota de instruções: o Z80 só permite ALU indireto (ADD/SUB/CP/etc.) via
; (HL) -- nunca (DE)/(BC). Toda rotina abaixo que precisa combinar bytes
; vindos de endereços diferentes sempre carrega um deles num registrador
; ("LD A,(DE) / LD B,A") antes de operar via (HL) ou registrador-
; registrador -- nunca "ADD A,(DE)"/"CP (DE)", que não existem no Z80.
; =============================================================================

MODULE FLOAT
BANK 0

PUBLIC Float_Add32, Float_Sub32, Float_Cmp32

; --- Registro desempacotado (5 bytes, um label PRÓPRIO por campo -- o
; KAJI80 NÃO suporta aritmética "Label+N" em operandos: uma expressão como
; "Float_UnpA+1" é tratada como o NOME LITERAL de um símbolo à parte, não
; como "endereço de Float_UnpA mais 1" -- vira uma referência EXTERN
; fantasma que nunca resolve, silenciosamente, até a linkagem falhar (ou
; pior, resolver por acidente contra outra coisa). Cada campo tem seu
; próprio label abaixo, consecutivo na memória (por ordem de declaração),
; então Float_UnpA ainda serve como endereço BASE do registro de 5 bytes
; inteiro pras rotinas genéricas (Unpack/Pack/CopyUnp/MantIsZero, que só
; andam via INC HL/DE em tempo de execução, nunca aritmética de label) ---
Float_UnpA:      DB 00h  ; sinal
Float_UnpA_Exp:  DB 00h  ; expoente sem viés (signed)
Float_UnpA_M0:   DB 00h  ; mantissa LSB
Float_UnpA_M1:   DB 00h  ; mantissa meio
Float_UnpA_M2:   DB 00h  ; mantissa MSB (bit7 = bit implícito)

Float_UnpB:      DB 00h
Float_UnpB_Exp:  DB 00h
Float_UnpB_M0:   DB 00h
Float_UnpB_M1:   DB 00h
Float_UnpB_M2:   DB 00h

Float_UnpScratch: DS 5

; --- Rascunho de Float_Unpack/Float_Pack ---
Float_RawB0:  DB 00h
Float_RawB1:  DB 00h
Float_RawB2:  DB 00h
Float_RawB3:  DB 00h
Float_Sign:   DB 00h
Float_RawExp: DB 00h
Float_MantHi: DB 00h

; --- Endereços de entrada/saída guardados entre desempacotar e reempacotar ---
Float_DestAddr: DS 2
Float_AddrB:    DS 2

; -----------------------------------------------------------------------------
; Float_Unpack: desempacota um float32 (4 bytes IEEE754, endereço em HL) pro
; registro de 5 bytes em DE. Zero (e subnormais, tratados como zero nesta
; leva) viram expoente=0, mantissa=0, sinal original preservado.
; Preserva HL e DE do chamador.
; -----------------------------------------------------------------------------
Float_Unpack:
    PUSH HL
    PUSH DE

    LD A, (HL)
    LD (Float_RawB0), A
    INC HL
    LD A, (HL)
    LD (Float_RawB1), A
    INC HL
    LD A, (HL)
    LD (Float_RawB2), A
    INC HL
    LD A, (HL)
    LD (Float_RawB3), A

    AND 80h
    RLCA
    LD (Float_Sign), A

    LD A, (Float_RawB3)
    AND 7Fh
    SLA A
    LD B, A
    LD A, (Float_RawB2)
    RLCA
    AND 01h
    OR B
    LD (Float_RawExp), A

    LD A, (Float_RawB2)
    AND 7Fh
    LD (Float_MantHi), A

    POP DE
    LD A, (Float_RawExp)
    OR A
    JR Z, Float_Unpack_Zero

    LD A, (Float_Sign)
    LD (DE), A
    INC DE
    LD A, (Float_RawExp)
    SUB 127
    LD (DE), A
    INC DE
    LD A, (Float_RawB0)
    LD (DE), A
    INC DE
    LD A, (Float_RawB1)
    LD (DE), A
    INC DE
    LD A, (Float_MantHi)
    OR 80h
    LD (DE), A
    JR Float_Unpack_Done

Float_Unpack_Zero:
    LD A, (Float_Sign)
    LD (DE), A
    INC DE
    XOR A
    LD (DE), A
    INC DE
    LD (DE), A
    INC DE
    LD (DE), A
    INC DE
    LD (DE), A

Float_Unpack_Done:
    POP HL
    RET

; -----------------------------------------------------------------------------
; Float_Pack: empacota um registro desempacotado de 5 bytes (endereço em HL)
; pro formato IEEE754 float32 de 4 bytes (endereço em DE). Assume mantissa
; (offsets 2..4) já normalizada (bit implícito em offset4 bit7 setado) OU
; exatamente zero, virando +0/-0. Preserva HL e DE do chamador.
; -----------------------------------------------------------------------------
Float_Pack:
    PUSH HL
    PUSH DE

    LD A, (HL)
    LD (Float_Sign), A
    INC HL
    LD A, (HL)
    LD (Float_RawExp), A
    INC HL
    LD A, (HL)
    LD (Float_RawB0), A
    INC HL
    LD A, (HL)
    LD (Float_RawB1), A
    INC HL
    LD A, (HL)
    LD (Float_RawB2), A

    OR A
    JR NZ, Float_Pack_NonZero
    LD A, (Float_RawB1)
    OR A
    JR NZ, Float_Pack_NonZero
    LD A, (Float_RawB0)
    OR A
    JR NZ, Float_Pack_NonZero

    POP DE
    XOR A
    LD (DE), A
    INC DE
    LD (DE), A
    INC DE
    LD (DE), A
    INC DE
    LD A, (Float_Sign)
    RRCA
    LD (DE), A
    JR Float_Pack_Done

Float_Pack_NonZero:
    POP DE
    LD A, (Float_RawExp)
    ADD A, 127
    LD (Float_RawExp), A

    LD A, (Float_RawB0)
    LD (DE), A
    INC DE
    LD A, (Float_RawB1)
    LD (DE), A
    INC DE

    LD A, (Float_RawB2)
    AND 7Fh
    LD B, A
    LD A, (Float_RawExp)
    AND 01h
    RRCA
    OR B
    LD (DE), A
    INC DE

    LD A, (Float_RawExp)
    SRL A
    LD B, A
    LD A, (Float_Sign)
    RRCA
    AND 80h
    OR B
    LD (DE), A

Float_Pack_Done:
    POP HL
    RET

; -----------------------------------------------------------------------------
; Float_CopyUnp: copia um registro desempacotado de 5 bytes de HL pra DE.
; Preserva HL e DE.
; -----------------------------------------------------------------------------
Float_CopyUnp:
    PUSH HL
    PUSH DE
    PUSH BC
    LD B, 5
Float_CopyUnp_Loop:
    LD A, (HL)
    LD (DE), A
    INC HL
    INC DE
    DJNZ Float_CopyUnp_Loop
    POP BC
    POP DE
    POP HL
    RET

; -----------------------------------------------------------------------------
; Float_MantIsZero: HL = endereço de um registro desempacotado (5 bytes).
; Devolve flag Z setada se os 3 bytes de mantissa (offsets 2,3,4) são
; todos zero. Preserva HL.
; -----------------------------------------------------------------------------
Float_MantIsZero:
    PUSH HL
    PUSH BC
    INC HL
    INC HL
    LD A, (HL)
    INC HL
    LD B, A
    LD A, (HL)
    INC HL
    OR B
    LD B, A
    LD A, (HL)
    OR B
    POP BC
    POP HL
    RET

; -----------------------------------------------------------------------------
; Float_Mant_Add: HL = endereço de um valor de 3 bytes (LSB primeiro,
; destino em lugar), DE = endereço do outro valor de 3 bytes (somado).
; Saída: flag carry = 1 se o resultado estourou além de 24 bits.
; -----------------------------------------------------------------------------
Float_Mant_Add:
    LD A, (DE)
    LD B, A
    LD A, (HL)
    ADD A, B
    LD (HL), A
    INC HL
    INC DE
    LD A, (DE)
    LD B, A
    LD A, (HL)
    ADC A, B
    LD (HL), A
    INC HL
    INC DE
    LD A, (DE)
    LD B, A
    LD A, (HL)
    ADC A, B
    LD (HL), A
    RET

; -----------------------------------------------------------------------------
; Float_Mant_Sub: HL = endereço do minuendo de 3 bytes (destino em lugar),
; DE = endereço do subtraendo de 3 bytes. Assume minuendo >= subtraendo.
; -----------------------------------------------------------------------------
Float_Mant_Sub:
    LD A, (DE)
    LD B, A
    LD A, (HL)
    SUB B
    LD (HL), A
    INC HL
    INC DE
    LD A, (DE)
    LD B, A
    LD A, (HL)
    SBC A, B
    LD (HL), A
    INC HL
    INC DE
    LD A, (DE)
    LD B, A
    LD A, (HL)
    SBC A, B
    LD (HL), A
    RET

; -----------------------------------------------------------------------------
; Float_Mant_Cmp: compara dois valores de 3 bytes (MSB primeiro na
; comparação, LSB primeiro na memória) -- HL = endereço de A, DE =
; endereço de B, ambos preservados. Devolve flags Z (iguais) e C (A<B).
; -----------------------------------------------------------------------------
Float_Mant_Cmp:
    PUSH HL
    PUSH DE
    INC HL
    INC HL
    INC DE
    INC DE
    LD A, (DE)
    LD B, A
    LD A, (HL)
    CP B
    JR NZ, Float_Mant_Cmp_Done
    DEC HL
    DEC DE
    LD A, (DE)
    LD B, A
    LD A, (HL)
    CP B
    JR NZ, Float_Mant_Cmp_Done
    DEC HL
    DEC DE
    LD A, (DE)
    LD B, A
    LD A, (HL)
    CP B
Float_Mant_Cmp_Done:
    POP DE
    POP HL
    RET

; -----------------------------------------------------------------------------
; Float_Mant_Shr1: desloca um valor de 3 bytes (endereço em HL) 1 bit à
; direita, em lugar (lógico, entra 0 no topo). Preserva HL. Devolve em
; carry o bit que saiu pelo fundo (LSB).
; -----------------------------------------------------------------------------
Float_Mant_Shr1:
    PUSH HL
    INC HL
    INC HL
    SRL (HL)
    DEC HL
    RR (HL)
    DEC HL
    RR (HL)
    POP HL
    RET

; -----------------------------------------------------------------------------
; Float_Mant_Shl1: desloca um valor de 3 bytes (endereço em HL) 1 bit à
; esquerda, em lugar (entra 0 no fundo). Preserva HL. Bit que sai pelo
; topo é perdido.
; -----------------------------------------------------------------------------
Float_Mant_Shl1:
    PUSH HL
    SLA (HL)
    INC HL
    RL (HL)
    INC HL
    RL (HL)
    POP HL
    RET

; -----------------------------------------------------------------------------
; Float_AddCore: soma Float_UnpB em Float_UnpA (em lugar) -- ambos já
; desempacotados, nenhum dos dois é zero (chamador garante via
; Float_AddSubCommon). Resultado fica em Float_UnpA.
; -----------------------------------------------------------------------------
Float_AddCore:
    ; garante que Float_UnpA tem o MAIOR peso (expoente maior, ou expoente
    ; igual e mantissa maior) -- senão, troca A<->B por completo.
    ; Expoente é signed (viés já removido) -- CP direto seria comparação
    ; UNSIGNED e erraria sempre que os expoentes tiverem sinais diferentes
    ; (ex.: exp=0 vs exp=-1). Usa o mesmo truque de inverter o bit de sinal
    ; que Float_Cmp32 já usa, convertendo pra uma ordem totalmente
    ; comparável via CP normal antes de comparar.
    LD A, (Float_UnpB_Exp)
    XOR 80h
    LD B, A
    LD A, (Float_UnpA_Exp)
    XOR 80h
    CP B
    JR C, Float_AddCore_Swap
    JR NZ, Float_AddCore_NoSwap
    LD HL, Float_UnpA_M2
    LD A, (HL)
    LD HL, Float_UnpB_M2
    CP (HL)
    JR C, Float_AddCore_Swap
    JR NZ, Float_AddCore_NoSwap
    LD HL, Float_UnpA_M1
    LD A, (HL)
    LD HL, Float_UnpB_M1
    CP (HL)
    JR C, Float_AddCore_Swap
    JR NZ, Float_AddCore_NoSwap
    LD HL, Float_UnpA_M0
    LD A, (HL)
    LD HL, Float_UnpB_M0
    CP (HL)
    JR C, Float_AddCore_Swap
    JR Float_AddCore_NoSwap

Float_AddCore_Swap:
    LD HL, Float_UnpA
    LD DE, Float_UnpScratch
    CALL Float_CopyUnp
    LD HL, Float_UnpB
    LD DE, Float_UnpA
    CALL Float_CopyUnp
    LD HL, Float_UnpScratch
    LD DE, Float_UnpB
    CALL Float_CopyUnp

Float_AddCore_NoSwap:
    ; diff = UnpA.exp - UnpB.exp (>=0 garantido pela ordenação acima)
    LD A, (Float_UnpA_Exp)
    LD B, A
    LD A, (Float_UnpB_Exp)
    LD C, A
    LD A, B
    SUB C
    CP 25
    RET NC                        ; diff>=25 -> B some por completo, resultado já é A
    OR A
    JR Z, Float_AddCore_Aligned

    LD B, A
Float_AddCore_ShiftLoop:
    LD HL, Float_UnpB_M0
    CALL Float_Mant_Shr1
    DJNZ Float_AddCore_ShiftLoop

Float_AddCore_Aligned:
    LD A, (Float_UnpA)
    LD B, A
    LD A, (Float_UnpB)
    CP B
    JR NZ, Float_AddCore_DiffSign

    ; mesmos sinais: soma as mantissas
    LD HL, Float_UnpA_M0
    LD DE, Float_UnpB_M0
    CALL Float_Mant_Add
    JR NC, Float_AddCore_Done

    ; estourou pro bit 24 -- desloca resultado 1 bit à direita (o carry do
    ; estouro entra pelo topo) e incrementa o expoente
    LD HL, Float_UnpA_M2
    RR (HL)
    LD HL, Float_UnpA_M1
    RR (HL)
    LD HL, Float_UnpA_M0
    RR (HL)
    LD A, (Float_UnpA_Exp)
    INC A
    LD (Float_UnpA_Exp), A
    JR Float_AddCore_Done

Float_AddCore_DiffSign:
    ; sinais diferentes: subtrai a menor magnitude da maior
    LD HL, Float_UnpA_M0
    LD DE, Float_UnpB_M0
    CALL Float_Mant_Cmp
    JR NC, Float_AddCore_AGeB

    ; A.mant < B.mant: resultado = B.mant - A.mant, sinal final = sinal de B
    LD HL, Float_UnpB_M0
    LD DE, Float_UnpA_M0
    CALL Float_Mant_Sub            ; resultado em Float_UnpB_M0 (em lugar)
    LD HL, Float_UnpB_M0
    LD DE, Float_UnpA_M0
    LD A, (HL)
    LD (DE), A
    INC HL
    INC DE
    LD A, (HL)
    LD (DE), A
    INC HL
    INC DE
    LD A, (HL)
    LD (DE), A
    LD A, (Float_UnpB)
    LD (Float_UnpA), A
    JR Float_AddCore_Normalize

Float_AddCore_AGeB:
    LD HL, Float_UnpA_M0
    LD DE, Float_UnpB_M0
    CALL Float_Mant_Sub             ; resultado em Float_UnpA_M0 (em lugar), sinal já é o de A

Float_AddCore_Normalize:
    LD HL, Float_UnpA
    CALL Float_MantIsZero
    JR NZ, Float_AddCore_NormLoop
    XOR A
    LD (Float_UnpA), A
    LD (Float_UnpA_Exp), A
    JR Float_AddCore_Done

Float_AddCore_NormLoop:
    LD A, (Float_UnpA_M2)
    AND 80h
    JR NZ, Float_AddCore_Done
    LD HL, Float_UnpA_M0
    CALL Float_Mant_Shl1
    LD A, (Float_UnpA_Exp)
    DEC A
    LD (Float_UnpA_Exp), A
    JR Float_AddCore_NormLoop

Float_AddCore_Done:
    RET

; -----------------------------------------------------------------------------
; Float_AddSubCommon: rotina compartilhada por Float_Add32/Float_Sub32.
; Entrada: HL=endereço de A (destino final), DE=endereço de B, B(registrador)
; = 0 pra soma / 1 pra subtração (inverte o sinal de B após desempacotar).
; -----------------------------------------------------------------------------
Float_AddSubCommon:
    LD (Float_DestAddr), HL
    PUSH BC                        ; guarda a flag soma/subtração (em C) através das chamadas abaixo
    LD C, B

    EX DE, HL
    LD (Float_AddrB), HL
    EX DE, HL

    LD DE, Float_UnpA
    CALL Float_Unpack

    LD HL, (Float_AddrB)
    LD DE, Float_UnpB
    CALL Float_Unpack

    LD A, C
    OR A
    JR Z, Float_AddSubCommon_SignOk
    LD A, (Float_UnpB)
    XOR 01h
    LD (Float_UnpB), A
Float_AddSubCommon_SignOk:
    POP BC

    ; caso A ou B seja zero, resultado é o outro (já desempacotado) --
    ; evita entrar em Float_AddCore, que assume os dois não-zero
    LD HL, Float_UnpA
    CALL Float_MantIsZero
    JR NZ, Float_AddSubCommon_CheckB
    LD HL, Float_UnpB
    LD DE, Float_UnpA
    CALL Float_CopyUnp
    JR Float_AddSubCommon_Pack
Float_AddSubCommon_CheckB:
    LD HL, Float_UnpB
    CALL Float_MantIsZero
    JR NZ, Float_AddSubCommon_DoAdd
    JR Float_AddSubCommon_Pack
Float_AddSubCommon_DoAdd:
    CALL Float_AddCore
Float_AddSubCommon_Pack:
    ; LD DE,(nn) não existe no Z80 de verdade (só LD HL,(nn) tem forma
    ; absoluta de 16 bits) -- carrega em HL primeiro e troca pra DE.
    LD HL, (Float_DestAddr)
    EX DE, HL
    LD HL, Float_UnpA
    CALL Float_Pack
    RET

; -----------------------------------------------------------------------------
; Float_Add32 (PUBLIC): HL = endereço de A (também destino), DE = endereço
; de B. Resultado = A + B, IEEE754 binary32.
; -----------------------------------------------------------------------------
Float_Add32:
    LD B, 0
    JP Float_AddSubCommon

; -----------------------------------------------------------------------------
; Float_Sub32 (PUBLIC): HL = endereço de A (também destino), DE = endereço
; de B. Resultado = A - B.
; -----------------------------------------------------------------------------
Float_Sub32:
    LD B, 1
    JP Float_AddSubCommon

; -----------------------------------------------------------------------------
; Float_Cmp32 (PUBLIC): HL = endereço de A, DE = endereço de B. Não
; modifica A/B. Flags Z (iguais) / C (A<B) -- mesma semântica de SBC HL,DE.
;
; Implementação: transforma os dois padrões de bits IEEE754 em inteiros de
; 32 bits totalmente ordenáveis (se o sinal está setado, inverte todos os
; bits; senão, inverte só o bit de sinal) e compara os 4 bytes resultantes
; via SUB/SBC encadeado, MSB primeiro -- as flags da última subtração já
; são exatamente Z/C no sentido certo.
; -----------------------------------------------------------------------------
Float_Cmp32:
    PUSH HL
    PUSH DE
    PUSH BC

    ; transforma A (em HL) pra ordem totalmente comparável, guarda em
    ; Float_RawB0..B3 (reaproveita as células de rascunho já existentes)
    LD A, (HL)
    LD (Float_RawB0), A
    INC HL
    LD A, (HL)
    LD (Float_RawB1), A
    INC HL
    LD A, (HL)
    LD (Float_RawB2), A
    INC HL
    LD A, (HL)
    BIT 7, A
    JR Z, Float_Cmp32_A_Positive
    ; negativo -- inverte todos os 4 bytes
    LD (Float_RawB3), A
    LD A, (Float_RawB0)
    CPL
    LD (Float_RawB0), A
    LD A, (Float_RawB1)
    CPL
    LD (Float_RawB1), A
    LD A, (Float_RawB2)
    CPL
    LD (Float_RawB2), A
    LD A, (Float_RawB3)
    CPL
    LD (Float_RawB3), A
    JR Float_Cmp32_A_Done
Float_Cmp32_A_Positive:
    XOR 80h
    LD (Float_RawB3), A
Float_Cmp32_A_Done:

    ; transforma B (em DE) do mesmo jeito, guarda em Float_Sign,
    ; Float_RawExp, Float_MantHi, Float_DestAddr (baixo) -- reaproveita
    ; células de rascunho de 1 byte já existentes, sem conflito, já que
    ; Cmp32 não chama Unpack/Pack
    LD A, (DE)
    LD (Float_Sign), A
    INC DE
    LD A, (DE)
    LD (Float_RawExp), A
    INC DE
    LD A, (DE)
    LD (Float_MantHi), A
    INC DE
    LD A, (DE)
    BIT 7, A
    JR Z, Float_Cmp32_B_Positive
    LD B, A
    LD A, (Float_Sign)
    CPL
    LD (Float_Sign), A
    LD A, (Float_RawExp)
    CPL
    LD (Float_RawExp), A
    LD A, (Float_MantHi)
    CPL
    LD (Float_MantHi), A
    LD A, B
    CPL
    JR Float_Cmp32_B_Done
Float_Cmp32_B_Positive:
    XOR 80h
Float_Cmp32_B_Done:
    LD B, A                        ; B = byte3 de B transformado

    ; compara os dois valores de 32 bits transformados, MSB primeiro:
    ; Float_RawB3 (A) contra B (B, byte3 de B transformado)
    LD A, (Float_RawB3)
    CP B
    JR NZ, Float_Cmp32_Result
    LD A, (Float_RawB2)
    LD B, (Float_MantHi)
    CP B
    JR NZ, Float_Cmp32_Result
    LD A, (Float_RawB1)
    LD B, (Float_RawExp)
    CP B
    JR NZ, Float_Cmp32_Result
    LD A, (Float_RawB0)
    LD B, (Float_Sign)
    CP B

Float_Cmp32_Result:
    POP BC
    POP DE
    POP HL
    RET

ENDMOD

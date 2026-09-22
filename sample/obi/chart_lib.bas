' ============================================================
' KIZUNA sample -- OBI: modulo BASIC "biblioteca" (sem PROCEDURE Main,
' logo sem Start proprio -- nenhum conflito com o Start do KAJI80).
' Compilador: DIGNAC
' Chamado a partir de main.asm (KAJI80, banco 0) via MUSUBI.
'
' BANK 0 (area comum) de proposito: um teste real em hardware mostrou que
' o bootstrap multi-banco do MUSUBI (troca de pagina via ALL_SEG/EXTBIO,
' adicionado na v4.5.2) nao roda corretamente -- so tinha sido validado
' por analise estatica ate entao. Registrado como bug separado do MUSUBI
' (ver memoria do projeto); este exemplo evita esse caminho de proposito
' para continuar mostrando KAJI80+DIGNAC+resource+biblioteca orquestrados
' pelo OBI, sem depender de um bootstrap ainda nao confirmado em execucao.
' ============================================================

MODULE ChartLib
BANK 0
PUBLIC Desenhar

' Desenhar: recebe um valor pela convencao de pilha do KIZUNA (o chamador
' empilha o argumento antes do CALL) e traca uma moldura + curva de pontos
' em SCREEN 2. Assume que o chamador ja ajustou o modo de video.
PROCEDURE Desenhar(valor%)
    LOCAL x%, y%

    LINE (0,0)-(255,191), 1, BF
    LINE (8,8)-(247,183), 15

    FOR x% = 9 TO 246
        y% = 175 - (x% MOD (valor% + 1)) * 7
        PSET (x%, y%), 10
    NEXT x%
END PROCEDURE
END MODULE

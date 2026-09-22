' ============================================================
' KIZUNA sample -- OBI: modulo BASIC "biblioteca" (sem PROCEDURE Main,
' logo sem Start proprio -- nenhum conflito com o Start do KAJI80).
' Compilador: DIGNAC
' Chamado a partir de main.asm (KAJI80, banco 0) via MUSUBI, que gera o
' trampolim de troca de banco automaticamente.
' ============================================================

MODULE ChartLib
BANK 2
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

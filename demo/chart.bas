' ============================================================
' KIZUNA demo -- croqui hipotetico (nao compila / prova de conceito)
' Compilador: DIGNAC
' Banco: 2 (paginavel)
' Tamanho estimado: ~5Kb
' ============================================================

MODULE Chart
    PUBLIC Desenhar

    ' Desenhar: recebe um valor (via ABI de pilha do KIZUNA) e traca
    ' um grafico simples de barras/linha na tela ja ajustada pelo
    ' modulo ScreenFX (Assembly).
    PROCEDURE Desenhar(valor%)
        LOCAL x%, y%

        LINE (0,0)-(255,191), 1, BF

        FOR x% = 0 TO 255
            y% = 191 - (x% MOD (valor% + 1)) * 2
            PSET (x%, y%), 15
        NEXT x%

    END PROCEDURE
END MODULE

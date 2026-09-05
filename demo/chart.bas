' ============================================================
' KIZUNA sample -- Modulo Chart em MSX-BASIC Dignified
' Compilador: DIGNAC
' Banco: 2 (paginavel)
' ============================================================

MODULE Chart
BANK 0
PUBLIC Main, Desenhar
EXTERN BIOS_CHGET

' Ponto de entrada para demonstracao grafica standalone
PROCEDURE Main()
    SCREEN 2
    Desenhar(10)
    BIOS_CHGET()
    SCREEN 0
END PROCEDURE

' Desenhar: recebe um valor (via ABI de pilha do KIZUNA) e traca
' um grafico simples de barras/linha na tela ajustada em SCREEN 2.
PROCEDURE Desenhar(valor%)
        LOCAL x%, y%

        LINE (0,0)-(255,191), 1, BF

        FOR x% = 0 TO 255
            y% = 191 - (x% MOD (valor% + 1)) * 2
            PSET (x%, y%), 15
        NEXT x%

    END PROCEDURE
END MODULE

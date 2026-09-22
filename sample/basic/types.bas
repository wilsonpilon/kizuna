' ============================================================
' KIZUNA sample -- Sistema de tipos do DIGNAC: STRING de verdade
' (declarar/atribuir/imprimir), literais &O (octal) e &B (binario),
' e declaracao + atribuicao de SINGLE/DOUBLE (aritmetica de ponto
' flutuante ainda nao implementada -- fica pra uma leva futura).
' Compilador: DIGNAC
' ============================================================

MODULE Types
BANK 0
PUBLIC Main

DIM nome$, saudacao$
DIM idade%, permissoes%
DIM altura!, pi#

PROCEDURE Main()
    nome$ = "Kizuna"
    saudacao$ = "Ola, mundo!"
    idade% = 42
    permissoes% = &O17
    altura! = 1.75
    pi# = 3.14159265358979d0

    PRINT saudacao$
    PRINT nome$
    PRINT idade%
    PRINT permissoes%
END PROCEDURE
END MODULE

' ============================================================
' KIZUNA sample -- Motor de aritmetica SINGLE (Float_Add32/
' Float_Sub32/Float_Cmp32, lib/src/float.asm) exercitado pelo
' DIGNAC: "x! = a! + b!", "x! = a! - b!" e "IF a! > b! THEN".
' Escopo desta leva: so +, - e comparacao com operandos simples
' (sem aninhar, sem *,  /, sem DOUBLE, sem PRINT de float --
' fica tudo documentado em docs/manual-basic-dignified.md).
' Verificacao sem PRINT de float: cada resultado e comparado
' contra o valor esperado via IF, e o resultado (OK/FALHOU) vira
' texto visivel na tela -- efeito colateral observavel em hardware
' de verdade, ja que nao da pra imprimir o float em si ainda.
' Compilador: DIGNAC
' ============================================================

MODULE FloatMath
BANK 0
PUBLIC Main

DIM a!, b!, soma!, diferenca!

PROCEDURE Main()
    a! = 1.0
    b! = 0.5

    soma! = a! + b!
    IF soma! = 1.5 THEN
        PRINT "SOMA OK: 1.0+0.5=1.5"
    ELSE
        PRINT "SOMA FALHOU"
    END IF

    diferenca! = a! - b!
    IF diferenca! = 0.5 THEN
        PRINT "SUBTRACAO OK: 1.0-0.5=0.5"
    ELSE
        PRINT "SUBTRACAO FALHOU"
    END IF

    IF a! > b! THEN
        PRINT "COMPARACAO OK: 1.0 > 0.5"
    ELSE
        PRINT "COMPARACAO FALHOU"
    END IF

    IF b! < a! THEN
        PRINT "COMPARACAO OK: 0.5 < 1.0"
    ELSE
        PRINT "COMPARACAO FALHOU"
    END IF
END PROCEDURE
END MODULE

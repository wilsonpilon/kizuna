MODULE Calc
BANK 0
PUBLIC Main

PROCEDURE Main()
    LOCAL a%, b%, res%
    a% = 123
    b% = 45
    res% = a% * b%

    PRINT "========================================"
    PRINT " DIGNAC Math & Logic Demo (16-bit)      "
    PRINT "========================================"
    PRINT "a% = "; a%
    PRINT "b% = "; b%
    PRINT "Multiplicacao (a% * b%): "; res%
    PRINT "Divisao inteira (res% / 10): "; res% / 10
    PRINT "Resto / Modulo (res% MOD 10): "; res% MOD 10
END PROCEDURE
END MODULE

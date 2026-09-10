' ============================================================
' KIZUNA sample -- Teste de DUAS chamadas de PSET (nao LINE, nao loop)
' Duas instrucoes PSET separadas, geradas pelo DIGNAC como duas
' chamadas independentes a VDP_PSet -- nao passa pelo laco apertado
' do VDP_Line. Serve para isolar se o problema e "qualquer segunda
' chamada" ou especifico do loop interno do VDP_Line.
' ============================================================

MODULE PSetTwo
BANK 0
PUBLIC Main
EXTERN BIOS_CHGET

PROCEDURE Main()
    SCREEN 2
    PSET (8, 8), 15
    PSET (16, 8), 15
    BIOS_CHGET()
    SCREEN 0
END PROCEDURE
END MODULE

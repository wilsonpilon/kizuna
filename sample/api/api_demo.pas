{ Chama rotinas da MSXLIB direto do Pascal, pelos descritores lib/api/*.api: }
{ nenhum comando dedicado no compilador, nenhum EXTERN.                       }
program ApiDemoPas;

var
    k: Integer;

begin
    VDP_SetColor(15, 4);                { texto branco, fundo azul }
    BIOS_CLS();
    WriteLn('KIZUNA - MSXLIB chamada pelo Pascal via .api');
    WriteLn('--------------------------------------------');

    Write('6 * 7 = ');
    PrintDec16(Mul16(6, 7));            { 42 }
    WriteLn;

    Write('100 / 7 = ');
    PrintDec16(Div16(100, 7));          { 14 }
    WriteLn;

    Write('255 em hexa: 0x');
    PrintHex8(255);                     { FF }
    WriteLn;

    WriteLn('Tom de 440 Hz no canal A. Aperte uma tecla para parar.');
    PSG_PlayTone(0, 254, 12);           { periodo 254 = La 4 }
    k := BDOS_ReadChar();
    PSG_MuteAll();

    VDP_SetColor(15, 1);                { devolve o fundo preto }
    WriteLn('Fim.');
end.

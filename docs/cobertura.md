# Cobertura: MSXgl e Fusion-C → MSXLIB

> Arquivo **gerado** por `go run ./tools/cobertura` a partir dos cabeçalhos em `resource/` e do mapa
> `docs/cobertura.map`. Não edite à mão: edite o mapa e gere de novo.
>
> O MSXgl e o Fusion-C são apenas **guias do que cobrir** (ver `docs/plano-expansao-msxlib.md`):
> aqui só constam os nomes das funções deles; nenhuma implementação é copiada para a MSXLIB.

## Resumo

| Área | Itens | Feito | Planejado | Não faremos | Pendente |
| --- | ---: | ---: | ---: | ---: | ---: |
| Memória | 28 | 18 | 0 | 10 | 0 |
| Matemática | 29 | 27 | 0 | 0 | 2 |
| Strings e texto | 118 | 46 | 2 | 0 | 70 |
| VDP | 174 | 22 | 0 | 0 | 152 |
| Draw | 41 | 0 | 0 | 0 | 41 |
| Tile | 14 | 0 | 0 | 0 | 14 |
| Scroll | 7 | 0 | 0 | 0 | 7 |
| Teclado | 24 | 1 | 0 | 0 | 23 |
| Joystick e mouse | 17 | 0 | 0 | 0 | 17 |
| PSG | 28 | 5 | 0 | 0 | 23 |
| Play (players) | 172 | 0 | 0 | 0 | 172 |
| MSX-Music | 7 | 0 | 0 | 0 | 7 |
| MSX-Audio | 6 | 0 | 0 | 0 | 6 |
| SCC | 11 | 0 | 0 | 0 | 11 |
| BIOS | 95 | 0 | 0 | 0 | 95 |
| DOS | 105 | 0 | 0 | 0 | 105 |
| System | 87 | 2 | 0 | 0 | 85 |
| Clock | 30 | 0 | 0 | 0 | 30 |
| V9990 | 185 | 0 | 0 | 0 | 185 |
| Fora da lista (por enquanto) | 323 | 0 | 0 | 0 | 323 |
| **Total** | **1501** | **121** | **2** | **10** | **1368** |

## Memória

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `MMalloc` | Fusion-C | msx_fusion.h | feito | MEM_Alloc |
| `MemChr` | Fusion-C | msx_fusion.h | feito | MEM_Find |
| `MemCompare` | Fusion-C | msx_fusion.h | feito | MEM_Compare |
| `MemCopy` | Fusion-C | msx_fusion.h | feito | MEM_Copy |
| `MemCopyReverse` | Fusion-C | msx_fusion.h | feito | MEM_CopyRev |
| `MemFill` | Fusion-C | msx_fusion.h | feito | MEM_Fill |
| `Mem_Copy` | MSXgl | memory.h | feito | MEM_Copy |
| `Mem_Copy_16b` | MSXgl | memory.h | feito | MEM_CopyWords |
| `Mem_DynamicAlloc` | MSXgl | memory.h | feito | MEM_Alloc |
| `Mem_DynamicFree` | MSXgl | memory.h | feito | MEM_Free |
| `Mem_DynamicInitialize` | MSXgl | memory.h | feito | MEM_HeapInit |
| `Mem_DynamicInitializeHeap` | MSXgl | memory.h | não faremos | depende do alocador estático; use MEM_HeapInit/MEM_HeapInitToStack |
| `Mem_FastCopy` | MSXgl | memory.h | feito | MEM_CopyFast |
| `Mem_FastCopy_16b` | MSXgl | memory.h | feito | MEM_CopyFastWords |
| `Mem_FastSet` | MSXgl | memory.h | feito | MEM_Fill |
| `Mem_GetDynamicSize` | MSXgl | memory.h | feito | MEM_BlockSize |
| `Mem_GetHeapAddress` | MSXgl | memory.h | não faremos | estado do alocador estático do guia; não existe no nosso modelo (ver MEM_HeapSize/MEM_HeapFree) |
| `Mem_GetHeapSize` | MSXgl | memory.h | não faremos | idem: o guia mede "SP menos endereço do heap"; use MEM_HeapFree para o espaço livre do heap dinâmico |
| `Mem_GetStackAddress` | MSXgl | memory.h | feito | MEM_GetSP |
| `Mem_HeapAlloc` | MSXgl | memory.h | não faremos | alocador "de ponteiro que só avança" do guia; o heap dinâmico (MEM_Alloc/MEM_Free) cobre o uso e a lista de blocos é mais geral |
| `Mem_HeapFree` | MSXgl | memory.h | não faremos | idem Mem_HeapAlloc |
| `Mem_Set` | MSXgl | memory.h | feito | MEM_Fill |
| `Mem_Set_16b` | MSXgl | memory.h | feito | MEM_Fill16 |
| `Mutex_Gate` | MSXgl | mutex.h | não faremos | idem Mutex_Init |
| `Mutex_Init` | MSXgl | mutex.h | não faremos | MSX-DOS é monotarefa; o único paralelismo é o tratador de interrupção, tratado com DI/EI. Rever se surgir um uso real |
| `Mutex_Lock` | MSXgl | mutex.h | não faremos | idem Mutex_Init |
| `Mutex_Release` | MSXgl | mutex.h | não faremos | idem Mutex_Init |
| `Mutex_Wait` | MSXgl | mutex.h | não faremos | idem Mutex_Init |

## Matemática

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `Math_Abs` | MSXgl | math.h | feito | MATH_Abs8 |
| `Math_Abs_16b` | MSXgl | math.h | feito | MATH_Abs16 |
| `Math_Abs_32b` | MSXgl | math.h | feito | MATH_Abs32 |
| `Math_Div10` | MSXgl | math.h | feito | MATH_DivS10 |
| `Math_Div10_16b` | MSXgl | math.h | feito | MATH_DivS10 |
| `Math_Flip` | MSXgl | math.h | feito | MATH_Flip8 |
| `Math_Flip_16b` | MSXgl | math.h | feito | MATH_Flip16 |
| `Math_GetRandom16` | MSXgl | math.h | feito | MATH_Rand16 |
| `Math_GetRandom8` | MSXgl | math.h | feito | MATH_Rand8 |
| `Math_GetRandomMax16` | MSXgl | math.h | feito | MATH_RandRange16 |
| `Math_GetRandomMax8` | MSXgl | math.h | feito | MATH_RandRange8 |
| `Math_GetRandomRange16` | MSXgl | math.h | feito | MATH_RandBetween16 |
| `Math_GetRandomRange8` | MSXgl | math.h | feito | MATH_RandBetween8 |
| `Math_Mod10` | MSXgl | math.h | feito | MATH_Mod10 |
| `Math_Mod10_16b` | MSXgl | math.h | feito | MATH_Mod10 |
| `Math_Negative` | MSXgl | math.h | feito | MATH_Neg8 |
| `Math_Negative16` | MSXgl | math.h | feito | MATH_Neg16 |
| `Math_SetRandomSeed16` | MSXgl | math.h | feito | MATH_RandSeed |
| `Math_SetRandomSeed8` | MSXgl | math.h | feito | MATH_RandSeed |
| `Math_SignedDiv16` | MSXgl | math.h | feito | MATH_Sar8 |
| `Math_SignedDiv2` | MSXgl | math.h | feito | MATH_Sar8 |
| `Math_SignedDiv32` | MSXgl | math.h | feito | MATH_Sar8 |
| `Math_SignedDiv4` | MSXgl | math.h | feito | MATH_Sar8 |
| `Math_SignedDiv8` | MSXgl | math.h | feito | MATH_Sar8 |
| `Math_Swap` | MSXgl | math.h | feito | MATH_Swap16 |
| `QMN_Get16` | MSXgl | fixed_point.h | pendente |  |
| `QMN_Get8` | MSXgl | fixed_point.h | pendente |  |
| `QMN_Set16` | MSXgl | fixed_point.h | feito | MATH_Shl16 |
| `QMN_Set8` | MSXgl | fixed_point.h | feito | MATH_Shl8 |

## Strings e texto

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `Beep` | Fusion-C | msx_fusion.h | feito | BIOS_BEEP |
| `CharToLower` | Fusion-C | msx_fusion.h | feito | CHAR_ToLower |
| `CharToUpper` | Fusion-C | msx_fusion.h | feito | CHAR_ToUpper |
| `Cls` | Fusion-C | msx_fusion.h | feito | BIOS_CLS |
| `Getche` | Fusion-C | msx_fusion.h | feito | BDOS_ReadChar |
| `InputChar` | Fusion-C | msx_fusion.h | pendente |  |
| `InputString` | Fusion-C | msx_fusion.h | feito | CON_ReadLine |
| `IntSwap` | Fusion-C | msx_fusion.h | pendente |  |
| `IntToFloat` | Fusion-C | msx_fusion.h | pendente |  |
| `IsAlpha` | Fusion-C | msx_fusion.h | feito | CHAR_IsAlpha |
| `IsAlphaNum` | Fusion-C | msx_fusion.h | feito | CHAR_IsAlNum |
| `IsAscii` | Fusion-C | msx_fusion.h | pendente |  |
| `IsCntrl` | Fusion-C | msx_fusion.h | feito | CHAR_IsControl |
| `IsDigit` | Fusion-C | msx_fusion.h | feito | CHAR_IsDigit |
| `IsGraph` | Fusion-C | msx_fusion.h | feito | CHAR_IsGraph |
| `IsHexDigit` | Fusion-C | msx_fusion.h | feito | CHAR_IsHexDigit |
| `IsLower` | Fusion-C | msx_fusion.h | feito | CHAR_IsLower |
| `IsPositive` | Fusion-C | msx_fusion.h | pendente |  |
| `IsPrintable` | Fusion-C | msx_fusion.h | feito | CHAR_IsPrint |
| `IsPunctuation` | Fusion-C | msx_fusion.h | feito | CHAR_IsPunct |
| `IsSpace` | Fusion-C | msx_fusion.h | feito | CHAR_IsSpace |
| `IsUpper` | Fusion-C | msx_fusion.h | feito | CHAR_IsUpper |
| `Itoa` | Fusion-C | msx_fusion.h | pendente |  |
| `Locate` | Fusion-C | msx_fusion.h | feito | CON_Locate |
| `NStrCompare` | Fusion-C | msx_fusion.h | feito | CSTR_CompareN |
| `NStrConcat` | Fusion-C | msx_fusion.h | feito | CSTR_CatN |
| `NStrCopy` | Fusion-C | msx_fusion.h | feito | CSTR_CopyN |
| `Print` | Fusion-C | msx_fusion.h | feito | CON_PrintCStr |
| `PrintChar` | Fusion-C | msx_fusion.h | feito | BDOS_PrintChar |
| `PrintDec` | Fusion-C | msx_fusion.h | feito | CON_PrintI16 |
| `PrintFNumber` | Fusion-C | msx_fusion.h | pendente |  |
| `PrintHex` | Fusion-C | msx_fusion.h | feito | PrintHex16 |
| `PrintNumber` | Fusion-C | msx_fusion.h | feito | PrintDec16 |
| `PrintString` | Fusion-C | msx_fusion.h | feito | CON_PrintCStr |
| `PutCharHex` | Fusion-C | msx_fusion.h | feito | PrintHex8 |
| `StrChr` | Fusion-C | msx_fusion.h | feito | CSTR_FindChar |
| `StrCompare` | Fusion-C | msx_fusion.h | feito | CSTR_Compare |
| `StrConcat` | Fusion-C | msx_fusion.h | feito | CSTR_Cat |
| `StrCopy` | Fusion-C | msx_fusion.h | feito | CSTR_Copy |
| `StrLeftTrim` | Fusion-C | msx_fusion.h | feito | CSTR_TrimLeft |
| `StrLen` | Fusion-C | msx_fusion.h | feito | CSTR_Len |
| `StrPosChr` | Fusion-C | msx_fusion.h | feito | CSTR_FindLastChar |
| `StrPosStr` | Fusion-C | msx_fusion.h | feito | CSTR_FindStr |
| `StrReplaceChar` | Fusion-C | msx_fusion.h | feito | CSTR_ReplaceChar |
| `StrReverse` | Fusion-C | msx_fusion.h | feito | CSTR_Reverse |
| `StrRightTrim` | Fusion-C | msx_fusion.h | feito | CSTR_TrimRight |
| `StrSearch` | Fusion-C | msx_fusion.h | pendente |  |
| `bchput` | Fusion-C | msx_fusion.h | pendente |  |
| `num2Dec16` | Fusion-C | msx_fusion.h | pendente |  |
| `Char_IsAlpha` | MSXgl | string.h | feito | CHAR_IsAlpha |
| `Char_IsAlphaNum` | MSXgl | string.h | feito | CHAR_IsAlNum |
| `Char_IsNum` | MSXgl | string.h | feito | CHAR_IsDigit |
| `Print_Backspace` | MSXgl | print.h | pendente |  |
| `Print_Clear` | MSXgl | print.h | pendente |  |
| `Print_DrawBin8` | MSXgl | print.h | pendente |  |
| `Print_DrawBin8At` | MSXgl | print.h | pendente |  |
| `Print_DrawBox` | MSXgl | print.h | pendente |  |
| `Print_DrawChar` | MSXgl | print.h | pendente |  |
| `Print_DrawCharAt` | MSXgl | print.h | pendente |  |
| `Print_DrawCharX` | MSXgl | print.h | pendente |  |
| `Print_DrawCharXAt` | MSXgl | print.h | pendente |  |
| `Print_DrawCharYAt` | MSXgl | print.h | pendente |  |
| `Print_DrawFormat` | MSXgl | print.h | pendente |  |
| `Print_DrawHex16` | MSXgl | print.h | pendente |  |
| `Print_DrawHex16At` | MSXgl | print.h | pendente |  |
| `Print_DrawHex32` | MSXgl | print.h | pendente |  |
| `Print_DrawHex8` | MSXgl | print.h | pendente |  |
| `Print_DrawHex8At` | MSXgl | print.h | pendente |  |
| `Print_DrawInt` | MSXgl | print.h | pendente |  |
| `Print_DrawIntAt` | MSXgl | print.h | pendente |  |
| `Print_DrawLineH` | MSXgl | print.h | pendente |  |
| `Print_DrawLineV` | MSXgl | print.h | pendente |  |
| `Print_DrawText` | MSXgl | print.h | pendente |  |
| `Print_DrawTextAlign` | MSXgl | print.h | pendente |  |
| `Print_DrawTextAlignAt` | MSXgl | print.h | pendente |  |
| `Print_DrawTextAt` | MSXgl | print.h | pendente |  |
| `Print_DrawTextAtV` | MSXgl | print.h | pendente |  |
| `Print_DrawTextOutline` | MSXgl | print.h | pendente |  |
| `Print_DrawTextShadow` | MSXgl | print.h | pendente |  |
| `Print_EnableOutline` | MSXgl | print.h | pendente |  |
| `Print_EnableShadow` | MSXgl | print.h | pendente |  |
| `Print_GetFontInfo` | MSXgl | print.h | pendente |  |
| `Print_GetPatternOffset` | MSXgl | print.h | pendente |  |
| `Print_GetSpriteID` | MSXgl | print.h | pendente |  |
| `Print_GetSpritePattern` | MSXgl | print.h | pendente |  |
| `Print_Initialize` | MSXgl | print.h | pendente |  |
| `Print_Return` | MSXgl | print.h | pendente |  |
| `Print_SelectTextFont` | MSXgl | print.h | pendente |  |
| `Print_SetBitmapFont` | MSXgl | print.h | pendente |  |
| `Print_SetCharSize` | MSXgl | print.h | pendente |  |
| `Print_SetColor` | MSXgl | print.h | pendente |  |
| `Print_SetColorShade` | MSXgl | print.h | pendente |  |
| `Print_SetDirection` | MSXgl | print.h | pendente |  |
| `Print_SetFont` | MSXgl | print.h | pendente |  |
| `Print_SetFontData` | MSXgl | print.h | pendente |  |
| `Print_SetFontEx` | MSXgl | print.h | pendente |  |
| `Print_SetMode` | MSXgl | print.h | pendente |  |
| `Print_SetOutline` | MSXgl | print.h | pendente |  |
| `Print_SetPatternOffset` | MSXgl | print.h | pendente |  |
| `Print_SetPosition` | MSXgl | print.h | pendente |  |
| `Print_SetPositionX` | MSXgl | print.h | pendente |  |
| `Print_SetPositionY` | MSXgl | print.h | pendente |  |
| `Print_SetShadow` | MSXgl | print.h | pendente |  |
| `Print_SetSpriteFont` | MSXgl | print.h | pendente |  |
| `Print_SetSpriteID` | MSXgl | print.h | pendente |  |
| `Print_SetTabSize` | MSXgl | print.h | pendente |  |
| `Print_SetTextFont` | MSXgl | print.h | pendente |  |
| `Print_SetVRAMFont` | MSXgl | print.h | pendente |  |
| `Print_Space` | MSXgl | print.h | pendente |  |
| `Print_Tab` | MSXgl | print.h | pendente |  |
| `String_Copy` | MSXgl | string.h | feito | CSTR_Copy |
| `String_Format` | MSXgl | string.h | planejado | CSTR_Format |
| `String_FormatVA` | MSXgl | string.h | planejado | CSTR_Format |
| `String_FromUInt16` | MSXgl | string.h | pendente |  |
| `String_FromUInt16ZT` | MSXgl | string.h | feito | NUM_U16ToDecW |
| `String_FromUInt8` | MSXgl | string.h | pendente |  |
| `String_FromUInt8ZT` | MSXgl | string.h | feito | NUM_U16ToDecW |
| `String_Length` | MSXgl | string.h | feito | CSTR_Len |

## VDP

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `CopyRamToVram` | Fusion-C | msx_fusion.h | pendente |  |
| `CopyVramToRam` | Fusion-C | msx_fusion.h | pendente |  |
| `FillVram` | Fusion-C | msx_fusion.h | pendente |  |
| `GetVramSize` | Fusion-C | msx_fusion.h | pendente |  |
| `HMCM` | Fusion-C | vdp_graph2.h | pendente |  |
| `HMCM_SC8` | Fusion-C | vdp_graph2.h | pendente |  |
| `HMMC` | Fusion-C | vdp_graph2.h | pendente |  |
| `HMMM` | Fusion-C | vdp_graph2.h | pendente |  |
| `HMMV` | Fusion-C | vdp_graph2.h | pendente |  |
| `HideDisplay` | Fusion-C | msx_fusion.h | pendente |  |
| `LMMC` | Fusion-C | vdp_graph2.h | pendente |  |
| `LMMM` | Fusion-C | vdp_graph2.h | pendente |  |
| `LMMV` | Fusion-C | vdp_graph2.h | pendente |  |
| `PutSprite` | Fusion-C | vdp_sprites.h | pendente |  |
| `RestoreSC5Palette` | Fusion-C | vdp_graph2.h | pendente |  |
| `SC5SpriteAttribs` | Fusion-C | vdp_sprites.h | pendente |  |
| `SC5SpriteColors` | Fusion-C | vdp_sprites.h | pendente |  |
| `SC5SpritePattern` | Fusion-C | vdp_sprites.h | pendente |  |
| `SC8SpriteColors` | Fusion-C | vdp_sprites.h | pendente |  |
| `SC8SpritePattern` | Fusion-C | vdp_sprites.h | pendente |  |
| `Screen` | Fusion-C | msx_fusion.h | pendente |  |
| `SetActivePage` | Fusion-C | msx_fusion.h | pendente |  |
| `SetBorderColor` | Fusion-C | msx_fusion.h | pendente |  |
| `SetColors` | Fusion-C | msx_fusion.h | pendente |  |
| `SetDisplayPage` | Fusion-C | msx_fusion.h | pendente |  |
| `SetSC5ColorPalette` | Fusion-C | vdp_graph2.h | pendente |  |
| `SetSC5Palette` | Fusion-C | vdp_graph2.h | pendente |  |
| `SetSpritePattern` | Fusion-C | vdp_sprites.h | pendente |  |
| `ShowDisplay` | Fusion-C | msx_fusion.h | pendente |  |
| `Sprite16` | Fusion-C | vdp_sprites.h | pendente |  |
| `Sprite32Bytes` | Fusion-C | vdp_sprites.h | pendente |  |
| `Sprite8` | Fusion-C | vdp_sprites.h | pendente |  |
| `SpriteCollision` | Fusion-C | vdp_sprites.h | pendente |  |
| `SpriteCollisionX` | Fusion-C | vdp_sprites.h | pendente |  |
| `SpriteCollisionY` | Fusion-C | vdp_sprites.h | pendente |  |
| `SpriteDouble` | Fusion-C | vdp_sprites.h | pendente |  |
| `SpriteOff` | Fusion-C | vdp_sprites.h | pendente |  |
| `SpriteOn` | Fusion-C | vdp_sprites.h | pendente |  |
| `SpriteReset` | Fusion-C | vdp_sprites.h | pendente |  |
| `SpriteSmall` | Fusion-C | vdp_sprites.h | pendente |  |
| `VDP50Hz` | Fusion-C | msx_fusion.h | pendente |  |
| `VDP60Hz` | Fusion-C | msx_fusion.h | pendente |  |
| `VDPLinesSwitch` | Fusion-C | msx_fusion.h | pendente |  |
| `VDPstatus` | Fusion-C | msx_fusion.h | pendente |  |
| `VDPstatusNi` | Fusion-C | msx_fusion.h | pendente |  |
| `VDPwrite` | Fusion-C | msx_fusion.h | pendente |  |
| `VDPwriteNi` | Fusion-C | msx_fusion.h | pendente |  |
| `Vpeek` | Fusion-C | msx_fusion.h | pendente |  |
| `VpeekFirst` | Fusion-C | msx_fusion.h | pendente |  |
| `Vpoke` | Fusion-C | msx_fusion.h | pendente |  |
| `VpokeFirst` | Fusion-C | msx_fusion.h | pendente |  |
| `Width` | Fusion-C | msx_fusion.h | pendente |  |
| `YMMM` | Fusion-C | vdp_graph2.h | pendente |  |
| `fLMMM` | Fusion-C | vdp_graph2.h | pendente |  |
| `VDP_CleanBlinkScreen` | MSXgl | vdp.h | pendente |  |
| `VDP_ClearVRAM` | MSXgl | vdp.h | feito | VDP_ClearVRAM |
| `VDP_CommandCustomR32` | MSXgl | vdp.h | pendente |  |
| `VDP_CommandCustomR36` | MSXgl | vdp.h | pendente |  |
| `VDP_CommandReadLoop` | MSXgl | vdp.h | pendente |  |
| `VDP_CommandSetupR32` | MSXgl | vdp.h | pendente |  |
| `VDP_CommandSetupR36` | MSXgl | vdp.h | pendente |  |
| `VDP_CommandWait` | MSXgl | vdp.h | pendente |  |
| `VDP_CommandWriteLoop` | MSXgl | vdp.h | pendente |  |
| `VDP_DisableSprite` | MSXgl | vdp.h | pendente |  |
| `VDP_DisableSpritesFrom` | MSXgl | vdp.h | pendente |  |
| `VDP_EnableDisplay` | MSXgl | vdp.h | pendente |  |
| `VDP_EnableHBlank` | MSXgl | vdp.h | pendente |  |
| `VDP_EnableMask` | MSXgl | vdp.h | pendente |  |
| `VDP_EnableSprite` | MSXgl | vdp.h | pendente |  |
| `VDP_EnableTransparency` | MSXgl | vdp.h | pendente |  |
| `VDP_EnableVBlank` | MSXgl | vdp.h | pendente |  |
| `VDP_ExpendCommand` | MSXgl | vdp.h | pendente |  |
| `VDP_FastFillVRAM_16K` | MSXgl | vdp.h | pendente |  |
| `VDP_FillLayout_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_FillScreen_GM1` | MSXgl | vdp.h | pendente |  |
| `VDP_FillScreen_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_FillVRAM_128K` | MSXgl | vdp.h | pendente |  |
| `VDP_FillVRAM_16K` | MSXgl | vdp.h | feito | VDP_FillVRAM |
| `VDP_GetColorTable` | MSXgl | vdp.h | pendente |  |
| `VDP_GetColorTable_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_GetFrequency` | MSXgl | vdp.h | pendente |  |
| `VDP_GetLayoutTable` | MSXgl | vdp.h | pendente |  |
| `VDP_GetMode` | MSXgl | vdp.h | feito | VDP_GetMode |
| `VDP_GetPatternTable` | MSXgl | vdp.h | pendente |  |
| `VDP_GetPatternTable_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_GetSpriteAttributeTable` | MSXgl | vdp.h | pendente |  |
| `VDP_GetSpriteColorTable` | MSXgl | vdp.h | pendente |  |
| `VDP_GetSpritePatternTable` | MSXgl | vdp.h | pendente |  |
| `VDP_GetVersion` | MSXgl | vdp.h | pendente |  |
| `VDP_HideAllSprites` | MSXgl | vdp.h | feito | VDP_SpriteHideAll |
| `VDP_HideSprite` | MSXgl | vdp.h | feito | VDP_SpriteHide |
| `VDP_Initialize` | MSXgl | vdp.h | pendente |  |
| `VDP_IsBitmapMode` | MSXgl | vdp.h | pendente |  |
| `VDP_IsPatternMode` | MSXgl | vdp.h | pendente |  |
| `VDP_LoadBankColor_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_LoadBankPattern_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_LoadColor_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_LoadPattern_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_LoadSpritePattern` | MSXgl | vdp.h | feito | VDP_SpriteDefine |
| `VDP_Peek_128K` | MSXgl | vdp.h | feito | VDP_VPeek |
| `VDP_Peek_16K` | MSXgl | vdp.h | pendente |  |
| `VDP_Peek_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_Poke_128K` | MSXgl | vdp.h | feito | VDP_VPoke |
| `VDP_Poke_16K` | MSXgl | vdp.h | pendente |  |
| `VDP_Poke_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_ReadDefaultStatus` | MSXgl | vdp.h | pendente |  |
| `VDP_ReadStatus` | MSXgl | vdp.h | feito | VDP_ReadStatus |
| `VDP_ReadVRAM_128K` | MSXgl | vdp.h | pendente |  |
| `VDP_ReadVRAM_16K` | MSXgl | vdp.h | feito | VDP_ReadVRAM |
| `VDP_RegWrite` | MSXgl | vdp.h | pendente |  |
| `VDP_RegWriteBak` | MSXgl | vdp.h | feito | VDP_SetReg |
| `VDP_RegWriteBakMask` | MSXgl | vdp.h | feito | VDP_UpdateReg |
| `VDP_ResetVRAMAddrMSB` | MSXgl | vdp.h | pendente |  |
| `VDP_SetAdjustOffset` | MSXgl | vdp.h | pendente |  |
| `VDP_SetAdjustOffsetXY` | MSXgl | vdp.h | pendente |  |
| `VDP_SetBackdropColor` | MSXgl | vdp.h | feito | VDP_Backdrop |
| `VDP_SetBlinkChunk` | MSXgl | vdp.h | pendente |  |
| `VDP_SetBlinkChunkMask` | MSXgl | vdp.h | pendente |  |
| `VDP_SetBlinkChunkX` | MSXgl | vdp.h | pendente |  |
| `VDP_SetBlinkColor` | MSXgl | vdp.h | pendente |  |
| `VDP_SetBlinkColor2` | MSXgl | vdp.h | pendente |  |
| `VDP_SetBlinkLine` | MSXgl | vdp.h | pendente |  |
| `VDP_SetBlinkScreen` | MSXgl | vdp.h | pendente |  |
| `VDP_SetBlinkTile` | MSXgl | vdp.h | pendente |  |
| `VDP_SetBlinkTime` | MSXgl | vdp.h | pendente |  |
| `VDP_SetBlinkTime2` | MSXgl | vdp.h | pendente |  |
| `VDP_SetColor` | MSXgl | vdp.h | pendente |  |
| `VDP_SetColor2` | MSXgl | vdp.h | pendente |  |
| `VDP_SetColorTable` | MSXgl | vdp.h | pendente |  |
| `VDP_SetColorTableEx` | MSXgl | vdp.h | pendente |  |
| `VDP_SetDefaultPalette` | MSXgl | vdp.h | feito | VDP_SetDefaultPalette |
| `VDP_SetFrameRender` | MSXgl | vdp.h | pendente |  |
| `VDP_SetFrequency` | MSXgl | vdp.h | feito | VDP_SetRefresh |
| `VDP_SetGrayScale` | MSXgl | vdp.h | pendente |  |
| `VDP_SetHBlankLine` | MSXgl | vdp.h | pendente |  |
| `VDP_SetHorizontalMode` | MSXgl | vdp.h | pendente |  |
| `VDP_SetHorizontalOffset` | MSXgl | vdp.h | pendente |  |
| `VDP_SetInfiniteBlink` | MSXgl | vdp.h | pendente |  |
| `VDP_SetInterlace` | MSXgl | vdp.h | pendente |  |
| `VDP_SetLayoutTable` | MSXgl | vdp.h | pendente |  |
| `VDP_SetLayoutTableEx` | MSXgl | vdp.h | pendente |  |
| `VDP_SetLineCount` | MSXgl | vdp.h | feito | VDP_SetLines |
| `VDP_SetMSX1Palette` | MSXgl | vdp.h | feito | VDP_SetMSX1Palette |
| `VDP_SetMode` | MSXgl | vdp.h | feito | VDP_SetMode |
| `VDP_SetModeFlag` | MSXgl | vdp.h | pendente |  |
| `VDP_SetPage` | MSXgl | vdp.h | feito | VDP_SetDisplayPage |
| `VDP_SetPageAlternance` | MSXgl | vdp.h | pendente |  |
| `VDP_SetPalette` | MSXgl | vdp.h | feito | VDP_SetPaletteBlock |
| `VDP_SetPaletteEntry` | MSXgl | vdp.h | feito | VDP_SetPaletteEntry |
| `VDP_SetPatternTable` | MSXgl | vdp.h | pendente |  |
| `VDP_SetPatternTableEx` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSprite` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteAttributeTable` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteAttributeTableEx` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteColorSM1` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteData` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteExMultiColor` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteExUniColor` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteFlag` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteMultiColor` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpritePattern` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpritePatternTable` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpritePatternTableEx` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpritePosition` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpritePositionX` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpritePositionY` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteSM1` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteTables` | MSXgl | vdp.h | pendente |  |
| `VDP_SetSpriteUniColor` | MSXgl | vdp.h | pendente |  |
| `VDP_SetVerticalOffset` | MSXgl | vdp.h | pendente |  |
| `VDP_SetYJK` | MSXgl | vdp.h | pendente |  |
| `VDP_WriteLayout_GM2` | MSXgl | vdp.h | pendente |  |
| `VDP_WriteVRAM_128K` | MSXgl | vdp.h | pendente |  |
| `VDP_WriteVRAM_16K` | MSXgl | vdp.h | feito | VDP_WriteVRAM |

## Draw

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `BoxFill` | Fusion-C | vdp_graph2.h | pendente |  |
| `BoxLine` | Fusion-C | vdp_graph2.h | pendente |  |
| `Circle` | Fusion-C | vdp_circle.h | pendente |  |
| `CircleFilled` | Fusion-C | vdp_circle.h | pendente |  |
| `Clear1px` | Fusion-C | vdp_graph1.h | pendente |  |
| `Clear8px` | Fusion-C | vdp_graph1.h | pendente |  |
| `Draw` | Fusion-C | vdp_graph2.h | pendente |  |
| `Get1px` | Fusion-C | vdp_graph1.h | pendente |  |
| `Get8px` | Fusion-C | vdp_graph1.h | pendente |  |
| `GetCol8px` | Fusion-C | vdp_graph1.h | pendente |  |
| `Line` | Fusion-C | vdp_graph2.h | pendente |  |
| `Paint` | Fusion-C | vdp_graph2.h | pendente |  |
| `PgetXY` | Fusion-C | vdp_graph2.h | pendente |  |
| `Point` | Fusion-C | vdp_graph2.h | pendente |  |
| `Pset` | Fusion-C | vdp_graph2.h | pendente |  |
| `PsetXY` | Fusion-C | vdp_graph2.h | pendente |  |
| `ReadBlock` | Fusion-C | vdp_graph1.h | pendente |  |
| `ReadScr` | Fusion-C | vdp_graph2.h | pendente |  |
| `SC2Circle` | Fusion-C | vdp_circle.h | pendente |  |
| `SC2CircleFilled` | Fusion-C | vdp_circle.h | pendente |  |
| `SC2Draw` | Fusion-C | vdp_graph1.h | pendente |  |
| `SC2Line` | Fusion-C | vdp_graph1.h | pendente |  |
| `SC2Paint` | Fusion-C | vdp_graph1.h | pendente |  |
| `SC2Point` | Fusion-C | vdp_graph1.h | pendente |  |
| `SC2Pset` | Fusion-C | vdp_graph1.h | pendente |  |
| `SC2ReadScr` | Fusion-C | vdp_graph1.h | pendente |  |
| `SC2Rect` | Fusion-C | vdp_graph1.h | pendente |  |
| `SC2WriteScr` | Fusion-C | vdp_graph1.h | pendente |  |
| `Set1px` | Fusion-C | vdp_graph1.h | pendente |  |
| `Set8px` | Fusion-C | vdp_graph1.h | pendente |  |
| `SetCol8px` | Fusion-C | vdp_graph1.h | pendente |  |
| `WriteBlock` | Fusion-C | vdp_graph1.h | pendente |  |
| `WriteScr` | Fusion-C | vdp_graph2.h | pendente |  |
| `vMSX` | Fusion-C | vdp_graph2.h | pendente |  |
| `Draw_Box` | MSXgl | draw.h | pendente |  |
| `Draw_Circle` | MSXgl | draw.h | pendente |  |
| `Draw_FillBox` | MSXgl | draw.h | pendente |  |
| `Draw_Line` | MSXgl | draw.h | pendente |  |
| `Draw_LineH` | MSXgl | draw.h | pendente |  |
| `Draw_LineV` | MSXgl | draw.h | pendente |  |
| `Draw_Point` | MSXgl | draw.h | pendente |  |

## Tile

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `Tile_DrawBlock` | MSXgl | tile.h | pendente |  |
| `Tile_DrawMapChunk` | MSXgl | tile.h | pendente |  |
| `Tile_DrawScreen` | MSXgl | tile.h | pendente |  |
| `Tile_DrawTile` | MSXgl | tile.h | pendente |  |
| `Tile_FillBank` | MSXgl | tile.h | pendente |  |
| `Tile_FillScreen` | MSXgl | tile.h | pendente |  |
| `Tile_FillTile` | MSXgl | tile.h | pendente |  |
| `Tile_GetBankAddress` | MSXgl | tile.h | pendente |  |
| `Tile_GetBankAddressEx` | MSXgl | tile.h | pendente |  |
| `Tile_LoadBank` | MSXgl | tile.h | pendente |  |
| `Tile_LoadBankEx` | MSXgl | tile.h | pendente |  |
| `Tile_SelectBank` | MSXgl | tile.h | pendente |  |
| `Tile_SetBankPage` | MSXgl | tile.h | pendente |  |
| `Tile_SetDrawPage` | MSXgl | tile.h | pendente |  |

## Scroll

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `SetScrollH` | Fusion-C | msx_fusion.h | pendente |  |
| `SetScrollV` | Fusion-C | msx_fusion.h | pendente |  |
| `Scroll_HBlankAdjust` | MSXgl | scroll.h | pendente |  |
| `Scroll_Initialize` | MSXgl | scroll.h | pendente |  |
| `Scroll_SetOffsetH` | MSXgl | scroll.h | pendente |  |
| `Scroll_SetOffsetV` | MSXgl | scroll.h | pendente |  |
| `Scroll_Update` | MSXgl | scroll.h | pendente |  |

## Teclado

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `ChangeCap` | Fusion-C | msx_fusion.h | pendente |  |
| `CheckBreak` | Fusion-C | msx_fusion.h | pendente |  |
| `FunctionKeys` | Fusion-C | msx_fusion.h | pendente |  |
| `GetKeyMatrix` | Fusion-C | msx_fusion.h | pendente |  |
| `Inkey` | Fusion-C | msx_fusion.h | pendente |  |
| `KeySound` | Fusion-C | msx_fusion.h | pendente |  |
| `KeyboardRead` | Fusion-C | msx_fusion.h | pendente |  |
| `KillKeyBuffer` | Fusion-C | msx_fusion.h | pendente |  |
| `WaitKey` | Fusion-C | msx_fusion.h | feito | CON_ReadKey |
| `IPM_GetInputState` | MSXgl | input_manager.h | pendente |  |
| `IPM_GetInputTimer` | MSXgl | input_manager.h | pendente |  |
| `IPM_GetStatus` | MSXgl | input_manager.h | pendente |  |
| `IPM_GetStickDirection` | MSXgl | input_manager.h | pendente |  |
| `IPM_Initialize` | MSXgl | input_manager.h | pendente |  |
| `IPM_RegisterEvent` | MSXgl | input_manager.h | pendente |  |
| `IPM_SetTimer` | MSXgl | input_manager.h | pendente |  |
| `IPM_Update` | MSXgl | input_manager.h | pendente |  |
| `Input_Detect` | MSXgl | input.h | pendente |  |
| `Keyboard_IsKeyPressed` | MSXgl | keyboard.h | pendente |  |
| `Keyboard_IsKeyPushed` | MSXgl | keyboard.h | pendente |  |
| `Keyboard_Read` | MSXgl | keyboard.h | pendente |  |
| `Keyboard_ReadAsJoystick` | MSXgl | keyboard.h | pendente |  |
| `Keyboard_SetBuffer` | MSXgl | keyboard.h | pendente |  |
| `Keyboard_Update` | MSXgl | keyboard.h | pendente |  |

## Joystick e mouse

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `JoystickRead` | Fusion-C | msx_fusion.h | pendente |  |
| `MouseRead` | Fusion-C | msx_fusion.h | pendente |  |
| `MouseReadTo` | Fusion-C | msx_fusion.h | pendente |  |
| `TriggerRead` | Fusion-C | msx_fusion.h | pendente |  |
| `Joystick_GetDirection` | MSXgl | joystick.h | pendente |  |
| `Joystick_GetDirectionChange` | MSXgl | joystick.h | pendente |  |
| `Joystick_IsButtonPressed` | MSXgl | joystick.h | pendente |  |
| `Joystick_IsButtonPushed` | MSXgl | joystick.h | pendente |  |
| `Joystick_Read` | MSXgl | joystick.h | pendente |  |
| `Joystick_Update` | MSXgl | joystick.h | pendente |  |
| `Mouse_GetAdjustedOffsetX` | MSXgl | mouse.h | pendente |  |
| `Mouse_GetAdjustedOffsetY` | MSXgl | mouse.h | pendente |  |
| `Mouse_GetOffsetX` | MSXgl | mouse.h | pendente |  |
| `Mouse_GetOffsetY` | MSXgl | mouse.h | pendente |  |
| `Mouse_IsButtonClick` | MSXgl | mouse.h | pendente |  |
| `Mouse_IsButtonPress` | MSXgl | mouse.h | pendente |  |
| `Mouse_Read` | MSXgl | mouse.h | pendente |  |

## PSG

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `GetSound` | Fusion-C | psg.h | pendente |  |
| `InitPSG` | Fusion-C | msx_fusion.h | pendente |  |
| `PSGRead` | Fusion-C | msx_fusion.h | feito | PSG_Read |
| `PSGwrite` | Fusion-C | msx_fusion.h | feito | PSG_Write |
| `PlayEnvelope` | Fusion-C | psg.h | pendente |  |
| `SetChannel` | Fusion-C | psg.h | pendente |  |
| `SetChannelA` | Fusion-C | psg.h | pendente |  |
| `SetEnvelopePeriod` | Fusion-C | psg.h | pendente |  |
| `SetNoisePeriod` | Fusion-C | psg.h | pendente |  |
| `SetTonePeriod` | Fusion-C | psg.h | pendente |  |
| `SetVolume` | Fusion-C | psg.h | pendente |  |
| `SilencePSG` | Fusion-C | psg.h | feito | PSG_MuteAll |
| `Sound` | Fusion-C | psg.h | pendente |  |
| `SoundFX` | Fusion-C | psg.h | pendente |  |
| `PSG_Apply` | MSXgl | psg.h | pendente |  |
| `PSG_EnableEnvelope` | MSXgl | psg.h | pendente |  |
| `PSG_EnableNoise` | MSXgl | psg.h | pendente |  |
| `PSG_EnableTone` | MSXgl | psg.h | pendente |  |
| `PSG_GetRegister` | MSXgl | psg.h | feito | PSG_Read |
| `PSG_Mute` | MSXgl | psg.h | pendente |  |
| `PSG_Resume` | MSXgl | psg.h | pendente |  |
| `PSG_SetEnvelope` | MSXgl | psg.h | pendente |  |
| `PSG_SetMixer` | MSXgl | psg.h | pendente |  |
| `PSG_SetNoise` | MSXgl | psg.h | pendente |  |
| `PSG_SetRegister` | MSXgl | psg.h | feito | PSG_Write |
| `PSG_SetShape` | MSXgl | psg.h | pendente |  |
| `PSG_SetTone` | MSXgl | psg.h | pendente |  |
| `PSG_SetVolume` | MSXgl | psg.h | pendente |  |

## Play (players)

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `CovoxPlay` | Fusion-C | msx_fusion.h | pendente |  |
| `FreeFX` | Fusion-C | ayfx_player.h | pendente |  |
| `InitFX` | Fusion-C | ayfx_player.h | pendente |  |
| `InstallReplayer` | Fusion-C | pt3replayer.h | pendente |  |
| `PCMPlay` | Fusion-C | msx_fusion.h | pendente |  |
| `PT3Init` | Fusion-C | pt3replayer.h | pendente |  |
| `PT3Mute` | Fusion-C | pt3replayer.h | pendente |  |
| `PT3Play` | Fusion-C | pt3replayer.h | pendente |  |
| `PT3Rout` | Fusion-C | pt3replayer.h | pendente |  |
| `PlayFX` | Fusion-C | ayfx_player.h | pendente |  |
| `Reg7Patch` | Fusion-C | ayfx_player.h | pendente |  |
| `TestFX` | Fusion-C | ayfx_player.h | pendente |  |
| `UninstallReplayer` | Fusion-C | pt3replayer.h | pendente |  |
| `UpdateFX` | Fusion-C | ayfx_player.h | pendente |  |
| `playsnd` | Fusion-C | ayfx_player.h | pendente |  |
| `setnoises` | Fusion-C | ayfx_player.h | pendente |  |
| `AKG_InitSFX` | MSXgl | arkos/akg_player.h | pendente |  |
| `AKG_IsPlaying` | MSXgl | arkos/akg_player.h | pendente |  |
| `AKG_Play` | MSXgl | arkos/akg_player.h | pendente |  |
| `AKG_PlaySFX` | MSXgl | arkos/akg_player.h | pendente |  |
| `AKG_SetEventCallback` | MSXgl | arkos/akg_player.h | pendente |  |
| `AKG_Stop` | MSXgl | arkos/akg_player.h | pendente |  |
| `AKG_StopSFX` | MSXgl | arkos/akg_player.h | pendente |  |
| `AKG_Update` | MSXgl | arkos/akg_player.h | pendente |  |
| `AKM_InitSFX` | MSXgl | arkos/akm_player.h | pendente |  |
| `AKM_IsPlaying` | MSXgl | arkos/akm_player.h | pendente |  |
| `AKM_Play` | MSXgl | arkos/akm_player.h | pendente |  |
| `AKM_PlaySFX` | MSXgl | arkos/akm_player.h | pendente |  |
| `AKM_Stop` | MSXgl | arkos/akm_player.h | pendente |  |
| `AKM_StopSFX` | MSXgl | arkos/akm_player.h | pendente |  |
| `AKM_Update` | MSXgl | arkos/akm_player.h | pendente |  |
| `AKY6ch_Play` | MSXgl | arkos/aky_6ch_player.h | pendente |  |
| `AKY6ch_Update` | MSXgl | arkos/aky_6ch_player.h | pendente |  |
| `AKYDarky_Play` | MSXgl | arkos/aky_darky_player.h | pendente |  |
| `AKYDarky_Update` | MSXgl | arkos/aky_darky_player.h | pendente |  |
| `AKY_Play` | MSXgl | arkos/aky_player.h | pendente |  |
| `AKY_Update` | MSXgl | arkos/aky_player.h | pendente |  |
| `EmptyCB` | MSXgl | pt3/pt3_player.h | pendente |  |
| `LVGM_Decode` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_GetDefaultPSGValue` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_GetDevices` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_IncludeOPL` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_IncludeOPLL` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_IncludePSG` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_IncludeSCC` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_IsFrequency50Hz` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_IsFrequency60Hz` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_IsPlaying` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_Pause` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_Play` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_Resume` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_SetFrequency50Hz` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_SetFrequency60Hz` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_SetNotifyCallback` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_SetPointer` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `LVGM_Stop` | MSXgl | vgm/lvgm_player.h | pendente |  |
| `MGLV_Decode` | MSXgl | mglv/mglv_player.h | pendente |  |
| `MGLV_GetSegmentSize` | MSXgl | mglv/mglv_player.h | pendente |  |
| `MGLV_GetVersion` | MSXgl | mglv/mglv_player.h | pendente |  |
| `MGLV_Init` | MSXgl | mglv/mglv_player.h | pendente |  |
| `MGLV_IsLooping` | MSXgl | mglv/mglv_player.h | pendente |  |
| `MGLV_Play` | MSXgl | mglv/mglv_player.h | pendente |  |
| `MGLV_SetFrameDuration` | MSXgl | mglv/mglv_player.h | pendente |  |
| `MGLV_SetLoop` | MSXgl | mglv/mglv_player.h | pendente |  |
| `MGLV_VBlankHandler` | MSXgl | mglv/mglv_player.h | pendente |  |
| `NDP_FadeOut` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_GetFadeType` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_GetLoopCount` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_GetSFXNumber` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_GetSFXOffset` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_GetStatus` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_GetVersion` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_HasMusicEnded` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_HasTrackEnded` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_Initialize` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_IsFadding` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_IsInitialized` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_IsPlaying` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_IsStopped` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_MuteChannel` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_MuteChannelA` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_MuteChannelB` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_MuteChannelC` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_Play` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_PlayFadeIn` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_PlaySFX` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_Release` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_SetInitFlag` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_SetMusicData` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_SetSFXData` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_SetToneColor` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_SetVolume` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_Stop` | MSXgl | ndp/ndp_player.h | pendente |  |
| `NDP_Update` | MSXgl | ndp/ndp_player.h | pendente |  |
| `PCM_Play` | MSXgl | pcm/pcmplay.h | pendente |  |
| `PCM_Play_11K` | MSXgl | pcm/pcmenc.h | pendente |  |
| `PCM_Play_22K` | MSXgl | pcm/pcmenc.h | pendente |  |
| `PCM_Play_44K` | MSXgl | pcm/pcmenc.h | pendente |  |
| `PCM_Play_8K` | MSXgl | pcm/pcmenc.h | pendente |  |
| `PT3_Decode` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_GetFrequency` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_GetLoop` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_GetPSGRegister` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_GetPattern` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_GetVolume` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_Init` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_InitSong` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_IsPlaying` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_Mute` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_Pause` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_Play` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_ResetFinishCB` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_Resume` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_SetFinishCB` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_SetLoop` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_SetNoteTable` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_Silence` | MSXgl | pt3/pt3_player.h | pendente |  |
| `PT3_UpdatePSG` | MSXgl | pt3/pt3_player.h | pendente |  |
| `TriloSCC_Apply` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_FadeOut` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_Initialize` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_LoadMusic` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_Pause` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_Resume` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_SetBalancePSG` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_SetBalanceSCC` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_SetFrequency` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_SetToneTable` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_Silent` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSCC_Update` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSFX_GetNumber` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSFX_Initialize` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSFX_Play` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSFX_SetBalancePSG` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSFX_SetBalanceSCC` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSFX_SetBank` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `TriloSFX_Update` | MSXgl | trilo/trilo_scc_player.h | pendente |  |
| `VGM_ContainsMSXAudio` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_ContainsMSXMusic` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_ContainsPSG` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_ContainsSCC` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_Decode` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_IsPlaying` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_Pause` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_Play` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_Resume` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_SetFrequency50Hz` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_SetFrequency60Hz` | MSXgl | vgm/vgm_player.h | pendente |  |
| `VGM_Stop` | MSXgl | vgm/vgm_player.h | pendente |  |
| `WYZ_Decode` | MSXgl | wyz/wyz_player.h | pendente |  |
| `WYZ_InitPlayer` | MSXgl | wyz/wyz_player2.h | pendente |  |
| `WYZ_Initialize` | MSXgl | wyz/wyz_player2.h | pendente |  |
| `WYZ_IsFinished` | MSXgl | wyz/wyz_player.h | pendente |  |
| `WYZ_Pause` | MSXgl | wyz/wyz_player.h | pendente |  |
| `WYZ_Play` | MSXgl | wyz/wyz_player2.h | pendente |  |
| `WYZ_PlayAY` | MSXgl | wyz/wyz_player.h | pendente |  |
| `WYZ_PlayFX` | MSXgl | wyz/wyz_player.h | pendente |  |
| `WYZ_Resume` | MSXgl | wyz/wyz_player.h | pendente |  |
| `WYZ_SetLoop` | MSXgl | wyz/wyz_player.h | pendente |  |
| `WYZ_Stop` | MSXgl | wyz/wyz_player2.h | pendente |  |
| `ayFX_GetBankNumber` | MSXgl | ayfx/ayfx_player.h | pendente |  |
| `ayFX_GetChannel` | MSXgl | ayfx/ayfx_player.h | pendente |  |
| `ayFX_InitBank` | MSXgl | ayfx/ayfx_player.h | pendente |  |
| `ayFX_Mute` | MSXgl | ayfx/ayfx_player.h | pendente |  |
| `ayFX_Play` | MSXgl | ayfx/ayfx_player.h | pendente |  |
| `ayFX_PlayBank` | MSXgl | ayfx/ayfx_player.h | pendente |  |
| `ayFX_SendToPSG` | MSXgl | standalone/ayfx_player.h | pendente |  |
| `ayFX_SetChannel` | MSXgl | ayfx/ayfx_player.h | pendente |  |
| `ayFX_SetFinishCB` | MSXgl | ayfx/ayfx_player.h | pendente |  |
| `ayFX_SetMode` | MSXgl | ayfx/ayfx_player.h | pendente |  |
| `ayFX_Stop` | MSXgl | ayfx/ayfx_player.h | pendente |  |
| `ayFX_Update` | MSXgl | ayfx/ayfx_player.h | pendente |  |

## MSX-Music

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `MSXMusic_Detect` | MSXgl | msx-music.h | pendente |  |
| `MSXMusic_GetRegister` | MSXgl | msx-music.h | pendente |  |
| `MSXMusic_GetSlotId` | MSXgl | msx-music.h | pendente |  |
| `MSXMusic_Initialize` | MSXgl | msx-music.h | pendente |  |
| `MSXMusic_Mute` | MSXgl | msx-music.h | pendente |  |
| `MSXMusic_Resume` | MSXgl | msx-music.h | pendente |  |
| `MSXMusic_SetRegister` | MSXgl | msx-music.h | pendente |  |

## MSX-Audio

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `MSXAudio_Detect` | MSXgl | msx-audio.h | pendente |  |
| `MSXAudio_GetRegister` | MSXgl | msx-audio.h | pendente |  |
| `MSXAudio_Initialize` | MSXgl | msx-audio.h | pendente |  |
| `MSXAudio_Mute` | MSXgl | msx-audio.h | pendente |  |
| `MSXAudio_Resume` | MSXgl | msx-audio.h | pendente |  |
| `MSXAudio_SetRegister` | MSXgl | msx-audio.h | pendente |  |

## SCC

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `SCC_GetRegister` | MSXgl | scc.h | pendente |  |
| `SCC_Initialize` | MSXgl | scc.h | pendente |  |
| `SCC_LoadWaveform` | MSXgl | scc.h | pendente |  |
| `SCC_Mute` | MSXgl | scc.h | pendente |  |
| `SCC_Resume` | MSXgl | scc.h | pendente |  |
| `SCC_Select` | MSXgl | scc.h | pendente |  |
| `SCC_SetFrequency` | MSXgl | scc.h | pendente |  |
| `SCC_SetMixer` | MSXgl | scc.h | pendente |  |
| `SCC_SetRegister` | MSXgl | scc.h | pendente |  |
| `SCC_SetSlot` | MSXgl | scc.h | pendente |  |
| `SCC_SetVolume` | MSXgl | scc.h | pendente |  |

## BIOS

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `BIOS_ApplyBorder` | MSXgl | bios.h | pendente |  |
| `BIOS_ApplyColor` | MSXgl | bios.h | pendente |  |
| `BIOS_BackupHook` | MSXgl | bios_hook.h | pendente |  |
| `BIOS_Beep` | MSXgl | bios.h | pendente |  |
| `BIOS_ClearHook` | MSXgl | bios_hook.h | pendente |  |
| `BIOS_ClearScreen` | MSXgl | bios.h | pendente |  |
| `BIOS_ClearSprites` | MSXgl | bios.h | pendente |  |
| `BIOS_CopyFromVRAM` | MSXgl | bios.h | pendente |  |
| `BIOS_CopyToVRAM` | MSXgl | bios.h | pendente |  |
| `BIOS_DisableScreen` | MSXgl | bios.h | pendente |  |
| `BIOS_DisplayScreen` | MSXgl | bios.h | pendente |  |
| `BIOS_EnableScreen` | MSXgl | bios.h | pendente |  |
| `BIOS_Exit` | MSXgl | bios.h | pendente |  |
| `BIOS_FillVRAM` | MSXgl | bios.h | pendente |  |
| `BIOS_GetCPUMode` | MSXgl | bios.h | pendente |  |
| `BIOS_GetCharAddress` | MSXgl | bios.h | pendente |  |
| `BIOS_GetFontAddress` | MSXgl | bios.h | pendente |  |
| `BIOS_GetJoystickDirection` | MSXgl | bios.h | pendente |  |
| `BIOS_GetJoystickTrigger` | MSXgl | bios.h | pendente |  |
| `BIOS_GetKeyboardMatrix` | MSXgl | bios.h | pendente |  |
| `BIOS_GetMSXVersion` | MSXgl | bios.h | pendente |  |
| `BIOS_GetPaddle` | MSXgl | bios.h | pendente |  |
| `BIOS_GetSpriteAttributeAddress` | MSXgl | bios.h | pendente |  |
| `BIOS_GetSpriteOverScanId` | MSXgl | bios.h | pendente |  |
| `BIOS_GetSpritePatternAddress` | MSXgl | bios.h | pendente |  |
| `BIOS_GetSpriteSize` | MSXgl | bios.h | pendente |  |
| `BIOS_GetTouchPad` | MSXgl | bios.h | pendente |  |
| `BIOS_GetVDPReadPort` | MSXgl | bios.h | pendente |  |
| `BIOS_GetVDPWritePort` | MSXgl | bios.h | pendente |  |
| `BIOS_GraphPrint` | MSXgl | bios.h | pendente |  |
| `BIOS_GraphPrintAt` | MSXgl | bios.h | pendente |  |
| `BIOS_GraphPrintChar` | MSXgl | bios.h | pendente |  |
| `BIOS_GraphSetCursor` | MSXgl | bios.h | pendente |  |
| `BIOS_GraphSetOperator` | MSXgl | bios.h | pendente |  |
| `BIOS_HasCharacter` | MSXgl | bios.h | pendente |  |
| `BIOS_HideSprite` | MSXgl | bios.h | pendente |  |
| `BIOS_InitPSG` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen0` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen0Color` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen0Ex` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen1` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen1Color` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen1Ex` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen2` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen2Color` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen2Ex` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen3` | MSXgl | bios.h | pendente |  |
| `BIOS_InitScreen3Ex` | MSXgl | bios.h | pendente |  |
| `BIOS_InterSlotCall` | MSXgl | bios.h | pendente |  |
| `BIOS_InterSlotRead` | MSXgl | bios.h | pendente |  |
| `BIOS_InterSlotWrite` | MSXgl | bios.h | pendente |  |
| `BIOS_IsKeyPressed` | MSXgl | bios.h | pendente |  |
| `BIOS_IsPSGPlaying` | MSXgl | bios.h | pendente |  |
| `BIOS_IsPrinterReady` | MSXgl | bios.h | pendente |  |
| `BIOS_IsSpriteCollision` | MSXgl | bios.h | pendente |  |
| `BIOS_IsSpriteOverScan` | MSXgl | bios.h | pendente |  |
| `BIOS_PlayPSG` | MSXgl | bios.h | pendente |  |
| `BIOS_PrinterChangePage` | MSXgl | bios.h | pendente |  |
| `BIOS_PrinterOutput` | MSXgl | bios.h | pendente |  |
| `BIOS_PrinterSendChar` | MSXgl | bios.h | pendente |  |
| `BIOS_PrinterSendString` | MSXgl | bios.h | pendente |  |
| `BIOS_ReadPSG` | MSXgl | bios.h | pendente |  |
| `BIOS_ReadVDP` | MSXgl | bios.h | pendente |  |
| `BIOS_ReadVRAM` | MSXgl | bios.h | pendente |  |
| `BIOS_Reboot` | MSXgl | bios.h | pendente |  |
| `BIOS_Set1BitSound` | MSXgl | bios.h | pendente |  |
| `BIOS_SetAddressForRead` | MSXgl | bios.h | pendente |  |
| `BIOS_SetAddressForWrite` | MSXgl | bios.h | pendente |  |
| `BIOS_SetCPUMode` | MSXgl | bios.h | pendente |  |
| `BIOS_SetColor` | MSXgl | bios.h | pendente |  |
| `BIOS_SetHookCallback` | MSXgl | bios_hook.h | pendente |  |
| `BIOS_SetHookDirectCallback` | MSXgl | bios_hook.h | pendente |  |
| `BIOS_SetHookInterSlotCallback` | MSXgl | bios_hook.h | pendente |  |
| `BIOS_SetKeyClick` | MSXgl | bios.h | pendente |  |
| `BIOS_SetScreen0` | MSXgl | bios.h | pendente |  |
| `BIOS_SetScreen1` | MSXgl | bios.h | pendente |  |
| `BIOS_SetScreen2` | MSXgl | bios.h | pendente |  |
| `BIOS_SetScreen3` | MSXgl | bios.h | pendente |  |
| `BIOS_SetScreenMode` | MSXgl | bios.h | pendente |  |
| `BIOS_SetSprite` | MSXgl | bios.h | pendente |  |
| `BIOS_SetSpriteColor` | MSXgl | bios.h | pendente |  |
| `BIOS_SetSpriteData` | MSXgl | bios.h | pendente |  |
| `BIOS_SetSpriteMode` | MSXgl | bios.h | pendente |  |
| `BIOS_SetSpritePattern` | MSXgl | bios.h | pendente |  |
| `BIOS_SetSpritePosition` | MSXgl | bios.h | pendente |  |
| `BIOS_SetWidth32` | MSXgl | bios.h | pendente |  |
| `BIOS_SetWidth40` | MSXgl | bios.h | pendente |  |
| `BIOS_SwitchSlot` | MSXgl | bios.h | pendente |  |
| `BIOS_TextPrint` | MSXgl | bios.h | pendente |  |
| `BIOS_TextPrintAt` | MSXgl | bios.h | pendente |  |
| `BIOS_TextPrintChar` | MSXgl | bios.h | pendente |  |
| `BIOS_TextSetCursor` | MSXgl | bios.h | pendente |  |
| `BIOS_WritePSG` | MSXgl | bios.h | pendente |  |
| `BIOS_WriteVDP` | MSXgl | bios.h | pendente |  |
| `BIOS_WriteVRAM` | MSXgl | bios.h | pendente |  |

## DOS

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `AllocateSegment` | Fusion-C | rammapper.h | pendente |  |
| `DosCLS` | Fusion-C | msx_fusion.h | pendente |  |
| `Exit` | Fusion-C | msx_fusion.h | pendente |  |
| `FreeSegment` | Fusion-C | rammapper.h | pendente |  |
| `Get_PN` | Fusion-C | rammapper.h | pendente |  |
| `InitRamMapperInfo` | Fusion-C | rammapper.h | pendente |  |
| `IntBios` | Fusion-C | msx_fusion.h | pendente |  |
| `IntDos` | Fusion-C | msx_fusion.h | pendente |  |
| `Put_PN` | Fusion-C | rammapper.h | pendente |  |
| `_GetRamMapperBaseTable` | Fusion-C | rammapper.h | pendente |  |
| `fcb_close` | Fusion-C | msx_fusion.h | pendente |  |
| `fcb_create` | Fusion-C | msx_fusion.h | pendente |  |
| `fcb_find_first` | Fusion-C | msx_fusion.h | pendente |  |
| `fcb_find_next` | Fusion-C | msx_fusion.h | pendente |  |
| `fcb_open` | Fusion-C | msx_fusion.h | pendente |  |
| `fcb_read` | Fusion-C | msx_fusion.h | pendente |  |
| `fcb_write` | Fusion-C | msx_fusion.h | pendente |  |
| `DOSMapper_Alloc` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_Free` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_FreeStruct` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_GetPage` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_GetPage0` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_GetPage1` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_GetPage2` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_GetPage3` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_GetVarTable` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_Init` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_ReadByte` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_SetPage` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_SetPage0` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_SetPage1` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_SetPage2` | MSXgl | dos_mapper.h | pendente |  |
| `DOSMapper_WriteByte` | MSXgl | dos_mapper.h | pendente |  |
| `DOS_AvailableDrives` | MSXgl | dos.h | pendente |  |
| `DOS_Beep` | MSXgl | dos.h | pendente |  |
| `DOS_Call` | MSXgl | dos.h | pendente |  |
| `DOS_ChangeDirectory` | MSXgl | dos.h | pendente |  |
| `DOS_CharOutput` | MSXgl | dos.h | pendente |  |
| `DOS_ClearScreen` | MSXgl | dos.h | pendente |  |
| `DOS_CloseFCB` | MSXgl | dos.h | pendente |  |
| `DOS_CloseHandle` | MSXgl | dos.h | pendente |  |
| `DOS_CreateFCB` | MSXgl | dos.h | pendente |  |
| `DOS_CreateHandle` | MSXgl | dos.h | pendente |  |
| `DOS_Delete` | MSXgl | dos.h | pendente |  |
| `DOS_DeleteFCB` | MSXgl | dos.h | pendente |  |
| `DOS_DeleteHandle` | MSXgl | dos.h | pendente |  |
| `DOS_DuplicateHandle` | MSXgl | dos.h | pendente |  |
| `DOS_EnsureHandle` | MSXgl | dos.h | pendente |  |
| `DOS_Exit` | MSXgl | dos.h | pendente |  |
| `DOS_Exit0` | MSXgl | dos.h | pendente |  |
| `DOS_Explain` | MSXgl | dos.h | pendente |  |
| `DOS_FindFirstEntry` | MSXgl | dos.h | pendente |  |
| `DOS_FindFirstFileFCB` | MSXgl | dos.h | pendente |  |
| `DOS_FindNextEntry` | MSXgl | dos.h | pendente |  |
| `DOS_FindNextFileFCB` | MSXgl | dos.h | pendente |  |
| `DOS_GetAttribute` | MSXgl | dos.h | pendente |  |
| `DOS_GetAttributeHandle` | MSXgl | dos.h | pendente |  |
| `DOS_GetCurrentDiskDirectory` | MSXgl | dos.h | pendente |  |
| `DOS_GetCurrentDrive` | MSXgl | dos.h | pendente |  |
| `DOS_GetDirectory` | MSXgl | dos.h | pendente |  |
| `DOS_GetDiskInfo` | MSXgl | dos.h | pendente |  |
| `DOS_GetDiskParam` | MSXgl | dos.h | pendente |  |
| `DOS_GetFileDay` | MSXgl | dos.h | pendente |  |
| `DOS_GetFileHour` | MSXgl | dos.h | pendente |  |
| `DOS_GetFileMinute` | MSXgl | dos.h | pendente |  |
| `DOS_GetFileMonth` | MSXgl | dos.h | pendente |  |
| `DOS_GetFileSecond` | MSXgl | dos.h | pendente |  |
| `DOS_GetFileYear` | MSXgl | dos.h | pendente |  |
| `DOS_GetFreeBytes` | MSXgl | dos.h | pendente |  |
| `DOS_GetFreeClusters` | MSXgl | dos.h | pendente |  |
| `DOS_GetFreeSectors` | MSXgl | dos.h | pendente |  |
| `DOS_GetFreeSpace` | MSXgl | dos.h | pendente |  |
| `DOS_GetLastError` | MSXgl | dos.h | pendente |  |
| `DOS_GetLastFileInfo` | MSXgl | dos.h | pendente |  |
| `DOS_GetSizeFCB` | MSXgl | dos.h | pendente |  |
| `DOS_GetTime` | MSXgl | dos.h | pendente |  |
| `DOS_GetTotalBytes` | MSXgl | dos.h | pendente |  |
| `DOS_GetTotalClusters` | MSXgl | dos.h | pendente |  |
| `DOS_GetTotalSectors` | MSXgl | dos.h | pendente |  |
| `DOS_GetVersion` | MSXgl | dos.h | pendente |  |
| `DOS_InstallErrorHandler` | MSXgl | dos.h | pendente |  |
| `DOS_InterSlotCall` | MSXgl | dos.h | pendente |  |
| `DOS_InterSlotRead` | MSXgl | dos.h | pendente |  |
| `DOS_InterSlotWrite` | MSXgl | dos.h | pendente |  |
| `DOS_Move` | MSXgl | dos.h | pendente |  |
| `DOS_MoveHandle` | MSXgl | dos.h | pendente |  |
| `DOS_OpenFCB` | MSXgl | dos.h | pendente |  |
| `DOS_OpenHandle` | MSXgl | dos.h | pendente |  |
| `DOS_RandomBlockReadFCB` | MSXgl | dos.h | pendente |  |
| `DOS_RandomBlockWriteFCB` | MSXgl | dos.h | pendente |  |
| `DOS_ReadHandle` | MSXgl | dos.h | pendente |  |
| `DOS_Rename` | MSXgl | dos.h | pendente |  |
| `DOS_RenameHandle` | MSXgl | dos.h | pendente |  |
| `DOS_ResetLastError` | MSXgl | dos.h | pendente |  |
| `DOS_Return` | MSXgl | dos.h | pendente |  |
| `DOS_SeekHandle` | MSXgl | dos.h | pendente |  |
| `DOS_SelectDrive` | MSXgl | dos.h | pendente |  |
| `DOS_SelectDriveLetter` | MSXgl | dos.h | pendente |  |
| `DOS_SequentialReadFCB` | MSXgl | dos.h | pendente |  |
| `DOS_SequentialWriteFCB` | MSXgl | dos.h | pendente |  |
| `DOS_SetAttribute` | MSXgl | dos.h | pendente |  |
| `DOS_SetAttributeHandle` | MSXgl | dos.h | pendente |  |
| `DOS_SetTransferAddr` | MSXgl | dos.h | pendente |  |
| `DOS_StringOutput` | MSXgl | dos.h | pendente |  |
| `DOS_WriteHandle` | MSXgl | dos.h | pendente |  |

## System

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `ChangeCPU` | Fusion-C | msx_fusion.h | pendente |  |
| `ChangeDir` | Fusion-C | io.h | pendente |  |
| `Close` | Fusion-C | io.h | pendente |  |
| `Create` | Fusion-C | io.h | pendente |  |
| `CreateAttrib` | Fusion-C | io.h | pendente |  |
| `DiskLoad` | Fusion-C | io.h | pendente |  |
| `EndInterruptHandler` | Fusion-C | msx_fusion.h | pendente |  |
| `FCBs` | Fusion-C | io.h | pendente |  |
| `FindFirst` | Fusion-C | io.h | pendente |  |
| `FindNext` | Fusion-C | io.h | pendente |  |
| `GetCPU` | Fusion-C | msx_fusion.h | pendente |  |
| `GetCWD` | Fusion-C | io.h | pendente |  |
| `GetDisk` | Fusion-C | io.h | pendente |  |
| `GetDiskParam` | Fusion-C | io.h | pendente |  |
| `GetDiskTrAddress` | Fusion-C | io.h | pendente |  |
| `GetOSVersion` | Fusion-C | io.h | pendente |  |
| `InPort` | Fusion-C | msx_fusion.h | pendente |  |
| `InitInterruptHandler` | Fusion-C | msx_fusion.h | pendente |  |
| `Lseek` | Fusion-C | io.h | pendente |  |
| `Ltell` | Fusion-C | io.h | pendente |  |
| `MakeDir` | Fusion-C | io.h | pendente |  |
| `Open` | Fusion-C | io.h | pendente |  |
| `OpenAttrib` | Fusion-C | io.h | pendente |  |
| `OutPort` | Fusion-C | msx_fusion.h | pendente |  |
| `OutPorts` | Fusion-C | msx_fusion.h | pendente |  |
| `PutText` | Fusion-C | msx_fusion.h | pendente |  |
| `Read` | Fusion-C | io.h | pendente |  |
| `ReadMSXtype` | Fusion-C | msx_fusion.h | pendente |  |
| `ReadSP` | Fusion-C | msx_fusion.h | feito | MEM_GetSP |
| `ReadTPA` | Fusion-C | msx_fusion.h | feito | MEM_TPATop |
| `Remove` | Fusion-C | io.h | pendente |  |
| `RemoveDir` | Fusion-C | io.h | pendente |  |
| `Rename` | Fusion-C | io.h | pendente |  |
| `SectorRead` | Fusion-C | io.h | pendente |  |
| `SectorWrite` | Fusion-C | io.h | pendente |  |
| `SetDisk` | Fusion-C | io.h | pendente |  |
| `SetDiskTrAddress` | Fusion-C | io.h | pendente |  |
| `SetInterruptHandler` | Fusion-C | msx_fusion.h | pendente |  |
| `Suspend` | Fusion-C | msx_fusion.h | pendente |  |
| `Write` | Fusion-C | io.h | pendente |  |
| `_REGs` | Fusion-C | msx_fusion.h | pendente |  |
| `_seek` | Fusion-C | io.h | pendente |  |
| `_size` | Fusion-C | io.h | pendente |  |
| `_tell` | Fusion-C | io.h | pendente |  |
| `Basic_GetByte` | MSXgl | basic_usr.h | pendente |  |
| `Basic_GetFloat` | MSXgl | basic_usr.h | pendente |  |
| `Basic_GetStringLength` | MSXgl | basic_usr.h | pendente |  |
| `Basic_GetType` | MSXgl | basic_usr.h | pendente |  |
| `Basic_GetWord` | MSXgl | basic_usr.h | pendente |  |
| `Basic_SetByte` | MSXgl | basic_usr.h | pendente |  |
| `Basic_SetFloat` | MSXgl | basic_usr.h | pendente |  |
| `Basic_SetString` | MSXgl | basic_usr.h | pendente |  |
| `Basic_SetWord` | MSXgl | basic_usr.h | pendente |  |
| `Call` | MSXgl | system.h | pendente |  |
| `CallA` | MSXgl | system.h | pendente |  |
| `CallAToA` | MSXgl | system.h | pendente |  |
| `CallDriver` | MSXgl | system.h | pendente |  |
| `CallHL` | MSXgl | system.h | pendente |  |
| `CallHLToA` | MSXgl | system.h | pendente |  |
| `CallL` | MSXgl | system.h | pendente |  |
| `CallLToA` | MSXgl | system.h | pendente |  |
| `CallToA` | MSXgl | system.h | pendente |  |
| `DisableInterrupt` | MSXgl | system.h | pendente |  |
| `EnableInterrupt` | MSXgl | system.h | pendente |  |
| `Halt` | MSXgl | system.h | pendente |  |
| `Peek` | MSXgl | system.h | pendente |  |
| `Peek16` | MSXgl | system.h | pendente |  |
| `Poke` | MSXgl | system.h | pendente |  |
| `Poke16` | MSXgl | system.h | pendente |  |
| `Sys_CheckSlot` | MSXgl | system.h | pendente |  |
| `Sys_GetBIOSInfo` | MSXgl | system.h | pendente |  |
| `Sys_GetFirstAddr` | MSXgl | system.h | pendente |  |
| `Sys_GetFrequency` | MSXgl | system.h | pendente |  |
| `Sys_GetHeaderAddr` | MSXgl | system.h | pendente |  |
| `Sys_GetLastAddr` | MSXgl | system.h | pendente |  |
| `Sys_GetMSXVersion` | MSXgl | system.h | pendente |  |
| `Sys_GetPageSlot` | MSXgl | system.h | pendente |  |
| `Sys_Is50Hz` | MSXgl | system.h | pendente |  |
| `Sys_Is60Hz` | MSXgl | system.h | pendente |  |
| `Sys_IsSlotExpanded` | MSXgl | system.h | pendente |  |
| `Sys_PlayClickSound` | MSXgl | system.h | pendente |  |
| `Sys_SetPage0Slot` | MSXgl | system.h | pendente |  |
| `Sys_SetPageSlot` | MSXgl | system.h | pendente |  |
| `Sys_SlotGetPrimary` | MSXgl | system.h | pendente |  |
| `Sys_SlotGetSecondary` | MSXgl | system.h | pendente |  |
| `Sys_SlotIsExpended` | MSXgl | system.h | pendente |  |
| `Sys_StopClickSound` | MSXgl | system.h | pendente |  |

## Clock

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `GetDate` | Fusion-C | msx_fusion.h | pendente |  |
| `GetTime` | Fusion-C | msx_fusion.h | pendente |  |
| `RealTimer` | Fusion-C | msx_fusion.h | pendente |  |
| `SetDate` | Fusion-C | msx_fusion.h | pendente |  |
| `SetRealTimer` | Fusion-C | msx_fusion.h | pendente |  |
| `SetTime` | Fusion-C | msx_fusion.h | pendente |  |
| `RTC_GetAreaCode` | MSXgl | clock.h | pendente |  |
| `RTC_GetDataType` | MSXgl | clock.h | pendente |  |
| `RTC_GetDay` | MSXgl | clock.h | pendente |  |
| `RTC_GetDayOfWeek` | MSXgl | clock.h | pendente |  |
| `RTC_GetHour` | MSXgl | clock.h | pendente |  |
| `RTC_GetMinute` | MSXgl | clock.h | pendente |  |
| `RTC_GetMonth` | MSXgl | clock.h | pendente |  |
| `RTC_GetSecond` | MSXgl | clock.h | pendente |  |
| `RTC_GetYear` | MSXgl | clock.h | pendente |  |
| `RTC_GetYear4` | MSXgl | clock.h | pendente |  |
| `RTC_Initialize` | MSXgl | clock.h | pendente |  |
| `RTC_IsPM` | MSXgl | clock.h | pendente |  |
| `RTC_IsSettingOK` | MSXgl | clock.h | pendente |  |
| `RTC_LoadData` | MSXgl | clock.h | pendente |  |
| `RTC_LoadDataSigned` | MSXgl | clock.h | pendente |  |
| `RTC_Read` | MSXgl | clock.h | pendente |  |
| `RTC_ReadRaw` | MSXgl | clock.h | pendente |  |
| `RTC_SaveData` | MSXgl | clock.h | pendente |  |
| `RTC_SaveDataSigned` | MSXgl | clock.h | pendente |  |
| `RTC_Set24H` | MSXgl | clock.h | pendente |  |
| `RTC_SetAreaCode` | MSXgl | clock.h | pendente |  |
| `RTC_SetMode` | MSXgl | clock.h | pendente |  |
| `RTC_Write` | MSXgl | clock.h | pendente |  |
| `RTC_WriteRaw` | MSXgl | clock.h | pendente |  |

## V9990

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `G9Close` | Fusion-C | g9klib.h | pendente |  |
| `G9CmdWait` | Fusion-C | g9klib.h | pendente |  |
| `G9CopyRamToVram` | Fusion-C | g9klib.h | pendente |  |
| `G9CopyRamToXY` | Fusion-C | g9klib.h | pendente |  |
| `G9CopyXYToRam` | Fusion-C | g9klib.h | pendente |  |
| `G9CopyXYToXY` | Fusion-C | g9klib.h | pendente |  |
| `G9Detect` | Fusion-C | g9klib.h | pendente |  |
| `G9DisableAllSprites` | Fusion-C | g9klib.h | pendente |  |
| `G9DisableLayerA` | Fusion-C | g9klib.h | pendente |  |
| `G9DisableLayerB` | Fusion-C | g9klib.h | pendente |  |
| `G9DisplayDisable` | Fusion-C | g9klib.h | pendente |  |
| `G9DisplayEnable` | Fusion-C | g9klib.h | pendente |  |
| `G9DrawFilledBox` | Fusion-C | g9klib.h | pendente |  |
| `G9DrawLine` | Fusion-C | g9klib.h | pendente |  |
| `G9EnableLayerA` | Fusion-C | g9klib.h | pendente |  |
| `G9EnableLayerB` | Fusion-C | g9klib.h | pendente |  |
| `G9GetPattern` | Fusion-C | g9klib.h | pendente |  |
| `G9GetPatternData` | Fusion-C | g9klib.h | pendente |  |
| `G9GetPoint` | Fusion-C | g9klib.h | pendente |  |
| `G9InitPalette` | Fusion-C | g9klib.h | pendente |  |
| `G9InitPatternMode` | Fusion-C | g9klib.h | pendente |  |
| `G9InitSpritePattern` | Fusion-C | g9klib.h | pendente |  |
| `G9LoadFont` | Fusion-C | g9klib.h | pendente |  |
| `G9Locate` | Fusion-C | g9klib.h | pendente |  |
| `G9OpenG9B` | Fusion-C | g9klib.h | pendente |  |
| `G9OpenVff` | Fusion-C | g9klib.h | pendente |  |
| `G9PCopyVramToVram` | Fusion-C | g9klib.h | pendente |  |
| `G9PCopyVramToVramARegister` | Fusion-C | g9klib.h | pendente |  |
| `G9PrintStringVram` | Fusion-C | g9klib.h | pendente |  |
| `G9PrintTilesA` | Fusion-C | g9klib.h | pendente |  |
| `G9PrintTilesB` | Fusion-C | g9klib.h | pendente |  |
| `G9ReadG9B` | Fusion-C | g9klib.h | pendente |  |
| `G9ReadPalette` | Fusion-C | g9klib.h | pendente |  |
| `G9Reset` | Fusion-C | g9klib.h | pendente |  |
| `G9SetAdjust` | Fusion-C | g9klib.h | pendente |  |
| `G9SetBackDropColor` | Fusion-C | g9klib.h | pendente |  |
| `G9SetCmdBackColor` | Fusion-C | g9klib.h | pendente |  |
| `G9SetCmdColor` | Fusion-C | g9klib.h | pendente |  |
| `G9SetCmdWriteMask` | Fusion-C | g9klib.h | pendente |  |
| `G9SetFont` | Fusion-C | g9klib.h | pendente |  |
| `G9SetIntLine` | Fusion-C | g9klib.h | pendente |  |
| `G9SetPattern` | Fusion-C | g9klib.h | pendente |  |
| `G9SetPatternData` | Fusion-C | g9klib.h | pendente |  |
| `G9SetPoint` | Fusion-C | g9klib.h | pendente |  |
| `G9SetScreenMode` | Fusion-C | g9klib.h | pendente |  |
| `G9SetScrollMode` | Fusion-C | g9klib.h | pendente |  |
| `G9SetScrollX` | Fusion-C | g9klib.h | pendente |  |
| `G9SetScrollXB` | Fusion-C | g9klib.h | pendente |  |
| `G9SetScrollY` | Fusion-C | g9klib.h | pendente |  |
| `G9SetScrollYB` | Fusion-C | g9klib.h | pendente |  |
| `G9SetSprite` | Fusion-C | g9klib.h | pendente |  |
| `G9SetVramRead` | Fusion-C | g9klib.h | pendente |  |
| `G9SetVramWrite` | Fusion-C | g9klib.h | pendente |  |
| `G9SetupCopyRamToXY` | Fusion-C | g9klib.h | pendente |  |
| `G9SetupCopyXYToRam` | Fusion-C | g9klib.h | pendente |  |
| `G9SpritesDisable` | Fusion-C | g9klib.h | pendente |  |
| `G9SpritesEnable` | Fusion-C | g9klib.h | pendente |  |
| `G9WaitVsync` | Fusion-C | g9klib.h | pendente |  |
| `G9WritePalette` | Fusion-C | g9klib.h | pendente |  |
| `G9WriteReg` | Fusion-C | g9klib.h | pendente |  |
| `GetCommandBX` | MSXgl | v9990.h | pendente |  |
| `V9_ClearVRAM` | MSXgl | v9990.h | pendente |  |
| `V9_CommandADVANCE` | MSXgl | v9990.h | pendente |  |
| `V9_CommandBMLL` | MSXgl | v9990.h | pendente |  |
| `V9_CommandBMLX` | MSXgl | v9990.h | pendente |  |
| `V9_CommandBMXL` | MSXgl | v9990.h | pendente |  |
| `V9_CommandCMMC` | MSXgl | v9990.h | pendente |  |
| `V9_CommandCMMM` | MSXgl | v9990.h | pendente |  |
| `V9_CommandLINE` | MSXgl | v9990.h | pendente |  |
| `V9_CommandLMMM` | MSXgl | v9990.h | pendente |  |
| `V9_CommandLMMV` | MSXgl | v9990.h | pendente |  |
| `V9_CommandPOINT` | MSXgl | v9990.h | pendente |  |
| `V9_CommandPSET` | MSXgl | v9990.h | pendente |  |
| `V9_CommandSEARCH` | MSXgl | v9990.h | pendente |  |
| `V9_CommandSTOP` | MSXgl | v9990.h | pendente |  |
| `V9_Detect` | MSXgl | v9990.h | pendente |  |
| `V9_DisableInterrupt` | MSXgl | v9990.h | pendente |  |
| `V9_ExecCommand` | MSXgl | v9990.h | pendente |  |
| `V9_FillVRAM` | MSXgl | v9990.h | pendente |  |
| `V9_FillVRAM16` | MSXgl | v9990.h | pendente |  |
| `V9_FillVRAM16_CurrentAddr` | MSXgl | v9990.h | pendente |  |
| `V9_FillVRAM_CurrentAddr` | MSXgl | v9990.h | pendente |  |
| `V9_GetBPP` | MSXgl | v9990.h | pendente |  |
| `V9_GetBackgroundColor` | MSXgl | v9990.h | pendente |  |
| `V9_GetImageSpaceWidth` | MSXgl | v9990.h | pendente |  |
| `V9_GetPort` | MSXgl | v9990.h | pendente |  |
| `V9_GetRegister` | MSXgl | v9990.h | pendente |  |
| `V9_GetStatus` | MSXgl | v9990.h | pendente |  |
| `V9_IsCmdComplete` | MSXgl | v9990.h | pendente |  |
| `V9_IsCmdDataReady` | MSXgl | v9990.h | pendente |  |
| `V9_IsCmdRunning` | MSXgl | v9990.h | pendente |  |
| `V9_IsHBlank` | MSXgl | v9990.h | pendente |  |
| `V9_IsSecondField` | MSXgl | v9990.h | pendente |  |
| `V9_IsVBlank` | MSXgl | v9990.h | pendente |  |
| `V9_Peek` | MSXgl | v9990.h | pendente |  |
| `V9_Peek16` | MSXgl | v9990.h | pendente |  |
| `V9_Peek16_CurrentAddr` | MSXgl | v9990.h | pendente |  |
| `V9_Peek_CurrentAddr` | MSXgl | v9990.h | pendente |  |
| `V9_Poke` | MSXgl | v9990.h | pendente |  |
| `V9_Poke16` | MSXgl | v9990.h | pendente |  |
| `V9_Poke16_CurrentAddr` | MSXgl | v9990.h | pendente |  |
| `V9_Poke_CurrentAddr` | MSXgl | v9990.h | pendente |  |
| `V9_ReadVRAM` | MSXgl | v9990.h | pendente |  |
| `V9_ReadVRAM_CurrentAddr` | MSXgl | v9990.h | pendente |  |
| `V9_SelectPaletteBP2` | MSXgl | v9990.h | pendente |  |
| `V9_SelectPaletteBP4` | MSXgl | v9990.h | pendente |  |
| `V9_SelectPaletteP1` | MSXgl | v9990.h | pendente |  |
| `V9_SelectPaletteP2` | MSXgl | v9990.h | pendente |  |
| `V9_SetAdjustOffset` | MSXgl | v9990.h | pendente |  |
| `V9_SetAdjustOffsetXY` | MSXgl | v9990.h | pendente |  |
| `V9_SetBPP` | MSXgl | v9990.h | pendente |  |
| `V9_SetBackgroundColor` | MSXgl | v9990.h | pendente |  |
| `V9_SetCmdEndInterrupt` | MSXgl | v9990.h | pendente |  |
| `V9_SetColorMode` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandArgument` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandBC` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandDA` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandDX` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandDY` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandFC` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandLogicalOp` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandMI` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandMJ` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandNA` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandNX` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandNY` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandSA` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandSX` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandSY` | MSXgl | v9990.h | pendente |  |
| `V9_SetCommandWriteMask` | MSXgl | v9990.h | pendente |  |
| `V9_SetCursorAttribute` | MSXgl | v9990.h | pendente |  |
| `V9_SetCursorDisplay` | MSXgl | v9990.h | pendente |  |
| `V9_SetCursorEnable` | MSXgl | v9990.h | pendente |  |
| `V9_SetCursorPalette` | MSXgl | v9990.h | pendente |  |
| `V9_SetCursorPattern` | MSXgl | v9990.h | pendente |  |
| `V9_SetDisplayEnable` | MSXgl | v9990.h | pendente |  |
| `V9_SetFlag` | MSXgl | v9990.h | pendente |  |
| `V9_SetHBlankInterrupt` | MSXgl | v9990.h | pendente |  |
| `V9_SetImageSpaceWidth` | MSXgl | v9990.h | pendente |  |
| `V9_SetInterrupt` | MSXgl | v9990.h | pendente |  |
| `V9_SetInterruptEveryLine` | MSXgl | v9990.h | pendente |  |
| `V9_SetInterruptLine` | MSXgl | v9990.h | pendente |  |
| `V9_SetInterruptX` | MSXgl | v9990.h | pendente |  |
| `V9_SetLayerPriority` | MSXgl | v9990.h | pendente |  |
| `V9_SetPalette` | MSXgl | v9990.h | pendente |  |
| `V9_SetPaletteAll` | MSXgl | v9990.h | pendente |  |
| `V9_SetPaletteEntry` | MSXgl | v9990.h | pendente |  |
| `V9_SetPort` | MSXgl | v9990.h | pendente |  |
| `V9_SetReadAddress` | MSXgl | v9990.h | pendente |  |
| `V9_SetRegister` | MSXgl | v9990.h | pendente |  |
| `V9_SetRegister16` | MSXgl | v9990.h | pendente |  |
| `V9_SetScreenMode` | MSXgl | v9990.h | pendente |  |
| `V9_SetScrolling` | MSXgl | v9990.h | pendente |  |
| `V9_SetScrollingB` | MSXgl | v9990.h | pendente |  |
| `V9_SetScrollingBX` | MSXgl | v9990.h | pendente |  |
| `V9_SetScrollingBY` | MSXgl | v9990.h | pendente |  |
| `V9_SetScrollingX` | MSXgl | v9990.h | pendente |  |
| `V9_SetScrollingY` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpriteDisableP1` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpriteDisableP2` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpriteEnable` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpriteEnableP1` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpriteEnableP2` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpriteInfoP1` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpriteInfoP2` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpriteP1` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpriteP2` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpritePaletteOffset` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpritePaletteP1` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpritePaletteP2` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpritePatternAddr` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpritePatternP1` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpritePatternP2` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpritePositionP1` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpritePositionP2` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpritePriorityP1` | MSXgl | v9990.h | pendente |  |
| `V9_SetSpritePriorityP2` | MSXgl | v9990.h | pendente |  |
| `V9_SetVBlankInterrupt` | MSXgl | v9990.h | pendente |  |
| `V9_SetWriteAddress` | MSXgl | v9990.h | pendente |  |
| `V9_TileAddrP1A` | MSXgl | v9990.h | pendente |  |
| `V9_TileAddrP1B` | MSXgl | v9990.h | pendente |  |
| `V9_TileAddrP2` | MSXgl | v9990.h | pendente |  |
| `V9_WaitCmdEnd` | MSXgl | v9990.h | pendente |  |
| `V9_WriteVRAM` | MSXgl | v9990.h | pendente |  |
| `V9_WriteVRAM_CurrentAddr` | MSXgl | v9990.h | pendente |  |

## Fora da lista (por enquanto)

| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |
| --- | --- | --- | --- | --- |
| `RleWBToRam` | Fusion-C | msx_fusion.h | pendente |  |
| `RleWBToVram` | Fusion-C | msx_fusion.h | pendente |  |
| `tcpip_config_autoip_get` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_config_autoip_set` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_config_ip` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_config_ping_get` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_config_ping_set` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_config_ttltos_get` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_config_ttltos_set` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_dns_q` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_dns_s` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_enumerate` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_get_capab_connections` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_get_capab_dtg_sizes` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_get_capab_flags_llproto` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_get_ipinfo` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_ipraw_close` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_ipraw_open` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_ipraw_state` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_net_state` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_rcv_echo` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_send_echo` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_tcp_abort` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_tcp_close` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_tcp_flush` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_tcp_open` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_tcp_send` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_tcp_state` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_udp_close` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_udp_open` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `tcpip_udp_state` | Fusion-C | gr8net-tcpip.h | pendente |  |
| `BitField_Disable` | MSXgl | game/bitfield.h | pendente |  |
| `BitField_Enable` | MSXgl | game/bitfield.h | pendente |  |
| `BitField_Get` | MSXgl | game/bitfield.h | pendente |  |
| `BitField_GetBitNumber` | MSXgl | game/bitfield.h | pendente |  |
| `BitField_GetData` | MSXgl | game/bitfield.h | pendente |  |
| `BitField_GetSize` | MSXgl | game/bitfield.h | pendente |  |
| `BitField_Initialize` | MSXgl | game/bitfield.h | pendente |  |
| `BitField_Set` | MSXgl | game/bitfield.h | pendente |  |
| `BitField_Toggle` | MSXgl | game/bitfield.h | pendente |  |
| `Bitbuster2_UnpackToRAM` | MSXgl | compress/bitbuster2.h | pendente |  |
| `Bitbuster_UnpackToRAM` | MSXgl | compress/bitbuster.h | pendente |  |
| `Bitbuster_UnpackToVRAM` | MSXgl | compress/bitbuster.h | pendente |  |
| `Crypt_Decode` | MSXgl | crypt.h | pendente |  |
| `Crypt_Encode` | MSXgl | crypt.h | pendente |  |
| `Crypt_SetCode` | MSXgl | crypt.h | pendente |  |
| `Crypt_SetKey` | MSXgl | crypt.h | pendente |  |
| `Crypt_SetMap` | MSXgl | crypt.h | pendente |  |
| `FSM_GetPending` | MSXgl | fsm.h | pendente |  |
| `FSM_GetPrevious` | MSXgl | fsm.h | pendente |  |
| `FSM_GetState` | MSXgl | fsm.h | pendente |  |
| `FSM_IsPending` | MSXgl | fsm.h | pendente |  |
| `FSM_SetPrevious` | MSXgl | fsm.h | pendente |  |
| `FSM_SetState` | MSXgl | fsm.h | pendente |  |
| `FSM_Update` | MSXgl | fsm.h | pendente |  |
| `GamePawn_Disable` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_Draw` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_Enable` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_GetCallbackCellX` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_GetCallbackCellY` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_GetPhysicsState` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_Initialize` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_InitializePhysics` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_SetAction` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_SetEnable` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_SetMovement` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_SetPosition` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_SetTargetPosition` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_SetTileMap` | MSXgl | game_pawn.h | pendente |  |
| `GamePawn_Update` | MSXgl | game_pawn.h | pendente |  |
| `Game_Exit` | MSXgl | game/state.h | pendente |  |
| `Game_GetCurrentState` | MSXgl | game/state.h | pendente |  |
| `Game_GetFrameCount` | MSXgl | game/state.h | pendente |  |
| `Game_Initialize` | MSXgl | game/state.h | pendente |  |
| `Game_Release` | MSXgl | game/state.h | pendente |  |
| `Game_RestoreState` | MSXgl | game/state.h | pendente |  |
| `Game_SetIs60Hz` | MSXgl | game/state.h | pendente |  |
| `Game_SetState` | MSXgl | game/state.h | pendente |  |
| `Game_SetVSyncCallback` | MSXgl | game/state.h | pendente |  |
| `Game_Start` | MSXgl | game/state.h | pendente |  |
| `Game_Update` | MSXgl | game/state.h | pendente |  |
| `Game_UpdateState` | MSXgl | game/state.h | pendente |  |
| `Game_VSyncHook` | MSXgl | game/state.h | pendente |  |
| `Game_WaitVSync` | MSXgl | game/state.h | pendente |  |
| `HID_Detect` | MSXgl | device/msx-hid.h | pendente |  |
| `JSXC_Detect` | MSXgl | device/jsx.h | pendente |  |
| `JSXC_Read` | MSXgl | device/jsx.h | pendente |  |
| `JSX_Detect` | MSXgl | device/jsx.h | pendente |  |
| `JSX_GetAxisNumber` | MSXgl | device/jsx.h | pendente |  |
| `JSX_GetButtonsNumber` | MSXgl | device/jsx.h | pendente |  |
| `JSX_GetRowsNumber` | MSXgl | device/jsx.h | pendente |  |
| `JSX_Read` | MSXgl | device/jsx.h | pendente |  |
| `JoyMega_IsPressedA` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedB` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedC` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedDown` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedLeft` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedMode` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedRight` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedStart` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedUp` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedX` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedY` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_IsPressedZ` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_Read3` | MSXgl | device/joymega.h | pendente |  |
| `JoyMega_Read6` | MSXgl | device/joymega.h | pendente |  |
| `LZ48_UnpackToRAM` | MSXgl | compress/lz48.h | pendente |  |
| `LightGun_ASCII_GetLight` | MSXgl | device/lightgun.h | pendente |  |
| `LightGun_ASCII_GetTrigger` | MSXgl | device/lightgun.h | pendente |  |
| `LightGun_GunStick_GetLight` | MSXgl | device/lightgun.h | pendente |  |
| `LightGun_GunStick_GetTrigger` | MSXgl | device/lightgun.h | pendente |  |
| `LightGun_Phenix_GetLight` | MSXgl | device/lightgun.h | pendente |  |
| `LightGun_Phenix_GetTriggerA` | MSXgl | device/lightgun.h | pendente |  |
| `LightGun_Phenix_GetTriggerB` | MSXgl | device/lightgun.h | pendente |  |
| `LightGun_Phenix_Init` | MSXgl | device/lightgun.h | pendente |  |
| `LightGun_Phenix_Update` | MSXgl | device/lightgun.h | pendente |  |
| `LightGun_Read` | MSXgl | device/lightgun.h | pendente |  |
| `Lightgun_Phenix_GetCounter` | MSXgl | device/lightgun.h | pendente |  |
| `Lightgun_Phenix_GetRaw` | MSXgl | device/lightgun.h | pendente |  |
| `Lightgun_Phenix_GetState` | MSXgl | device/lightgun.h | pendente |  |
| `Loc_Initialize` | MSXgl | localize.h | pendente |  |
| `Loc_SetLanguage` | MSXgl | localize.h | pendente |  |
| `MSXi_UnpackToVRAM_Crop16_4_4` | MSXgl | msxi/msxi_unpack.h | pendente |  |
| `MSXi_UnpackToVRAM_Crop256_4_4` | MSXgl | msxi/msxi_unpack.h | pendente |  |
| `MSXi_UnpackToVRAM_Crop32_4_4` | MSXgl | msxi/msxi_unpack.h | pendente |  |
| `MSXi_UnpackToVRAM_CropLine16_4_4` | MSXgl | msxi/msxi_unpack.h | pendente |  |
| `MSXi_UnpackToVRAM_CropLine256_4_4` | MSXgl | msxi/msxi_unpack.h | pendente |  |
| `MSXi_UnpackToVRAM_CropLine32_4_4` | MSXgl | msxi/msxi_unpack.h | pendente |  |
| `MSXi_UnpackToVRAM_None_4_4` | MSXgl | msxi/msxi_unpack.h | pendente |  |
| `MSXi_UnpackToVRAM_RLE0_4_4` | MSXgl | msxi/msxi_unpack.h | pendente |  |
| `MSXi_UnpackToVRAM_RLE4_4_4` | MSXgl | msxi/msxi_unpack.h | pendente |  |
| `MSXi_UnpackToVRAM_RLE8_4_4` | MSXgl | msxi/msxi_unpack.h | pendente |  |
| `Menu_Display` | MSXgl | game/menu.h | pendente |  |
| `Menu_GetCurrentItem` | MSXgl | game/menu.h | pendente |  |
| `Menu_GetScreenWidth` | MSXgl | game/menu.h | pendente |  |
| `Menu_Initialize` | MSXgl | game/menu.h | pendente |  |
| `Menu_SetDirty` | MSXgl | game/menu.h | pendente |  |
| `Menu_SetDrawCallback` | MSXgl | game/menu.h | pendente |  |
| `Menu_SetEventCallback` | MSXgl | game/menu.h | pendente |  |
| `Menu_SetInputCallback` | MSXgl | game/menu.h | pendente |  |
| `Menu_SetScreenWidth` | MSXgl | game/menu.h | pendente |  |
| `Menu_Update` | MSXgl | game/menu.h | pendente |  |
| `NTap_Check` | MSXgl | device/ninjatap.h | pendente |  |
| `NTap_CheckDM` | MSXgl | device/ninjatap.h | pendente |  |
| `NTap_CheckMGL` | MSXgl | device/ninjatap.h | pendente |  |
| `NTap_CheckST` | MSXgl | device/ninjatap.h | pendente |  |
| `NTap_GetData` | MSXgl | device/ninjatap.h | pendente |  |
| `NTap_GetInfo` | MSXgl | device/ninjatap.h | pendente |  |
| `NTap_GetPortNum` | MSXgl | device/ninjatap.h | pendente |  |
| `NTap_IsPressed` | MSXgl | device/ninjatap.h | pendente |  |
| `NTap_IsPushed` | MSXgl | device/ninjatap.h | pendente |  |
| `NTap_IsReleased` | MSXgl | device/ninjatap.h | pendente |  |
| `NTap_Update` | MSXgl | device/ninjatap.h | pendente |  |
| `ONET_Activate` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_Activation` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_Disactivate` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_DiscardPacket` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_GetMACAddress` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_GetMulticastMask` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_GetPacket` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_GetPacketInfo` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_GetPhysicalAddr` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_GetReceptConfig` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_GetSendStatus` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_GetSlot` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_GetStatus` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_GetVersion` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_HasBIOS` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_Initialize` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_IsActivate` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_ReceptConfig` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_Reset` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_SendPacketAsync` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_SendPacketSync` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_SetMulticastMask` | MSXgl | network/obsonet.h | pendente |  |
| `ONET_SetReceptConfig` | MSXgl | network/obsonet.h | pendente |  |
| `PAC_Activate` | MSXgl | device/pac.h | pendente |  |
| `PAC_Check` | MSXgl | device/pac.h | pendente |  |
| `PAC_Format` | MSXgl | device/pac.h | pendente |  |
| `PAC_FormatAll` | MSXgl | device/pac.h | pendente |  |
| `PAC_GetDefaultSlot` | MSXgl | device/pac.h | pendente |  |
| `PAC_GetNumber` | MSXgl | device/pac.h | pendente |  |
| `PAC_GetSlot` | MSXgl | device/pac.h | pendente |  |
| `PAC_Initialize` | MSXgl | device/pac.h | pendente |  |
| `PAC_Read` | MSXgl | device/pac.h | pendente |  |
| `PAC_Select` | MSXgl | device/pac.h | pendente |  |
| `PAC_Write` | MSXgl | device/pac.h | pendente |  |
| `Paddle_GetAngle` | MSXgl | device/paddle.h | pendente |  |
| `Paddle_GetCalibratedAngle` | MSXgl | device/paddle.h | pendente |  |
| `Paddle_IsButtonPressed` | MSXgl | device/paddle.h | pendente |  |
| `Paddle_IsConnected` | MSXgl | device/paddle.h | pendente |  |
| `Paddle_SetCalibration` | MSXgl | device/paddle.h | pendente |  |
| `Paddle_Update` | MSXgl | device/paddle.h | pendente |  |
| `Pawn_Disable` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_Draw` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_Enable` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_ForceSetAction` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_GetAction` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_GetCallbackCellX` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_GetCallbackCellY` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_GetPhysicsState` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_GetPositionX` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_GetPositionY` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_Initialize` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_InitializePhysics` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_RestartAction` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_SetAction` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_SetColorBlend` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_SetDirty` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_SetEnable` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_SetMovement` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_SetPatternAddress` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_SetPosition` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_SetSpriteFX` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_SetTargetPosition` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_SetTileMap` | MSXgl | game/pawn.h | pendente |  |
| `Pawn_Update` | MSXgl | game/pawn.h | pendente |  |
| `Pletter_UnpackToRAM` | MSXgl | compress/pletter.h | pendente |  |
| `Pletter_UnpackToVRAM` | MSXgl | compress/pletter.h | pendente |  |
| `Printer_ChangePage` | MSXgl | device/printer.h | pendente |  |
| `Printer_CheckReady` | MSXgl | device/printer.h | pendente |  |
| `Printer_LineReturn` | MSXgl | device/printer.h | pendente |  |
| `Printer_SendChar` | MSXgl | device/printer.h | pendente |  |
| `Printer_SendString` | MSXgl | device/printer.h | pendente |  |
| `RLEp_UnpackToRAM` | MSXgl | compress.h | pendente |  |
| `Sequence_CheckArea` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_CheckInput` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_ClearCustomCursor` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_GetCurrent` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_GetCursorX` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_GetCursorY` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_GetCustomCursor` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_GetFrame` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_GetHoverAction` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_GetInput` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_HasCustomCursor` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_HideCursor` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_Initialize` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_InitializeTimeline` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_Interrupt` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_MoveCursor` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_Play` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_PlayPanTransition` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetActionCallback` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetActions` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetCursor` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetCustomCursor` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetDirty` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetDrawCallback` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetEventCallback` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetFinished` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetHoverAction` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetNextFrame` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_SetSynchFrames` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_ShowCursor` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_Update` | MSXgl | game/sequence.h | pendente |  |
| `Sequence_Wait` | MSXgl | game/sequence.h | pendente |  |
| `SpriteFX_CropBottom16` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_CropBottom8` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_CropLeft16` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_CropLeft8` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_CropRight16` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_CropRight8` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_CropTop16` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_CropTop8` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_FlipHorizontal16` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_FlipHorizontal8` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_FlipVertical16` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_FlipVertical8` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_Mask16` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_Mask8` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_RotateHalfTurn16` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_RotateHalfTurn8` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_RotateLeft16` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_RotateLeft8` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_RotateRight16` | MSXgl | sprite_fx.h | pendente |  |
| `SpriteFX_RotateRight8` | MSXgl | sprite_fx.h | pendente |  |
| `WaveGame_Continue` | MSXgl | device/wavegame.h | pendente |  |
| `WaveGame_Fade` | MSXgl | device/wavegame.h | pendente |  |
| `WaveGame_GetVersion` | MSXgl | device/wavegame.h | pendente |  |
| `WaveGame_Pause` | MSXgl | device/wavegame.h | pendente |  |
| `WaveGame_Play` | MSXgl | device/wavegame.h | pendente |  |
| `WaveGame_PlayLoop` | MSXgl | device/wavegame.h | pendente |  |
| `WaveGame_PlayNext` | MSXgl | device/wavegame.h | pendente |  |
| `WaveGame_PlayOnce` | MSXgl | device/wavegame.h | pendente |  |
| `WaveGame_PlayPause` | MSXgl | device/wavegame.h | pendente |  |
| `WaveGame_Stop` | MSXgl | device/wavegame.h | pendente |  |
| `WaveGame_StoreNext` | MSXgl | device/wavegame.h | pendente |  |
| `ZX0_UnpackToRAM` | MSXgl | compress/zx0.h | pendente |  |
| `tcpip_config_autoip_get` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_config_autoip_set` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_config_ip` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_config_ping_get` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_config_ping_set` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_config_ttltos_get` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_config_ttltos_set` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_dns_q` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_dns_s` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_enumerate` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_get_capab_connections` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_get_capab_dtg_sizes` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_get_capab_flags_llproto` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_get_ipinfo` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_ipraw_close` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_ipraw_open` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_ipraw_rcv` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_ipraw_send` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_ipraw_state` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_net_state` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_rcv_echo` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_send_echo` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_tcp_abort` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_tcp_close` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_tcp_flush` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_tcp_open` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_tcp_rcv` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_tcp_send` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_tcp_state` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_udp_close` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_udp_open` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_udp_rcv` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_udp_send` | MSXgl | network/unapi_tcp.h | pendente |  |
| `tcpip_udp_state` | MSXgl | network/unapi_tcp.h | pendente |  |


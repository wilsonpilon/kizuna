package kaji80

// =============================================================================
// KIZUNA KAJI80 - Rótulos pré-definidos (BIOS, BDOS, BIOSVARS)
//
// Três tabelas de símbolos sempre disponíveis, sem precisar de EXTERN nem
// INCLUDE -- resolvidas como ÚLTIMO recurso na busca de símbolo (depois de
// rótulo local, rótulo definido no arquivo, EQU e EXTERN). Código do
// usuário pode dar um nome igual a um destes e isso NÃO é erro --
// sombreamento silencioso é aceitável aqui, já que são só endereços/
// códigos numéricos conhecidos, não uma dependência viva de linkagem.
//
// biosLabels: endereços de chamada da BIOS principal (Main-ROM), incluindo
// as rotinas dos padrões MSX/MSX2/MSX2+/Turbo-R. Lista do asMSX (seção
// 2.2.1 da documentação) -- cruzada com lib/src/bios.asm, já testado em
// hardware: CHGMOD=005Fh, CALSLT=001Ch, EXPTBL=0FCC1h (esta última é uma
// BIOSVARS, não BIOS, ver abaixo) conferem exatamente.
//
// biosVars: variáveis de sistema da BIOS. Lista do asMSX (seção 2.5 da
// documentação).
//
// bdosFuncs: códigos de função do MSX-DOS/MSX-DOS2 pra CALL 0005h com o
// código em C -- o asMSX NÃO tem uma tabela pronta pra isso (só suporta
// `.CALLDOS X` com X numérico), então esta tabela é nova, compilada a
// partir da documentação padrão de MSX-DOS. Deliberadamente CONSERVADORA:
// só inclui os códigos que também estão verificados e testados em
// hardware em lib/src/bdos.asm (F_TERM0/F_CONIN/F_CONOUT/F_STROUT das
// rotinas BDOS_Exit/ReadChar/PrintChar/PrintString, e os 6 de I/O de
// arquivo do MSX-DOS 2 de BDOS_FileOpen/Create/Close/Read/Write/Seek) mais
// um pequeno núcleo de códigos CP/M-clássicos extremamente bem
// documentados e idênticos em toda referência de MSX-DOS consultada
// (F_CONST, F_CPMVER) -- não é uma lista exaustiva de todas as funções de
// MSX-DOS 1/2 (muitas são obscuras ou variam entre versões), só um núcleo
// confiável. Pode crescer numa leva futura se fizer falta.
// =============================================================================

var biosLabels = map[string]uint16{
	"CHKRAM": 0x0000, "SETT32": 0x007b, "GTPDL": 0x00de, "SNSMAT": 0x0141,
	"SYNCHR": 0x0008, "SETGRP": 0x007e, "TAPION": 0x00e1, "PHYDIO": 0x0144,
	"RDSLT": 0x000c, "SETMLT": 0x0081, "TAPIN": 0x00e4, "FORMAT": 0x0147,
	"CHRGTR": 0x0010, "CALPAT": 0x0084, "TAPIOF": 0x00e7, "ISFLIO": 0x014a,
	"WRSLT": 0x0014, "CALATR": 0x0087, "TAPOON": 0x00ea, "OUTDLP": 0x014d,
	"OUTDO": 0x0018, "GSPSIZ": 0x008a, "TAPOUT": 0x00ed, "GETVCP": 0x0150,
	"CALSLT": 0x001c, "GRPPRT": 0x008d, "TAPOOF": 0x00f0, "GETVC2": 0x0153,
	"DCOMPR": 0x0020, "GICINI": 0x0090, "STMOTR": 0x00f3, "KILBUF": 0x0156,
	"ENASLT": 0x0024, "WRTPSG": 0x0093, "LFTQ": 0x00f6, "CALBAS": 0x0159,
	"GETYPR": 0x0028, "RDPSG": 0x0096, "PUTQ": 0x00f9, "SUBROM": 0x015c,
	"CALLF": 0x0030, "STRTMS": 0x0099, "RIGHTC": 0x00fc, "EXTROM": 0x015f,
	"KEYINT": 0x0038, "CHSNS": 0x009c, "LEFTC": 0x00ff, "CHKSLZ": 0x0162,
	"INITIO": 0x003b, "CHGET": 0x009f, "UPC": 0x0102, "CHKNEW": 0x0165,
	"INIFNK": 0x003e, "CHPUT": 0x00a2, "TUPC": 0x0105, "EOL": 0x0168,
	"DISSCR": 0x0041, "LPTOUT": 0x00a5, "DOWNC": 0x0108, "BIGFIL": 0x016b,
	"ENASCR": 0x0044, "LPTSTT": 0x00a8, "TDOWNC": 0x010b, "NSETRD": 0x016e,
	"WRTVDP": 0x0047, "CNVCHR": 0x00ab, "SCALXY": 0x010e, "NSTWRT": 0x0171,
	"RDVRM": 0x004a, "PINLIN": 0x00ae, "MAPXYC": 0x0111, "NRDVRM": 0x0174,
	"WRTVRM": 0x004d, "INLIN": 0x00b1, "FETCHC": 0x0114, "NWRVRM": 0x0177,
	"SETRD": 0x0050, "QINLIN": 0x00b4, "STOREC": 0x0117, "RDBTST": 0x017a,
	"SETWRT": 0x0053, "BREAKX": 0x00b7, "SETATR": 0x011a, "WRBTST": 0x017d,
	"FILVRM": 0x0056, "ISCNTC": 0x00ba, "READC": 0x011d, "CHGCPU": 0x0180,
	"LDIRMV": 0x0059, "CKCNTC": 0x00bd, "SETC": 0x0120, "GETCPU": 0x0183,
	"LDIRVM": 0x005c, "BEEP": 0x00c0, "NSETCX": 0x0123, "PCMPLY": 0x0186,
	"CHGMOD": 0x005f, "CLS": 0x00c3, "GTASPC": 0x0126, "PCMREC": 0x0189,
	"CHGCLR": 0x0062, "POSIT": 0x00c6, "PNTINI": 0x0129,
	"NMI": 0x0066, "FNKSB": 0x00c9, "SCANR": 0x012c,
	"CLRSPR": 0x0069, "ERAFNK": 0x00cc, "SCANL": 0x012f,
	"INITXT": 0x006c, "DSPFNK": 0x00cf, "CHGCAP": 0x0132,
	"INIT32": 0x006f, "TOTEXT": 0x00d2, "CHGSND": 0x0135,
	"INIGRP": 0x0072, "GTSTCK": 0x00d5, "RSLREG": 0x0138,
	"INIMLT": 0x0075, "GTTRIG": 0x00d8, "WSLREG": 0x013b,
	"SETTXT": 0x0078, "GTPAD": 0x00db, "RDVDP": 0x013e,
}

var biosVars = map[string]uint16{
	"CGTABL": 0x0004, "VDP_DR": 0x0006, "VDP_DW": 0x0007,
	"MSXID1": 0x002b, "MSXID2": 0x002c, "MSXID3": 0x002d,
	"RDPRIM": 0xf380, "WRPRIM": 0xf385, "CLPRIM": 0xf38c,
	"LINL40": 0xf3ae, "LINL32": 0xf3af, "LINLEN": 0xf3b0, "CRTCNT": 0xf3b1,
	"CLMLST": 0xf3b2, "TXTNAM": 0xf3b3, "TXTCOL": 0xf3b5, "TXTCGP": 0xf3b7,
	"TXTATR": 0xf3b9, "TXTPAT": 0xf3bb, "T32NAM": 0xf3bd, "T32COL": 0xf3bf,
	"T32CGP": 0xf3c1, "T32ATR": 0xf3c3, "T32PAT": 0xf3c5, "GRPNAM": 0xf3c7,
	"GRPCOL": 0xf3c9, "GRPCGP": 0xf3cb, "GRPATR": 0xf3cd, "GRPPAT": 0xf3cf,
	"MLTNAM": 0xf3d1, "MLTCOL": 0xf3d3, "MLTCGP": 0xf3d5, "MLTATR": 0xf3d7,
	"MLTPAT": 0xf3d9, "CLIKSW": 0xf3db, "CSRY": 0xf3dc, "CSRX": 0xf3dd,
	"CNSDFG": 0xf3de, "RG0SAV": 0xf3df, "RG1SAV": 0xf3e0, "RG2SAV": 0xf3e1,
	"RG3SAV": 0xf3e2, "RG4SAV": 0xf3e3, "RG5SAV": 0xf3e4, "RG6SAV": 0xf3e5,
	"RG7SAV": 0xf3e6, "STATFL": 0xf3e7, "TRGFLG": 0xf3e8, "FORCLR": 0xf3e9,
	"BAKCLR": 0xf3ea, "BDRCLR": 0xf3eb, "MAXUPD": 0xf3ec, "MINUPD": 0xf3ef,
	"ATRBYT": 0xf3f2, "QUEUES": 0xf3f3, "FRCNEW": 0xf3f5, "SCNCNT": 0xf3f6,
	"REPCNT": 0xf3f7, "PUTPNT": 0xf3f8, "GETPNT": 0xf3fa, "CS120": 0xf3fc,
	"CS240": 0xf401, "LOW": 0xf406, "HIGH": 0xf408, "HEADER": 0xf40a,
	"ASPCT1": 0xf40b, "ASPCT2": 0xf40d, "ENDPRG": 0xf40f, "ERRFLG": 0xf414,
	"LPTPOS": 0xf415, "PRTFLG": 0xf416, "NTMSXP": 0xf417, "RAWPRT": 0xf418,
	"VLZADR": 0xf419, "VLZDAT": 0xf41b, "CURLIN": 0xf41c, "EXBRSA": 0xfaf8,
	"PRSCNT": 0xfb35, "SAVSP": 0xfb36, "VOICEN": 0xfb38, "SAVVOL": 0xfb39,
	"MCLLEN": 0xfb3b, "MCLPTR": 0xfb3c, "QUEUEN": 0xfb3e, "MUSICF": 0xfb3f,
	"PLYCNT": 0xfb40, "VCBA": 0xfb41, "VCBB": 0xfb66, "VCBC": 0xfb8b,
	"ENSTOP": 0xfbb0, "BASROM": 0xfbb1, "LINTTB": 0xfbb2, "FSTPOS": 0xfbca,
	"CODSAV": 0xfbcc, "FNKSWI": 0xfbcd, "FNKFLG": 0xfbce, "ONGSBF": 0xfbd8,
	"CLIKFL": 0xfbd9, "OLDKEY": 0xfbda, "NEWKEY": 0xfbe5, "KEYBUF": 0xfbf0,
	"BUFEND": 0xfc18, "LINWRK": 0xfc18, "PATWRK": 0xfc40, "BOTTOM": 0xfc48,
	"HIMEM": 0xfc4a, "TRPTBL": 0xfc4c, "RTYCNT": 0xfc9a, "INTFLG": 0xfc9b,
	"PADY": 0xfc9c, "PADX": 0xfc9d, "JIFFY": 0xfc9e, "INTVAL": 0xfca0,
	"INTCNT": 0xfca2, "LOWLIM": 0xfca4, "WINWID": 0xfca5, "GRPHED": 0xfca6,
	"ESCCNT": 0xfca7, "INSFLG": 0xfca8, "CSRSW": 0xfca9, "CSTYLE": 0xfcaa,
	"CAPST": 0xfcab, "KANAST": 0xfcac, "KANAMD": 0xfcad, "FLBMEM": 0xfcae,
	"SCRMOD": 0xfcaf, "OLDSCR": 0xfcb0, "CASPRV": 0xfcb1, "BRDATR": 0xfcb2,
	"GXPOS": 0xfcb3, "GYPOS": 0xfcb5, "GRPACX": 0xfcb7, "GRPACY": 0xfcb9,
	"DRWFLG": 0xfcbb, "DRWSCL": 0xfcbc, "DRWANG": 0xfcbd, "RUNBNF": 0xfcbe,
	"SAVENT": 0xfcbf, "EXPTBL": 0xfcc1, "SLTTBL": 0xfcc5, "SLTATR": 0xfcc9,
	"SLTWRK": 0xfd09, "PROCNM": 0xfd89, "DEVICE": 0xfd99,
}

var bdosFuncs = map[string]uint16{
	// Núcleo CP/M-clássico, idêntico em MSX-DOS 1 e 2 -- já usado e
	// testado em hardware em lib/src/bdos.asm (BDOS_Exit/ReadChar/
	// PrintChar/PrintString).
	"F_TERM0":  0x00,
	"F_CONIN":  0x01,
	"F_CONOUT": 0x02,
	"F_STROUT": 0x09,
	// Resto do núcleo CP/M clássico, bem documentado, não usado ainda em
	// nenhuma rotina da MSXLIB mas de alta confiança (idêntico em toda
	// referência de MSX-DOS consultada).
	"F_CONST":  0x0b,
	"F_CPMVER": 0x0c,
	// I/O de arquivo baseado em handle, MSX-DOS 2 -- já usado e testado
	// em hardware (BDOS_FileOpen/Create/Close/Read/Write/Seek).
	"F_OPEN":   0x43,
	"F_CREATE": 0x44,
	"F_CLOSE":  0x45,
	"F_READ":   0x48,
	"F_WRITE":  0x49,
	"F_SEEK":   0x4a,
}

// lookupPredefined procura "name" nas três tabelas pré-definidas, nesta
// ordem -- BIOS e BDOS raramente colidem em nome (convenções de nomeação
// diferentes), mas se colidirem, BIOS vence (é a tabela mais usada/
// consultada na prática).
func lookupPredefined(name string) (uint16, bool) {
	if v, ok := biosLabels[name]; ok {
		return v, true
	}
	if v, ok := biosVars[name]; ok {
		return v, true
	}
	if v, ok := bdosFuncs[name]; ok {
		return v, true
	}
	return 0, false
}

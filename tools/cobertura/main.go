// Comando cobertura gera docs/cobertura.md: a lista de TUDO o que o MSXgl e o
// Fusion-C oferecem (só nomes de função, agrupados por área da MSXLIB) com o
// estado de cada item no KIZUNA. É o checklist do plano de expansão da MSXLIB
// (docs/plano-expansao-msxlib.md): os dois projetos servem apenas de guia do
// que cobrir; nenhum código deles é lido para dentro da biblioteca.
//
// Estado de cada item:
//
//	feito          mapeado (docs/cobertura.map) para uma rotina que EXISTE em lib/
//	planejado      mapeado para uma rotina que ainda não existe
//	não faremos    mapeado com "-- motivo"
//	pendente       ainda sem decisão
//
// "Existe" = aparece num PUBLIC de lib/src/**/*.asm ou numa declaração de
// lib/api/*.api.
//
// Uso (na raiz do repositório):
//
//	go run ./tools/cobertura            # escreve docs/cobertura.md
//	go run ./tools/cobertura -check     # só valida o mapa (nomes do mapa que não existem nos guias)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	msxglDir    = "resource/msxgl/engine/src"
	fusionDir   = "resource/MSXFusionC/Working Folder/fusion-c/header"
	mapFile     = "docs/cobertura.map"
	outFile     = "docs/cobertura.md"
	areaOutside = "Fora da lista (por enquanto)"
)

// Ordem e nomes das áreas (a mesma do plano).
var areaOrder = []string{
	"Portas e constantes", "Memória", "Matemática", "Strings e texto", "VDP", "Draw", "Tile", "Scroll",
	"Teclado", "Joystick e mouse", "PSG", "Play (players)", "MSX-Music", "MSX-Audio", "SCC",
	"BIOS", "DOS", "System", "Clock", "V9990", areaOutside,
}

type item struct {
	name   string
	origin string // "MSXgl" ou "Fusion-C"
	file   string // cabeçalho onde foi declarada
	area   string
}

var declRe = regexp.MustCompile(`^\s*(?:extern\s+)?(?:static\s+)?(?:inline\s+)?(?:const\s+)?(?:unsigned\s+|signed\s+)?(?:void|char|int|short|long|float|double|bool|Bool|boolean|byte|uint|u8|u16|u32|i8|i16|i32|f32|FCB|TIME|DATE|Palette|MOUSE_DATA|[A-Z][A-Za-z0-9_]*)\s*\*?\s*\**\s*([A-Za-z_][A-Za-z0-9_]*)\s*\([^;{]*\)\s*(?:__[A-Z]+(?:\([^)]*\))?\s*)*(?:;|\{)`)

func main() {
	check := flag.Bool("check", false, "só valida o mapa")
	flag.Parse()

	items := collect()
	mapping, err := loadMap(mapFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	known := implemented()

	if *check {
		names := map[string]bool{}
		for _, it := range items {
			names[it.name] = true
			names["fusion:"+it.name] = true
		}
		bad := 0
		for k := range mapping {
			if !names[k] {
				fmt.Printf("mapa: %q não existe nos cabeçalhos guia\n", k)
				bad++
			}
		}
		if bad > 0 {
			os.Exit(1)
		}
		fmt.Printf("mapa ok: %d entradas\n", len(mapping))
		return
	}

	// agrupa
	byArea := map[string][]item{}
	for _, it := range items {
		byArea[it.area] = append(byArea[it.area], it)
	}

	var sb strings.Builder
	sb.WriteString("# Cobertura: MSXgl e Fusion-C → MSXLIB\n\n")
	sb.WriteString("> Arquivo **gerado** por `go run ./tools/cobertura` a partir dos cabeçalhos em `resource/` e do mapa\n")
	sb.WriteString("> `docs/cobertura.map`. Não edite à mão: edite o mapa e gere de novo.\n")
	sb.WriteString(">\n")
	sb.WriteString("> O MSXgl e o Fusion-C são apenas **guias do que cobrir** (ver `docs/plano-expansao-msxlib.md`):\n")
	sb.WriteString("> aqui só constam os nomes das funções deles; nenhuma implementação é copiada para a MSXLIB.\n\n")

	sb.WriteString("## Resumo\n\n| Área | Itens | Feito | Planejado | Não faremos | Pendente |\n| --- | ---: | ---: | ---: | ---: | ---: |\n")
	type counts struct{ total, done, plan, no, pend int }
	totals := counts{}
	stateOf := func(it item) (string, string) {
		key := it.name
		tgt, ok := mapping[key]
		if it.origin == "Fusion-C" {
			if t, ok2 := mapping["fusion:"+it.name]; ok2 {
				tgt, ok = t, true
			}
		}
		switch {
		case !ok:
			return "pendente", ""
		case strings.HasPrefix(tgt, "--"):
			return "não faremos", strings.TrimSpace(strings.TrimPrefix(tgt, "--"))
		case known[tgt]:
			return "feito", tgt
		default:
			return "planejado", tgt
		}
	}
	for _, area := range areaOrder {
		var c counts
		for _, it := range byArea[area] {
			st, _ := stateOf(it)
			c.total++
			switch st {
			case "feito":
				c.done++
			case "planejado":
				c.plan++
			case "não faremos":
				c.no++
			default:
				c.pend++
			}
		}
		if c.total == 0 {
			continue
		}
		fmt.Fprintf(&sb, "| %s | %d | %d | %d | %d | %d |\n", area, c.total, c.done, c.plan, c.no, c.pend)
		totals.total += c.total
		totals.done += c.done
		totals.plan += c.plan
		totals.no += c.no
		totals.pend += c.pend
	}
	fmt.Fprintf(&sb, "| **Total** | **%d** | **%d** | **%d** | **%d** | **%d** |\n\n", totals.total, totals.done, totals.plan, totals.no, totals.pend)

	for _, area := range areaOrder {
		list := byArea[area]
		if len(list) == 0 {
			continue
		}
		sort.Slice(list, func(i, j int) bool {
			if list[i].origin != list[j].origin {
				return list[i].origin < list[j].origin
			}
			return list[i].name < list[j].name
		})
		fmt.Fprintf(&sb, "## %s\n\n| Função (guia) | Origem | Cabeçalho | Estado | Nossa rotina / motivo |\n| --- | --- | --- | --- | --- |\n", area)
		for _, it := range list {
			st, note := stateOf(it)
			fmt.Fprintf(&sb, "| `%s` | %s | %s | %s | %s |\n", it.name, it.origin, it.file, st, note)
		}
		sb.WriteString("\n")
	}

	if err := os.WriteFile(outFile, []byte(sb.String()), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %d itens (%d feitos, %d planejados, %d não faremos, %d pendentes)\n",
		outFile, totals.total, totals.done, totals.plan, totals.no, totals.pend)
}

// ---- coleta ----------------------------------------------------------------

func collect() []item {
	var items []item
	seen := map[string]bool{}
	add := func(name, origin, file, area string) {
		key := origin + "|" + name
		if seen[key] {
			return
		}
		seen[key] = true
		items = append(items, item{name, origin, file, area})
	}

	// MSXgl: por arquivo/diretório
	filepath.WalkDir(msxglDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".h") {
			return nil
		}
		rel, _ := filepath.Rel(msxglDir, path)
		rel = filepath.ToSlash(rel)
		area := msxglArea(rel)
		if area == "" {
			return nil
		}
		for _, n := range declNames(path) {
			add(n, "MSXgl", rel, area)
		}
		return nil
	})

	// Fusion-C
	filepath.WalkDir(fusionDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".h") {
			return nil
		}
		rel, _ := filepath.Rel(fusionDir, path)
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "font8x8/") || rel == "newTypes.h" || strings.HasPrefix(rel, "vars_") {
			return nil
		}
		for _, n := range declNames(path) {
			add(n, "Fusion-C", rel, fusionArea(rel, n))
		}
		return nil
	})
	return items
}

func declNames(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "//") || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if m := declRe.FindStringSubmatch(line); m != nil {
			n := m[1]
			if len(n) < 2 || n == "if" || n == "while" || n == "for" || n == "return" || n == "sizeof" || n == "switch" {
				continue
			}
			out = append(out, n)
		}
	}
	return out
}

func msxglArea(rel string) string {
	base := strings.TrimSuffix(rel, ".h")
	switch {
	case base == "memory" || base == "mutex":
		return "Memória"
	case base == "math" || base == "fixed_point":
		return "Matemática"
	case base == "string" || base == "print":
		return "Strings e texto"
	case base == "vdp":
		return "VDP"
	case base == "draw":
		return "Draw"
	case base == "tile":
		return "Tile"
	case base == "scroll":
		return "Scroll"
	case base == "keyboard" || base == "input" || base == "input_manager":
		return "Teclado"
	case base == "joystick" || base == "mouse":
		return "Joystick e mouse"
	case base == "psg":
		return "PSG"
	case base == "msx-music":
		return "MSX-Music"
	case base == "msx-audio":
		return "MSX-Audio"
	case base == "scc":
		return "SCC"
	case base == "bios" || base == "bios_hook":
		return "BIOS"
	case base == "dos" || base == "dos_mapper":
		return "DOS"
	case base == "system" || base == "basic_usr":
		return "System"
	case base == "clock":
		return "Clock"
	case base == "v9990":
		return "V9990"
	case strings.HasPrefix(rel, "pcm/") || strings.HasPrefix(rel, "ndp/") || strings.HasPrefix(rel, "mglv/") ||
		strings.HasPrefix(rel, "standalone/") || strings.HasPrefix(rel, "arkos/") || strings.HasPrefix(rel, "ayfx/") ||
		strings.HasPrefix(rel, "pt3/") || strings.HasPrefix(rel, "trilo/") || strings.HasPrefix(rel, "vgm/") ||
		strings.HasPrefix(rel, "wyz/"):
		return "Play (players)"
	case base == "sprite_fx" || base == "fsm" || base == "game_pawn" || base == "crypt" || base == "compress" ||
		base == "localize" || strings.HasPrefix(rel, "device/") || strings.HasPrefix(rel, "game/") ||
		strings.HasPrefix(rel, "network/") || strings.HasPrefix(rel, "compress/") || strings.HasPrefix(rel, "msxi/"):
		return areaOutside
	}
	return "" // cabeçalhos sem funções (constantes, config, crt0...) não entram
}

func fusionArea(rel, name string) string {
	l := strings.ToLower(name)
	has := func(prefixes ...string) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(l, p) {
				return true
			}
		}
		return false
	}
	switch {
	case rel == "g9klib.h":
		return "V9990"
	case rel == "gr8net-tcpip.h":
		return areaOutside
	case rel == "rammapper.h":
		return "DOS"
	case rel == "pt3replayer.h" || rel == "ayfx_player.h":
		return "Play (players)"
	case rel == "psg.h" || has("psg", "initpsg", "sound", "soundfx", "setchannel", "settoneperiod", "setnoiseperiod",
		"setenvelopeperiod", "setvolume", "playenvelope", "silencepsg", "getsound"):
		return "PSG"
	case has("covoxplay", "pcmplay"):
		return "Play (players)"
	case rel == "vdp_graph1.h" || rel == "vdp_graph2.h" || rel == "vdp_circle.h":
		if has("hmm", "lmm", "hmcm", "ymmm", "flmmm", "setsc5", "restoresc5") {
			return "VDP"
		}
		return "Draw"
	case rel == "vdp_sprites.h":
		return "VDP"
	case has("vdp", "vpoke", "vpeek", "fillvram", "copyramtovram", "copyvramtoram", "getvramsize", "screen", "setcolors",
		"setbordercolor", "setpalette", "setdisplaypage", "setactivepage", "hidedisplay", "showdisplay", "width", "putsprite"):
		return "VDP"
	case has("setscroll"):
		return "Scroll"
	case has("joystick", "trigger", "mouse"):
		return "Joystick e mouse"
	case has("keyboard", "inkey", "waitkey", "killkeybuffer", "getkeymatrix", "checkbreak", "keysound", "functionkeys", "changecap"):
		return "Teclado"
	case has("mem", "mmalloc"):
		return "Memória"
	case has("getdate", "setdate", "gettime", "settime", "realtimer", "setrealtimer"):
		return "Clock"
	case has("fcb_", "exit", "intdos", "intbios", "doscls"):
		return "DOS"
	case has("readmsxtype", "getcpu", "changecpu", "readtpa", "readsp", "suspend", "initinterrupthandler",
		"endinterrupthandler", "setinterrupthandler", "inport", "outport", "outports"):
		return "System"
	case has("rlewb"):
		return areaOutside
	case has("cls", "beep", "bchput", "locate", "print", "inputchar", "inputstring", "getche", "num2dec", "putchar",
		"chartolower", "chartoupper", "str", "nstr", "is", "itoa", "inttofloat", "intswap", "ispositive"):
		return "Strings e texto"
	}
	return "System"
}

// ---- mapa e rotinas existentes --------------------------------------------

// loadMap lê "Guia = Nossa" ou "Guia = -- motivo"; '#' comenta.
func loadMap(path string) (map[string]string, error) {
	m := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("%s:%d: esperado \"guia = nossa\"", path, n)
		}
		m[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return m, sc.Err()
}

var publicRe = regexp.MustCompile(`(?i)^\s*PUBLIC\s+(.*)`)
var apiRe = regexp.MustCompile(`(?i)^\s*(?:proc|func)\s+([A-Za-z_][A-Za-z0-9_]*)`)

// implemented reúne o que existe em lib/: PUBLIC dos .asm e rotinas dos .api.
func implemented() map[string]bool {
	known := map[string]bool{}
	filepath.WalkDir("lib/src", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".asm") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			if i := strings.Index(line, ";"); i >= 0 {
				line = line[:i]
			}
			if m := publicRe.FindStringSubmatch(line); m != nil {
				for _, n := range strings.Split(m[1], ",") {
					known[strings.TrimSpace(n)] = true
				}
			}
		}
		return nil
	})
	filepath.WalkDir("lib/api", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".api") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if m := apiRe.FindStringSubmatch(sc.Text()); m != nil {
				known[m[1]] = true
			}
		}
		return nil
	})
	return known
}

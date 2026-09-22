package obi

import (
	"fmt"
	"strconv"
	"strings"
)

// rawLine é uma linha do Obifile já sem comentários/linhas em branco, com sua
// indentação (contagem de espaços à esquerda) e número original para mensagens
// de erro.
type rawLine struct {
	indent int
	text   string
	num    int
}

// ParseObifile interpreta o conteúdo de um Obifile no subconjunto de sintaxe
// suportado (chaves "campo: valor" no nível 0, listas via "  - campo: valor" e
// mapas aninhados de um nível via "  campo: valor" — não é um parser YAML
// genérico, só o necessário para a receita de build do KIZUNA).
func ParseObifile(content string) (*Config, error) {
	lines := tokenizeLines(content)

	cfg := &Config{}
	i := 0
	for i < len(lines) {
		ln := lines[i]
		if ln.indent != 0 {
			return nil, fmt.Errorf("linha %d: indentação inesperada fora de um bloco reconhecido", ln.num)
		}

		key, value, ok := splitKV(ln.text)
		if !ok {
			return nil, fmt.Errorf("linha %d: esperado 'campo: valor', obtido '%s'", ln.num, ln.text)
		}

		switch strings.ToLower(key) {
		case "target":
			if value == "" {
				return nil, fmt.Errorf("linha %d: 'target' não pode ser vazio", ln.num)
			}
			cfg.Target = value
			i++

		case "entry":
			cfg.Entry = value
			i++

		case "base":
			base, err := parseAddress(value)
			if err != nil {
				return nil, fmt.Errorf("linha %d: %w", ln.num, err)
			}
			cfg.Base = base
			i++

		case "modules":
			items, next, err := collectListBlock(lines, i+1)
			if err != nil {
				return nil, err
			}
			for _, fields := range items {
				mod, err := parseModuleItem(fields)
				if err != nil {
					return nil, fmt.Errorf("em 'modules' (bloco iniciado após linha %d): %w", ln.num, err)
				}
				cfg.Modules = append(cfg.Modules, mod)
			}
			i = next

		case "resources":
			items, next, err := collectListBlock(lines, i+1)
			if err != nil {
				return nil, err
			}
			for _, fields := range items {
				res, err := parseResourceItem(fields)
				if err != nil {
					return nil, fmt.Errorf("em 'resources' (bloco iniciado após linha %d): %w", ln.num, err)
				}
				cfg.Resources = append(cfg.Resources, res)
			}
			i = next

		case "link":
			fields, next, err := collectMapBlock(lines, i+1)
			if err != nil {
				return nil, err
			}
			cfg.Link.Map = fields["map"]
			i = next

		case "library":
			fields, next, err := collectMapBlock(lines, i+1)
			if err != nil {
				return nil, err
			}
			if archive := fields["archive"]; archive != "" {
				cfg.Libraries = append(cfg.Libraries, archive)
			}
			i = next

		case "libraries":
			items, next, err := collectStringListBlock(lines, i+1)
			if err != nil {
				return nil, err
			}
			cfg.Libraries = append(cfg.Libraries, items...)
			i = next

		default:
			return nil, fmt.Errorf("linha %d: chave desconhecida '%s'", ln.num, key)
		}
	}

	if cfg.Target == "" {
		return nil, fmt.Errorf("campo obrigatório 'target' ausente no Obifile")
	}

	return cfg, nil
}

func tokenizeLines(content string) []rawLine {
	var out []rawLine
	for idx, raw := range strings.Split(content, "\n") {
		line := strings.TrimRight(raw, " \t\r")
		trimmed := strings.TrimLeft(line, " ")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(trimmed)
		out = append(out, rawLine{indent: indent, text: trimmed, num: idx + 1})
	}
	return out
}

func splitKV(text string) (key, value string, ok bool) {
	rawKey, rawValue, found := strings.Cut(text, ":")
	if !found {
		return "", "", false
	}
	key = strings.TrimSpace(rawKey)
	value = strings.TrimSpace(rawValue)
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

// collectListBlock lê um bloco de itens "  - campo: valor" (com possíveis
// campos adicionais indentados mais profundamente sob o mesmo item) até
// encontrar uma linha de indentação 0 ou o fim do arquivo.
func collectListBlock(lines []rawLine, start int) ([]map[string]string, int, error) {
	var items []map[string]string
	i := start
	itemIndent := -1
	var current map[string]string

	for i < len(lines) {
		ln := lines[i]
		if ln.indent == 0 {
			break
		}

		if strings.HasPrefix(ln.text, "- ") || ln.text == "-" {
			if itemIndent == -1 {
				itemIndent = ln.indent
			} else if ln.indent != itemIndent {
				return nil, 0, fmt.Errorf("linha %d: indentação inconsistente na lista", ln.num)
			}

			current = map[string]string{}
			items = append(items, current)

			rest := strings.TrimSpace(strings.TrimPrefix(ln.text, "-"))
			if rest != "" {
				k, v, ok := splitKV(rest)
				if !ok {
					return nil, 0, fmt.Errorf("linha %d: esperado 'campo: valor' após '-'", ln.num)
				}
				current[strings.ToLower(k)] = v
			}
			i++
			continue
		}

		if current == nil {
			return nil, 0, fmt.Errorf("linha %d: campo fora de um item de lista (esperava '- campo: valor')", ln.num)
		}
		k, v, ok := splitKV(ln.text)
		if !ok {
			return nil, 0, fmt.Errorf("linha %d: esperado 'campo: valor'", ln.num)
		}
		current[strings.ToLower(k)] = v
		i++
	}

	return items, i, nil
}

// collectStringListBlock lê um bloco de itens escalares simples: "  - valor".
func collectStringListBlock(lines []rawLine, start int) ([]string, int, error) {
	var out []string
	i := start
	for i < len(lines) {
		ln := lines[i]
		if ln.indent == 0 {
			break
		}
		if !strings.HasPrefix(ln.text, "- ") && ln.text != "-" {
			return nil, 0, fmt.Errorf("linha %d: esperado item de lista '- valor'", ln.num)
		}
		v := strings.TrimSpace(strings.TrimPrefix(ln.text, "-"))
		out = append(out, v)
		i++
	}
	return out, i, nil
}

// collectMapBlock lê um bloco de campos "  campo: valor" de um único nível
// (usado por "link:" e "library:").
func collectMapBlock(lines []rawLine, start int) (map[string]string, int, error) {
	out := map[string]string{}
	i := start
	for i < len(lines) {
		ln := lines[i]
		if ln.indent == 0 {
			break
		}
		k, v, ok := splitKV(ln.text)
		if !ok {
			return nil, 0, fmt.Errorf("linha %d: esperado 'campo: valor'", ln.num)
		}
		out[strings.ToLower(k)] = v
		i++
	}
	return out, i, nil
}

func parseModuleItem(fields map[string]string) (ModuleSpec, error) {
	m := ModuleSpec{
		Name:     fields["name"],
		Source:   fields["source"],
		Compiler: strings.ToLower(fields["compiler"]),
	}
	if m.Source == "" {
		return m, fmt.Errorf("módulo sem campo obrigatório 'source'")
	}
	if bankStr, ok := fields["bank"]; ok && bankStr != "" {
		n, err := strconv.Atoi(bankStr)
		if err != nil {
			return m, fmt.Errorf("banco inválido '%s' no módulo '%s': %w", bankStr, m.Source, err)
		}
		m.Bank = n
		m.HasBank = true
	}
	return m, nil
}

func parseResourceItem(fields map[string]string) (ResourceSpec, error) {
	r := ResourceSpec{
		File:   fields["file"],
		Symbol: fields["symbol"],
	}
	if r.File == "" {
		return r, fmt.Errorf("resource sem campo obrigatório 'file'")
	}
	if bankStr, ok := fields["bank"]; ok && bankStr != "" {
		n, err := strconv.Atoi(bankStr)
		if err != nil {
			return r, fmt.Errorf("banco inválido '%s' no resource '%s': %w", bankStr, r.File, err)
		}
		r.Bank = n
	}
	if sizeStr, ok := fields["size"]; ok && sizeStr != "" {
		n, err := parseSize(sizeStr)
		if err != nil {
			return r, fmt.Errorf("tamanho inválido '%s' no resource '%s': %w", sizeStr, r.File, err)
		}
		r.Size = n
		r.HasSize = true
	}
	return r, nil
}

// parseSize aceita um inteiro decimal simples ou um inteiro seguido de 'K'/'k'
// (KiB), no estilo já usado no Obifile ilustrativo original (ex: "12K").
func parseSize(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("tamanho vazio")
	}
	mult := 1
	if up := strings.ToUpper(s); strings.HasSuffix(up, "K") {
		mult = 1024
		s = s[:len(s)-1]
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("tamanho inválido: %w", err)
	}
	return n * mult, nil
}

// parseAddress aceita "0x0100" (hex com prefixo), "0100h" (hex com sufixo) ou
// um decimal simples, no mesmo espírito dos formatos numéricos do KAJI80.
func parseAddress(s string) (uint16, error) {
	s = strings.TrimSpace(s)
	low := strings.ToLower(s)
	switch {
	case strings.HasPrefix(low, "0x"):
		v, err := strconv.ParseUint(s[2:], 16, 16)
		if err != nil {
			return 0, fmt.Errorf("endereço hexadecimal inválido '%s'", s)
		}
		return uint16(v), nil
	case strings.HasSuffix(low, "h"):
		v, err := strconv.ParseUint(s[:len(s)-1], 16, 16)
		if err != nil {
			return 0, fmt.Errorf("endereço hexadecimal inválido '%s'", s)
		}
		return uint16(v), nil
	default:
		v, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return 0, fmt.Errorf("endereço inválido '%s'", s)
		}
		return uint16(v), nil
	}
}

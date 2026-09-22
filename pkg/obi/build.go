package obi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/kizuna/pkg/dignac"
	"github.com/wilsonpilon/kizuna/pkg/hako"
	"github.com/wilsonpilon/kizuna/pkg/kaji80"
	"github.com/wilsonpilon/kizuna/pkg/mob"
	"github.com/wilsonpilon/kizuna/pkg/musubi"
	"github.com/wilsonpilon/kizuna/pkg/wirth80"
)

// BuildOptions controla o comportamento do orquestrador.
type BuildOptions struct {
	Verbose bool
}

// ModuleResult descreve o resultado da compilação de um módulo ou resource.
type ModuleResult struct {
	Kind     string // "module" ou "resource"
	Name     string
	Compiler string
	Source   string
	MobPath  string
	Bank     uint8
	Size     int
}

// BuildResult resume o que o orquestrador fez.
type BuildResult struct {
	Target      string
	IsLibrary   bool
	Modules     []ModuleResult
	Libraries   []string
	LinkResult  *musubi.LinkResult // nil quando Target é .hlib
	PackedFiles []string           // arquivos .mob empacotados quando Target é .hlib
}

// Build executa a receita descrita em cfg. baseDir é o diretório usado para
// resolver todos os caminhos relativos do Obifile (fontes, resources,
// bibliotecas e o próprio target).
func Build(cfg *Config, baseDir string, opts BuildOptions) (*BuildResult, error) {
	if len(cfg.Modules) == 0 && len(cfg.Resources) == 0 {
		return nil, fmt.Errorf("Obifile não declara nenhum módulo nem resource")
	}

	targetPath := resolvePath(baseDir, cfg.Target)
	res := &BuildResult{Target: targetPath}

	var linkInputs []string

	for _, m := range cfg.Modules {
		mr, err := buildModule(baseDir, m)
		if err != nil {
			return nil, fmt.Errorf("módulo '%s': %w", displayName(m.Name, m.Source), err)
		}
		res.Modules = append(res.Modules, *mr)
		linkInputs = append(linkInputs, mr.MobPath)
	}

	for _, r := range cfg.Resources {
		mr, err := buildResource(baseDir, r)
		if err != nil {
			return nil, fmt.Errorf("resource '%s': %w", r.File, err)
		}
		res.Modules = append(res.Modules, *mr)
		linkInputs = append(linkInputs, mr.MobPath)
	}

	for _, lib := range cfg.Libraries {
		libPath := resolvePath(baseDir, lib)
		if _, err := os.Stat(libPath); err != nil {
			return nil, fmt.Errorf("biblioteca '%s' não encontrada: %w", lib, err)
		}
		res.Libraries = append(res.Libraries, libPath)
	}

	switch strings.ToLower(filepath.Ext(targetPath)) {
	case ".hlib":
		res.IsLibrary = true
		if len(res.Libraries) > 0 {
			return nil, fmt.Errorf("target '%s' é uma biblioteca (.hlib): 'library'/'libraries' só se aplicam a target .com", cfg.Target)
		}
		if err := hako.Pack(targetPath, linkInputs...); err != nil {
			return nil, fmt.Errorf("hako: %w", err)
		}
		res.PackedFiles = linkInputs

	case ".com":
		lcfg := musubi.DefaultConfig()
		if cfg.Entry != "" {
			lcfg.EntryPoint = cfg.Entry
		}
		if cfg.Base != 0 {
			lcfg.BaseAddress = cfg.Base
		}
		if cfg.Link.Map != "" {
			lcfg.MapFile = resolvePath(baseDir, cfg.Link.Map)
		}
		lcfg.Verbose = opts.Verbose

		allInputs := append(append([]string{}, linkInputs...), res.Libraries...)
		linkRes, err := musubi.LinkToFile(targetPath, lcfg, allInputs...)
		if err != nil {
			return nil, fmt.Errorf("musubi: %w", err)
		}
		res.LinkResult = linkRes

	default:
		return nil, fmt.Errorf("target '%s': extensão não suportada (use .com ou .hlib)", cfg.Target)
	}

	return res, nil
}

func buildModule(baseDir string, m ModuleSpec) (*ModuleResult, error) {
	srcPath := resolvePath(baseDir, m.Source)
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler '%s': %w", m.Source, err)
	}

	compiler := m.Compiler
	if compiler == "" {
		compiler, err = inferCompiler(m.Source)
		if err != nil {
			return nil, err
		}
	}

	obj, err := compileModule(compiler, string(content))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", compiler, err)
	}

	if m.HasBank {
		if err := validateModuleBank(obj, m.Bank, m.Source); err != nil {
			return nil, err
		}
	}

	mobPath := strings.TrimSuffix(srcPath, filepath.Ext(srcPath)) + ".mob"
	if err := mob.SaveToFile(mobPath, obj); err != nil {
		return nil, fmt.Errorf("falha ao salvar '%s': %w", mobPath, err)
	}

	name := m.Name
	if name == "" {
		name = filepath.Base(m.Source)
	}

	size := 0
	for _, seg := range obj.Segments {
		size += int(seg.Size)
	}

	return &ModuleResult{
		Kind:     "module",
		Name:     name,
		Compiler: compiler,
		Source:   m.Source,
		MobPath:  mobPath,
		Bank:     firstBank(obj),
		Size:     size,
	}, nil
}

// compileModule despacha para o frontend certo e devolve o objeto compilado
// em memória, reusando exatamente as mesmas APIs que cmd/kaji80, cmd/wirth80
// e cmd/dignac já usam — nenhum subprocesso é lançado.
func compileModule(compiler, source string) (*mob.ObjectFile, error) {
	switch compiler {
	case "kaji80":
		return kaji80.NewAssembler().Assemble(source)

	case "wirth80":
		parser, err := wirth80.NewParser(wirth80.NewLexer(source))
		if err != nil {
			return nil, err
		}
		prog, err := parser.ParseProgram()
		if err != nil {
			return nil, err
		}
		obj, _, err := wirth80.NewCodeGenerator(prog).Compile()
		return obj, err

	case "dignac":
		parser, err := dignac.NewParser(dignac.NewLexer(source))
		if err != nil {
			return nil, err
		}
		modAst, err := parser.ParseModule()
		if err != nil {
			return nil, err
		}
		obj, _, err := dignac.NewCodeGenerator(modAst).Compile()
		return obj, err

	default:
		return nil, fmt.Errorf("compilador desconhecido '%s' (use kaji80, wirth80 ou dignac)", compiler)
	}
}

func buildResource(baseDir string, r ResourceSpec) (*ModuleResult, error) {
	filePath := resolvePath(baseDir, r.File)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler '%s': %w", r.File, err)
	}
	if r.HasSize && len(data) > r.Size {
		return nil, fmt.Errorf("arquivo '%s' (%d bytes) excede o tamanho declarado (%d bytes)", r.File, len(data), r.Size)
	}

	symbol := r.Symbol
	if symbol == "" {
		symbol = deriveSymbol(r.File)
	}

	obj := mob.NewObjectFile()
	segIdx := obj.AddSegment(mob.SegmentData, uint8(r.Bank), data, 0)
	obj.AddSymbol(symbol, mob.SymbolPublic, mob.SymbolData, segIdx, 0)

	mobPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".mob"
	if err := mob.SaveToFile(mobPath, obj); err != nil {
		return nil, fmt.Errorf("falha ao salvar '%s': %w", mobPath, err)
	}

	return &ModuleResult{
		Kind:     "resource",
		Name:     symbol,
		Compiler: "resource",
		Source:   r.File,
		MobPath:  mobPath,
		Bank:     uint8(r.Bank),
		Size:     len(data),
	}, nil
}

func validateModuleBank(obj *mob.ObjectFile, declared int, source string) error {
	for _, seg := range obj.Segments {
		if int(seg.Bank) != declared {
			return fmt.Errorf("banco declarado no Obifile (%d) não bate com o BANK do fonte '%s' (%d)", declared, source, seg.Bank)
		}
	}
	return nil
}

func firstBank(obj *mob.ObjectFile) uint8 {
	if len(obj.Segments) == 0 {
		return 0
	}
	return obj.Segments[0].Bank
}

func inferCompiler(source string) (string, error) {
	switch strings.ToLower(filepath.Ext(source)) {
	case ".asm":
		return "kaji80", nil
	case ".pas":
		return "wirth80", nil
	case ".bas":
		return "dignac", nil
	default:
		return "", fmt.Errorf("não foi possível inferir o compilador de '%s' (declare 'compiler:' explicitamente)", source)
	}
}

// deriveSymbol gera o nome do símbolo PUBLIC de um resource sem 'symbol:'
// explícito no Obifile, a partir do nome base do arquivo.
func deriveSymbol(file string) string {
	base := filepath.Base(file)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	var sb strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			sb.WriteRune(r)
		default:
			sb.WriteRune('_')
		}
	}
	name := sb.String()
	if name == "" {
		name = "Resource"
	}
	return "Res_" + name
}

func resolvePath(baseDir, p string) string {
	if p == "" {
		return p
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(baseDir, p)
}

func displayName(name, source string) string {
	if name != "" {
		return name
	}
	return source
}

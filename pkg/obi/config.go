package obi

// Config é a representação em memória de um Obifile já interpretado.
type Config struct {
	Target string // nome do arquivo de saída: <nome>.com (executável) ou <nome>.hlib (biblioteca)
	Entry  string // símbolo de entrada (só usado quando Target termina em .com); default "Start"
	Base   uint16 // endereço base de carregamento (só usado quando Target termina em .com); default 0x0100

	Resources []ResourceSpec
	Modules   []ModuleSpec

	Link      LinkSpec
	Libraries []string // caminhos .hlib já empacotados, incluídos na linkagem final
}

// ModuleSpec descreve um módulo-fonte a ser compilado por um dos frontends.
type ModuleSpec struct {
	Name     string // opcional, só para relatório/log
	Source   string // caminho do arquivo-fonte (.asm, .pas ou .bas)
	Compiler string // "kaji80" | "wirth80" | "dignac"; inferido da extensão se vazio
	Bank     int    // opcional, só informativo/validação; o BANK real vem do próprio fonte
	HasBank  bool   // true se "bank:" foi explicitado no Obifile
}

// ResourceSpec descreve um arquivo binário bruto a ser embutido como um módulo
// sintético de dados, exportando um único símbolo PUBLIC de dados.
type ResourceSpec struct {
	File   string // caminho do arquivo bruto
	Bank   int    // banco de destino (0 = área comum)
	Symbol string // nome do símbolo PUBLIC exportado; derivado do basename se vazio
	Size   int    // opcional, tamanho máximo declarado em bytes (checagem, sem padding)
	HasSize bool
}

// LinkSpec contém opções repassadas ao MUSUBI quando Target termina em .com.
type LinkSpec struct {
	Map string // caminho opcional do relatório de mapa de memória (.map)
}

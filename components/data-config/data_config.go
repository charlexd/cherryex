package cherryDataConfig

var (
	parserMap     = make(map[string]IDataParser) //文件格式解析器
	dataSourceMap = make(map[string]IDataSource) //数据配置数据源
)

func init() {
	RegisterParser(new(ParserJson))
	RegisterSource(new(SourceFile))
	RegisterSource(new(SourceRedis))
}

func GetParser(name string) IDataParser {
	return parserMap[name]
}

func RegisterParser(parser IDataParser) {
	parserMap[parser.TypeName()] = parser
}

func GetDataSource(name string) IDataSource {
	return dataSourceMap[name]
}

func RegisterSource(dataSource IDataSource) {
	dataSourceMap[dataSource.Name()] = dataSource
}

// BaseConfig 基础配置
type BaseConfig struct {
}

func (p *BaseConfig) Name() string {
	return "base_config"
}

func (p *BaseConfig) Init() {

}

func (p *BaseConfig) OnLoad(maps interface{}, reload bool) (int, error) {
	return 0, nil
}

func (p *BaseConfig) OnAfterLoad(reload bool) {

}

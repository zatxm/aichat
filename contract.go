package aichat

type AiCom interface {
	Proxy(string)
	// 通用chat请求(统一请求格式)
	Chat(*RequestChat) (*Stream, error)
	// 原始api json数据请求(不一定是json数据，以下同)
	ChatApi(string) (*Stream, error)
	// 原始json数据字符串chat请求
	ChatSource(string) (*Stream, error)
	// 通用chat请求(统一请求格式)转Api返回格式
	ChatToApi(*RequestChat) (*Stream, error)
	// 原始json数据字符串chat请求转Api返回格式
	ChatToApiSource(string) (*Stream, error)
	// api原始json数据字符串转chat请求再转Api返回格式
	ApiCrossChatToApi(string) (*Stream, error)
	// 通用chat请求(统一请求格式)转成openai api返回格式
	ChatToOpenai(*RequestChat) (*Stream, error)
	// 原始json数据字符串chat请求转成openai api返回格式
	ChatToOpenaiSource(string) (*Stream, error)
	// api原始json数据字符串请求转Api返回格式
	ApiToOpenai(string) (*Stream, error)
	// api原始json数据字符串转chat请求再转openai api返回格式
	ApiCrossChatToOpenai(string) (*Stream, error)
}

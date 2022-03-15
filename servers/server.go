package servers

const (
	APIVersion = "v1"
)

/*
Server 是一个服务器结构接口
抽象出一个接口是因为后续会提供TCP服务
*/
type Server interface {
	// Run 在address上启动服务器，并返回错误信息
	Run(address string) error
}

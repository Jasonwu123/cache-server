package caches

// Options 是一些选项的结构体
type Options struct {
	MaxEntrySize int64  // 单位GB,写满保护阈值，当缓存中的键值对占用空间达到该值就触发写满保护
	MaxGcCount   int    // 自动淘汰机制阈值，当清理的数据达到该值后就会停止清理
	GcDuration   int64  // 自动淘汰机制时间间隔，每隔固定的GcDuraton时间就会进行一次自动淘汰
	DumpFile     string // 持久化文件的路径
	DumpDuration int64  // 持久化时间间隔，单位是分钟
}

// DefaultOptions 返回一个默认的选项设置对象
func DefaultOptions() Options {
	return Options{
		MaxEntrySize: int64(4),            // 默认是4GB
		MaxGcCount:   1000,                // 默认是1000ge
		GcDuration:   60,                  // 默认是1小时
		DumpFile:     "cache-server.dump", // 默认的持久化文件在缓存启动位置，名字为cache-server.dump
		DumpDuration: 30,                  // 默认的持久化时间间隔设置为30分钟
	}
}

package caches

/*
Status 是一个代表缓存信息的结构体
因为需要被序列化为JSON字符串并通过网络传输，所以这里使用JSON标签
*/
type Status struct {
	Count     int   `json:"count"`     // 记录缓存中的数据项个数
	KeySize   int64 `json:"keySize"`   // 记录key占用的空间大小
	ValueSize int64 `json:"valueSize"` // 记录value占用的空间大小
}

// newStatus 返回一个缓存信息对象指针
func newStatus() *Status {
	return &Status{
		Count:     0,
		KeySize:   0,
		ValueSize: 0,
	}
}

// addEntry 可以将key和value的信息记录起来
func (s *Status) addEntry(key string, value []byte) {
	/*
		每添加一个键值对，Count就加1，key占用的空间就是string的长度
		同理，value占用的空间就是切片的长度
	*/
	s.Count++
	s.KeySize += int64(len(key))
	s.ValueSize += int64(len(value))
}

// subEntry 可以将key和value的信息从Status中减去
func (s *Status) subEntry(key string, value []byte) {
	// 每减少一个键值对，Count就减1，key和value占用的空间也需要减去相应的大小
	s.Count--
	s.KeySize -= int64(len(key))
	s.ValueSize -= int64(len(value))
}

// entrySize 返回键值对占用的总大小
func (s *Status) entrySize() int64 {
	return s.KeySize + s.ValueSize
}

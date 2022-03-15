package helpers

// Copy 复制src到新的[]byte中并返回
func Copy(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

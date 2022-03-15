package caches

import (
	"cacheServer/cache-server/helpers"
	"sync/atomic"
	"time"
)

const (
	NeverDie = 0 // 是一个常量，永不过期
)

// value是一个包装了数据的结构体
type value struct {
	data  []byte // 存储真正的数据
	ttl   int64  // 数据寿命，单位是秒
	ctime int64  //  数据创建时间
}

// newValue返回一个包装后的数据
func newValue(data []byte, ttl int64) *value {
	return &value{
		data:  helpers.Copy(data),
		ttl:   ttl,
		ctime: time.Now().Unix(),
	}
}

// alive返回数据是否存活
func (v *value) alive() bool {
	// 首先判断是否有过期时间，然后判断当前时间是否超过了这个数据的死期
	return v.ttl == NeverDie || time.Now().Unix()-v.ctime < v.ttl
}

// visit返回数据的实际存储数据
func (v *value) visit() []byte {
	// 这一步是为了实现 LRU 过期机制而加的
	// 在访问数据的时候，将创建时间更新为访问时间，这样就相当于最近访问的数据过期时间会延长
	// 因为获取数据一般都在读取操作中进行，读取操作使用的是读锁，尽可能保证并发的性能
	// 使用读锁就意味着没有保证写的并发安全，所以我们需要自己去处理并发安全的问题
	// 一般会使用锁来处理，但是这里牵扯到读取性能，如果使用锁，就会显得非常臃肿和拉跨
	// 于是就使用了 atomic 包轻量化地去处理，这里直接使用了交换更新数据，而没有使用 CAS 的方式
	// 后交换成功的会把先交换成功的时间改掉，所以这里不保证交换的时间一定是更加新的时间
	atomic.SwapInt64(&v.ctime, time.Now().Unix())
	return v.data
}

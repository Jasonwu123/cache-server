package caches

import (
	"errors"
	"sync"
	"time"
)

// Cache 结构体：封装缓存底层结构
type Cache struct {
	data    map[string]*value //存储着实际的键值对数据
	options Options           // 选项设置，值传递，只读不修改
	status  *Status           // 缓存的状态信息
	lock    *sync.RWMutex     // 读写锁，保证并发安全
}

// NewCache 返回Cache对象
func NewCache() *Cache {
	return NewCacheWith(DefaultOptions())
}

// NewCacheWith 返回一个指定配置的缓存对象
func NewCacheWith(options Options) *Cache {
	// 先从持久化文件进行恢复，如果恢复不成功，则返回一个空的缓存
	if cache, ok := recoverFromDumpFile(options.DumpFile); ok {
		return cache
	}

	return &Cache{
		data:    make(map[string]*value, 256),
		options: options,
		status:  newStatus(),
		lock:    &sync.RWMutex{},
	}
}

// recoverFromDumpFile 从dumpFile中恢复缓存，如果恢复不成功则返回nil和false
func recoverFromDumpFile(dumpFile string) (*Cache, bool) {
	cache, err := newEmptyDump().from(dumpFile)
	if err != nil {
		return nil, false
	}
	return cache, true
}

// Get 返回指定key的value，找不到则返回false
func (c *Cache) Get(key string) ([]byte, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value, ok := c.data[key]
	if !ok {
		return nil, false
	}

	/*
		如果数据不是存活的，将数据删除掉，返回找不到数据
		注意对锁的操作，由于一开始加的是读锁，无法保证写的并发安全，而删除需要加写锁，读锁和写锁又是互斥的
		所以先将读锁释放，再上写锁，删除操作里面会加写锁，删除完后写锁释放，再上读锁
	*/

	if !value.alive() {
		c.lock.RUnlock()
		c.Delete(key)
		c.lock.RLock()
		return nil, false
	}

	// visit方法会使用Swap的形式更新数据的创建时间，用于实现LRU(least recently used)
	return value.visit(), true
}

// Set 添加一个键值对到缓存中，不设定ttl，也就意味着数据不会过期
// 返回error是nil说明添加成功，否则就是添加失败，可能是触发了写满保护机制，拒绝写入数据
func (c *Cache) Set(key string, value []byte) error {
	return c.SetWithTTL(key, value, NeverDie)
}

// SetWithTTL 添加一个键值对到缓存中，使用给定的ttl去设定过期时间
// 返回error是nil说明添加成功，否则就是添加失败，可能是触发了写满保护机制，拒绝写入数据
func (c Cache) SetWithTTL(key string, value []byte, ttl int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if oldValue, ok := c.data[key]; ok {
		// 如果是已经存在的key，就不属于新增，为方便管理，把原来的键值对信息去除
		c.status.subEntry(key, oldValue.Data)
	}

	// 判断缓存容量是否足够，如果不够，则返回写满保护的错误信息
	if !c.checkEntrySize(key, value) {
		// 注意刚刚把旧的键值对信息去除了，现在要加回去，因为并没有添加新的键值对
		if oldValue, ok := c.data[key]; ok {
			c.status.addEntry(key, oldValue.Data)
		}
		return errors.New("the entry size will exceed if you set this entry.")
	}

	// 添加新的键值对，需要先更新缓存信息，然后保存数据
	c.status.addEntry(key, value)
	c.data[key] = newValue(value, ttl)

	return nil
}

// Delete 删除指定key的键值对数据
func (c *Cache) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if oldValue, ok := c.data[key]; ok {
		// 如果存在key才会删除，并且需要先把缓存信息更新掉
		c.status.subEntry(key, oldValue.Data)
		delete(c.data, key)
	}
}

// Status 返回缓存信息
func (c *Cache) Status() Status {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return *c.status
}

// checkEntrySize 检查要添加的键值对是否满足当前的缓存要求
func (c Cache) checkEntrySize(newKey string, newValue []byte) bool {
	// 将当前的键值对占用空间加上要被添加的键值对占用空间，然后和配置中的的最大键值对占用空间进行比较
	return c.status.entrySize()+int64(len(newKey))+int64(len(newValue)) <= c.options.MaxEntrySize*1024*1024
}

// gc会触发数据清理任务，主要是清理过期的数据
func (c *Cache) gc() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 使用count记录当前的清理个数
	count := 0
	for key, value := range c.data {
		if !value.alive() {
			c.status.subEntry(key, value.Data)
			delete(c.data, key)
			count++
			if count >= c.options.MaxGcCount {
				break
			}
		}
	}
}

// AutoGc 会开启一个定时GC异步任务
func (c *Cache) AutoGc() {
	go func() {
		// 根据配置中的GcDuration来设置定时的间隔
		ticker := time.NewTicker(time.Duration(c.options.GcDuration) * time.Minute)
		for {
			select {
			case <-ticker.C:
				c.gc()
			}
		}
	}()
}

// dump 持久化缓存方法
func (c *Cache) dump() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 创建出dump对象并持久化到文件
	return newDump(c).to(c.options.DumpFile)
}

// AutoDump 开启定时任务去持久化缓存
func (c *Cache) AutoDump() {
	go func() {
		ticker := time.NewTicker(time.Duration(c.options.DumpDuration) * time.Minute)
		for {
			select {
			case <-ticker.C:
				c.dump()
			}
		}
	}()
}
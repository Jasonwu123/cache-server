package caches

import (
	"encoding/gob"
	"os"
	"sync"
	"time"
)

// dump 持久化
type dump struct {
	Data    map[string]*value
	Options Options
	Status  *Status
}

// newEmptyDump 创建一个空的dump结构对象并返回
func newEmptyDump() *dump {
	return &dump{}
}

// newDump 创建一个dump对象并使用指定的Cache对象初始化
func newDump(c *Cache) *dump {
	return &dump{
		Data:    c.data,
		Options: c.options,
		Status:  c.status,
	}
}

// nowSuffix 返回一个类似于20060102150405的文件后缀名
func nowSuffix() string {
	return "." + time.Now().Format("20060102150405")
}

// to 会将dump持久化到dumpfile中
func (d *dump) to(dumpFile string) error {
	newDumpFile := dumpFile + nowSuffix()
	file, err := os.OpenFile(newDumpFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	err = gob.NewEncoder(file).Encode(d)
	if err != nil {
		// 注意这里先把文件关闭，不然os.Remove是没有权限删除文件
		file.Close()
		os.Remove(newDumpFile)
		return err
	}

	// 将旧的持久化文件删除
	os.Remove(dumpFile)

	// 将新的持久化文件改名为旧的持久化名字，相当于替换，这样就可以保证持久化文件的名字不变
	// 注意先关闭文件，不然无权限操作
	file.Close()
	return os.Rename(newDumpFile, dumpFile)
}

// from 会从dumpfile中恢复数据到一个Cache结构对象并返回
func (d *dump) from(dumpFile string) (*Cache, error) {
	// 读取dumpFile文件并使用反序列化器进行反序列化
	file, err := os.Open(dumpFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err = gob.NewDecoder(file).Decode(d); err != nil {
		return nil, err
	}

	// 初始化一个缓存对象并返回
	return &Cache{
		data:    d.Data,
		options: d.Options,
		status:  d.Status,
		lock:    &sync.RWMutex{},
	}, nil
}
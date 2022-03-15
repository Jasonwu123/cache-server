package servers

import (
	"cacheServer/cache-server/caches"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
)

// HTTPServer 是 HTTP 服务器结构
type HTTPServer struct {
	cache *caches.Cache
}

// NewHTTPServer 返回一个关于cache的新HTTP服务器
func NewHTTPServer(cache *caches.Cache) *HTTPServer {
	return &HTTPServer{cache: cache}
}

// Run 启动服务器
func (hs *HTTPServer) Run(address string) error {
	return http.ListenAndServe(address, hs.routerHandler())
}

// wrapUriWithVersion 会用API版本去包装uri，例如"v1"版本的API包装"/cache"就会变成"/v1/cache"
func wrapUriWithVersion(uri string) string {
	return path.Join("/", APIVersion, uri)
}

// routerHandler 返回路由处理器给http包中注册用
func (hs *HTTPServer) routerHandler() http.Handler {
	router := httprouter.New()
	router.GET("/cache/:key", hs.getHandler)
	router.PUT("/cache/:key", hs.setHandler)
	router.DELETE("/cache/:key", hs.deleteHandler)
	router.GET("/status", hs.statusHandler)
	return router
}

// getHandler 获取缓存数据项
func (hs *HTTPServer) getHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	value, ok := hs.cache.Get(key)
	if !ok {
		// 缓存中找不到数据，返回404状态码
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write(value)
}

// setHandler 保存缓存数据
func (hs *HTTPServer) setHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	// value从请求体中获取，整个请求体都被当做value
	value, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// 如果读取请求体失败，则返回500状态码
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 从请求中获取ttl
	ttl, err := ttlOf(r)
	if err != nil {
		// 返回500错误码
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 添加数据，并设置为指定的ttl
	err = hs.cache.SetWithTTL(key, value, ttl)
	if err != nil {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		w.Write([]byte("Error: " + err.Error()))
		return
	}

	// 成功添加就返回201状态码
	w.WriteHeader(http.StatusCreated)
}

// ttlOf 从请求中解析ttl并返回，如果error不为空，说明ttl解析出错
func ttlOf(r *http.Request) (int64, error) {
	// 从请求头中获取ttl头部，如果没有设置或者ttl为空均按照不设置ttl处理，也就是不过期
	ttls, ok := r.Header["Ttl"]
	if !ok || len(ttls) < 1 {
		return caches.NeverDie, nil
	}

	return strconv.ParseInt(ttls[0], 10, 64)
}

// deleteHandler 删除缓存数据
func (hs *HTTPServer) deleteHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	hs.cache.Delete(key)
}

// statusHandler 获取键值对个数
func (hs *HTTPServer) statusHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 将个数编码成JSON字符串
	status, err := json.Marshal(hs.cache.Status())
	if err != nil {
		// 编码失败，返回500状态码
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(status)
}

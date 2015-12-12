
package pools

import(
    "../memcache"
    "time"
)

type MemcachePool struct{
    *RoundRobin
    Host string
}

const connIdleTimeout = "5m" //5 minutes 

func NewMemcachePool(host string, capacity int) *MemcachePool {
    
    idleTimeout, err := time.ParseDuration(connIdleTimeout)
    if err!=nil {
        panic(err)
    }
    p := &MemcachePool{NewRoundRobin(capacity, idleTimeout), host}
    p.open(cacheCreator(host))
    return p
}

func (this *MemcachePool) open(connFactory CreateCacheFunc){
	if connFactory == nil {
		return
	}
	f := func() (Resource, error) {
		c, err := connFactory()
		if err != nil {
			return nil, err
		}
		return &Cache{c, this}, nil
	}
	this.RoundRobin.Open(f)
}

// You must call Recycle on the *Cache once done.
func (this *MemcachePool)Get() *Cache{
	r, err := this.RoundRobin.Get()
	if err != nil {
		panic(err)
	}
	return r.(*Cache)
}


type CreateCacheFunc func() (*memcache.Connection, error)

func cacheCreator(host string) CreateCacheFunc {
	return func() (*memcache.Connection, error) {
		return memcache.Connect(host)
	}
}

// Cache re-exposes memcache.Connection
type Cache struct {
	conn *memcache.Connection
	pool *MemcachePool
}


func (this *Cache) Recycle() {
	this.pool.Put(this)
}

func (this *Cache) Close() {
	this.conn.Close()
}    

func (this *Cache) IsClosed() bool {
	return this.conn.IsClosed()
}

func (this *Cache) Get(key string) []byte {
	value, _, err := this.conn.Get(key)
	if err != nil {
	    this.conn.Close()
	    return nil
	}
	return value
}

func (this *Cache) Set(key string, value []byte, timeout uint64) bool {
	stored, err := this.conn.Set(key, 0, timeout, value)
    if err != nil {
        this.conn.Close()
        return false
    }	
	return stored
}

func (this *Cache) Delete(key string) bool {
    deleted, err := this.conn.Delete(key)
    if err != nil {
        this.conn.Close()
        return false
    }
    return deleted
}


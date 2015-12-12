
package hashmemcache

import(
    "../pools"    
    "hash/crc32"
    "../relog"
)

const(
    maxConcurrentConn = 50    
)

type HashPools struct {
    connPools [] *pools.MemcachePool
    servers []string
}

func New(servers []string, maxConn int) *HashPools {
    if len(servers) < 1 {
        relog.Fatal("empty server list")
    }
    
    if maxConn < 1 {
        maxConn = maxConcurrentConn
    }
    
    var p [] *pools.MemcachePool = nil
    for _, val := range servers {
        p = append(p, pools.NewMemcachePool(val, maxConn))
    }
    return &HashPools{
        connPools : p,
        servers : servers,  
    }
}

func (this *HashPools) hashIndex(key string) int {
    return int(crc32.ChecksumIEEE([]byte(key))) % len(this.connPools)
}

func (this *HashPools) ConnGet(key string) *pools.Cache{
    pool := this.connPools[this.hashIndex(key)]
    return pool.Get()
}

func (this *HashPools) ConnRecycle(c *pools.Cache){
    c.Recycle()
}

func (this *HashPools) Get(key string, conn *pools.Cache) []byte {
    if conn == nil {
        conn = this.ConnGet(key)
        defer this.ConnRecycle(conn)
    }
     
    return conn.Get(key)
}

func (this *HashPools) Set(key string, value []byte, timeout uint64) bool {
    conn := this.ConnGet(key)
    defer this.ConnRecycle(conn)
    
    return conn.Set(key, value, timeout)
}

func (this *HashPools) Delete(key string) bool {
    conn := this.ConnGet(key)
    defer this.ConnRecycle(conn)
    
    return conn.Delete(key)
}

func (this *HashPools) Servers() []string {
    return this.servers
}

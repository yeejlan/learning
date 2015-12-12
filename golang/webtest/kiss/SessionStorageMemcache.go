
package kiss

import(
    "encoding/json"    
    "fmt"
    "../lib/hashmemcache"
    "../lib/relog"
    "errors"
)

var hashPools *hashmemcache.HashPools = nil  

type MemcacheSessionStorage struct{
    session SessionStorage
}

func setupMemcacheSessionStorage(servers []string, maxConn int){
    if len(servers) == 0 {
        servers = []string{"127.0.0.1:11211"}
    }
    hashPools = hashmemcache.New(servers, maxConn)
}

func NewMemcacheSessionStorage(ctx WebContext, expire int64) *MemcacheSessionStorage {
    if hashPools == nil {
        relog.Fatal("please call setupMemcacheSessionStorage() first")
    }
    
    s:= &MemcacheSessionStorage{
        session : NewSessionStorage(ctx, expire),
    }
    
    ctx["kiss.session"] = s
    s.Open()
    s.Read()
    return s    
}

func (this *MemcacheSessionStorage) Open() error{
    return nil
}

func (this *MemcacheSessionStorage) Close() error{
    if !this.session.closed {
        //make sure session is written to storage before close it
        this.Write()
        this.session.closed = true
    }
    return nil
}

func (this *MemcacheSessionStorage) Read() error {
    this.session.saved = true
    
    b := hashPools.Get(this.SessionId(), nil)

    var m interface{}
    err := json.Unmarshal(b, &m)
    if err != nil {
        return err
    }
    
    for k, v:= range m.(map[string]interface{}){
        this.Set(k, v.(string))
    }

    return nil
}

func (this *MemcacheSessionStorage) Write() error {
    if !this.session.saved {
        //write to storage
        bs, err := json.Marshal(this.session.content)
        if err != nil {
            return err
        }
        
        isOK := hashPools.Set(this.SessionId(), bs, uint64(this.session.expire))
        if !isOK {
            return errors.New("Write: store session failed")
        }
        
        this.session.saved = true       
    }
    return nil
}

func (this *MemcacheSessionStorage) Gc() error {
    return nil
}

func (this *MemcacheSessionStorage) Get(key string) string {
    return this.session.Get(key)
}

func (this *MemcacheSessionStorage) Set(key, value string) {
    this.session.Set(key, value)
}

func (this *MemcacheSessionStorage) Delete(key string) {
    this.session.Delete(key)
}

func (this *MemcacheSessionStorage) Purge(){
    this.session.Purge()
}

func (this *MemcacheSessionStorage) SessionId() string {
    return this.session.id
}

func (this *MemcacheSessionStorage) String() string {
    servers := fmt.Sprintf("\"Servers\": %v, ", hashPools.Servers())
    return servers + this.session.String()
}




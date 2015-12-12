
package kiss

import(
    "fmt"
    "sync"
    "os"
    "path/filepath"
    "../lib/relog"
)


const defaultSessionExpire = 25*3600

var sessionStatus = sessionStatusStruct{
    gcCounter : 0,
    gcRunning : false,
}

var sessionName = "SESSIONID"
var sessionDir = ""

type ISessionStorage interface{
    Open() error
    Close() error
    Read() error
    Write() error
    Gc() error
    
    Get(key string) string
    Set(key, value string)
    Delete(key string)
    Purge()
    
    SessionId() string
    String() string
}

type sessionStatusStruct struct{
    mu sync.Mutex
    counter int64
    gcCounter int64
    gcRunning bool
}

type SessionStorage struct{
    id string
    expire int64
    content map[string]string
    saved bool //session saved or not
    closed bool //session closed or not
}

func NewSessionStorage(ctx WebContext, expire int64) SessionStorage {
    if expire < 1 {
        expire = defaultSessionExpire
    }
    
    sessionId := ctx.GetCookie(GetSessionName())
    if sessionId == "" {
       sessionId = GetUniqueID()
       ctx.SetCookie(GetSessionName(), sessionId, -1)       
    }
    
    return SessionStorage{
            id : sessionId,
            expire : expire,
            content : make(map[string]string),
            saved : false,
            closed : false,
           }
}

func SetSessionName(name string){
    sessionName = name
}

func GetSessionName() string {
    return sessionName
}

func SetSessionDir(sessDir string){
    if sessDir == "" {
        dir := os.TempDir()
        sessDir = filepath.Join(dir, "gosession-" + GetSessionName())
    }
    err := os.Mkdir(sessDir, 0755)
    if err!=nil {
        if os.IsExist(err) {
            //pass
        }else{
            relog.Fatal("Mkdir(%s) error %s", sessDir, err)
        }
    }
    sessionDir = sessDir
}

func GetSessionDir() string {
    return sessionDir
}

///*-----------------------functions for SessionStorage----------*/
func (this *SessionStorage) Get(key string) string {
    val, exist := this.content[key]
    if !exist {
        return ""
    }
    return val
}

func (this *SessionStorage) Set(key, value string) {
    this.saved = false
    this.content[key] = value
}

func (this *SessionStorage) Delete(key string) {
    this.saved = false
    delete(this.content, key)
}

func (this *SessionStorage) Purge(){
    this.saved = false
    this.content = make(map[string]string)
}

func (this *SessionStorage) SessionId() string{
    return this.id
}

func (this *SessionStorage) String() string {
    return fmt.Sprintf("\"Id\": %v , \"Expire\": %v, \"Content\": %v", 
        this.id, this.expire, this.content)
}

///*-----------------------functions for sessionStatusStruct----------*/
func (this *sessionStatusStruct) GcCounter() int64{
    this.mu.Lock()
    defer this.mu.Unlock()
    
    return this.gcCounter
}

func (this *sessionStatusStruct) GcCounterIncr(){
    this.mu.Lock()
    defer this.mu.Unlock()
    
    this.gcCounter++
}

func (this *sessionStatusStruct) GcCounterReset(){
    this.mu.Lock()
    defer this.mu.Unlock()
    
    this.gcCounter = 0
}

func (this *sessionStatusStruct) GcRunning() bool{
    this.mu.Lock()
    defer this.mu.Unlock()
    
    return this.gcRunning
}

func (this *sessionStatusStruct) SetGcRunning(gcRunning bool){
    this.mu.Lock()
    defer this.mu.Unlock()
    
    this.gcRunning = gcRunning
}




package kiss

import(
    "encoding/json"
    "path/filepath"
    "io/ioutil"
    "time"
    "fmt"
    "strconv"
    "os"
)

const fileSessionStorageGcThreshold = 1000

type FileSessionStorage struct{
    session SessionStorage
}

func NewFileSessionStorage(ctx WebContext, expire int64) *FileSessionStorage {
    s:= &FileSessionStorage{
        session : NewSessionStorage(ctx, expire),
    }
    
    ctx["kiss.session"] = s
    s.Open()
    s.Read()
    return s
}

func (this *FileSessionStorage) Open() error{
    sessionStatus.GcCounterIncr()
    if sessionStatus.GcCounter() > fileSessionStorageGcThreshold {
        sessionStatus.GcCounterReset()
        go func(){
            this.Gc()
        }()
    }
    return nil
}

func (this *FileSessionStorage) Close() error{
    if !this.session.closed {
        //make sure session is written to storage before close it
        this.Write()
        this.session.closed = true
    }
    return nil
}

func (this *FileSessionStorage) Read() error {
    
    this.session.saved = true
    sessFile := this.getSessionFilePath()
    bytes, err := ioutil.ReadFile(sessFile)
    if err != nil{
        return err
    }
    
    var m interface{}
    err = json.Unmarshal(bytes, &m)
    if err != nil {
        return err
    }
    //check session expire
    sessionOk := false
    expire, exist := m.(map[string]interface{})["$expire"]
    if exist {
        expireTime, err := strconv.Atoi(expire.(string))
        if err!=nil{
            expireTime = 0
        }
        if time.Now().Unix() < int64(expireTime) {
            sessionOk = true
        }
    }
    
    if sessionOk {
        for k, v:= range m.(map[string]interface{}){
            this.Set(k, v.(string))
        }
    }
    
    return nil
}

func (this *FileSessionStorage) Write() error {
    if !this.session.saved {
        //write session expire info
        expire := time.Now().Unix() + this.session.expire
        this.Set("$expire", fmt.Sprintf("%d", expire))
        
        //write to storage
        bytes, err := json.Marshal(this.session.content)
        if err != nil {
            return err
        }
        
        sessFile := this.getSessionFilePath()
        err = ioutil.WriteFile(sessFile, bytes, 0644)
        if err != nil {
            return err
        }
        
        this.session.saved = true       
    }
    return nil
}

func (this *FileSessionStorage) Gc() error{
    filepath.Walk(GetSessionDir(),sessionFileWalker)
    return nil
}

func sessionFileWalker (path string, info os.FileInfo, err error) error{
    
    if !info.IsDir(){
        fname := path
        bytes, err := ioutil.ReadFile(fname)
        if err != nil{
            os.Remove(fname)
            return nil
        }
        
        var m interface{}
        err = json.Unmarshal(bytes, &m)
        if err != nil {
            return err
        }
        //check session expire
        expire, exist := m.(map[string]interface{})["$expire"]
        if !exist{
            os.Remove(fname)
            return nil        
        }

        expireTime, err := strconv.Atoi(expire.(string))
        if err!=nil{
            expireTime = 0
        }
        if time.Now().Unix() >= int64(expireTime) {
            os.Remove(fname)
        }
    }
    return nil
}

func (this *FileSessionStorage) Get(key string) string {
    return this.session.Get(key)
}

func (this *FileSessionStorage) Set(key, value string) {
    this.session.Set(key, value)
}

func (this *FileSessionStorage) Delete(key string) {
    this.session.Delete(key)
}

func (this *FileSessionStorage) Purge(){
    this.session.Purge()
}

func (this *FileSessionStorage) SessionId() string {
    return this.session.id
}

func (this *FileSessionStorage) String() string {
    return this.session.String()
}

///*----------------private functions-------------/
func (this *FileSessionStorage) getSessionFilePath() string {
    sessionFile := filepath.Join(GetSessionDir(), this.SessionId() + ".txt")
    return sessionFile
}


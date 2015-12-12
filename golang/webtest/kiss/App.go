
package kiss

import(
	"path"
	"../lib/relog"
	"strings"
)

type AppStruct struct{
    config ConfigMap
    env string
    domain string
    basepath string
}

var App = AppStruct{}

func init(){
    App.Bootstrap()
}

func (this *AppStruct) Bootstrap() {
    this.env = GetApplicationEnv()
    this.basepath = GetBasePath()
    
    defaultConfigFile := path.Join(App.BasePath(), "config/" + App.Env() + ".json")
    this.LoadConfig(defaultConfigFile)
    
    //set domain
    this.domain = this.Config().GetString("app", "domain")
    
    //set session
    sessionName := this.Config().GetString("session", "name")
    if sessionName != "" {
        SetSessionName(sessionName)
    }
    sessionStorage := this.Config().GetString("session", "storage")
    if sessionStorage=="" || sessionStorage == "file" {
        sessionDir := this.Config().GetString("session", "dir")
        SetSessionDir(sessionDir)
    }else if sessionStorage == "memcache" {
        serverStr := this.Config().GetString("session", "servers")
        servers := strings.Split(serverStr, ",")
        maxConn := this.Config().GetInt("session", "maxconn")
        setupMemcacheSessionStorage(servers, maxConn)
    }
}

func (this *AppStruct) LoadConfig(configFile string){
	
	config, err := ParseConfig(configFile)
	if err != nil{
		relog.Fatal("Loading config: " + configFile + " FAILED. Error: " + err.Error())
		config = NewConfigMap()
	}
	
	if this.Config() == nil{
		this.config = config
	}else{
		this.config = MergeConfig(this.Config(), config)
	}
}

func (this *AppStruct) Config() ConfigMap {
    return this.config
}

func (this *AppStruct) Env() string {
    return this.env
}

func (this *AppStruct) Domain() string {
    return this.domain
}

func (this *AppStruct) BasePath() string {
    return this.basepath
}

func (this *AppStruct) Shutdown(){
	
	LoggerCloseAll()
}


package kiss

import(
    "text/template"   
    "path/filepath"
    "io"
    "io/ioutil"
    "bytes"
    "../lib/relog"
)

var TplCache = map[string]*template.Template{}
var TplHelper = map[string]interface{}{}

func init(){
    AddHelper("include", includeHelper)
}

//render a template
func Render(w io.Writer, tplFile string, data interface{}){
    var tpl *template.Template
    var exist bool
    
    //get from cache
    tpl, exist = TplCache[tplFile]
    if !exist {
        tplFilePath := filepath.Join(App.BasePath(), "template", tplFile)
        tplContent, err := ioutil.ReadFile(tplFilePath)
        if err!= nil {
            relog.Warning("ReadFile error %s", err)    
        } 
        tpl = template.Must(template.New(tplFile).Funcs(TplHelper).Parse(string(tplContent)))
        TplCache[tplFile] = tpl //set cache
    }
    err := tpl.Execute(w, data)
    if err!= nil {
        relog.Warning("Execute %s error: %s", tplFile, err)     
    }
}

func AddHelper(name string, function interface{}){
    TplHelper[name] = function
}

func includeHelper(tplFile string, data interface{}) string {
    var b bytes.Buffer
    Render(&b, tplFile, data)
    return b.String()
}


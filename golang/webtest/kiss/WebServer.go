
package kiss

import(
    "net"
    "net/http"
    "net/http/fcgi"
    "net/url"
    "fmt"
    "os"
    "path"
    "regexp"
    "runtime"
    "strconv"
    "strings"
    "../lib/relog"
)

var webServer = &WebServer{}

type WebServer struct{
    listener net.Listener
}

func StopWebServer(){
    if webServer.listener != nil {
        webServer.listener.Close()
    }    
}

func StartWebServer(addr, serverType string){
    mux := createServerMuxAndListen(addr)
        
    relog.Info("Serving as " + serverType + " at " + addr + ", env: " + App.Env())

    var err error
    if serverType == "http" {
        err = http.Serve(webServer.listener, mux)
    }else if serverType == "fastcgi" || serverType == "fcgi" {
        err = fcgi.Serve(webServer.listener, mux)
    }
    if err != nil {
        relog.Info("StartWebServer: " + err.Error())
    }
    
    StopWebServer()
}

func createServerMuxAndListen(addr string) *http.ServeMux{
    var err error
    mux := createServerMux()
        
    webServer.listener, err = net.Listen("tcp", addr)
    if err != nil {
        relog.Fatal("createServerMuxAndListen: " + err.Error())
    }
    
    return mux
}

func createServerMux() *http.ServeMux {
    mux := http.NewServeMux()
    mux.Handle("/", http.HandlerFunc(defaultDispatcher))
    return mux
}

func defaultDispatcher(resp http.ResponseWriter, req *http.Request){
    defer func(){
        if e := recover(); e != nil {
            relog.Warning("Dispatch error: %v", e)
        }
    }()
    
    requestPath := req.URL.Path
    
    //serve static file
    staticFile := path.Join(App.BasePath(), "public", requestPath)
    if fileExists(staticFile) && (req.Method == "GET" || req.Method == "HEAD") {
        http.ServeFile(resp, req, staticFile)
        return
    }
    
    ctx := NewWebContext(req, resp)
    
    //parse request params
    if(req.Method == "POST"){
        req.ParseMultipartForm(5*1024*1024)
        
        //parse upload file field
        ctx["files"] = url.Values{}
        if req.MultipartForm != nil{
            files := ctx["files"].(url.Values)
            for key, fhArr := range req.MultipartForm.File{
                for _, fh := range fhArr{
                    fname := (*fh).Filename
                    files.Add(key, fname)
                }
            }
        }
    }else{
        req.ParseForm()
    }
    ctx["kiss.params"] = req.Form
    
    //create session storage
    sessionStorage := App.Config().GetString("session", "storage")
    sessionExpire := int64(App.Config().GetInt("session", "expire"))
    if sessionStorage=="" || sessionStorage == "file" {
        NewFileSessionStorage(ctx, sessionExpire)    
    }else{
        NewMemcacheSessionStorage(ctx, sessionExpire)
    }
    
    //set default content type
    ctx.SetHeader("Content-Type", "text/html; charset=utf-8")
    
    var pageHandler *func(WebContext)
    
    //handle regex router
    for i := 0; i < len(routes); i++ {
        route := routes[i]
        cr := route.cr

        if !cr.MatchString(requestPath) {
            continue
        }
        match := cr.FindStringSubmatch(requestPath)

        if len(match[0]) != len(requestPath) {
            continue
        }

        //adding regex match to $1, $2 etc
        for mIdx, mVal := range match[1:] {
            mIdxStr := "$" + strconv.Itoa(mIdx+1)
            ctx.Params().Set(mIdxStr, mVal)
        }
        
        ctx.Params().Set("$r", route.r)
        
        pageHandler = &route.handler
    }
    
    //handle actions mapping router
    if pageHandler == nil {
        var curSection, curAction string
        
        if requestPath == "/" { //home page
            curSection = "home"
            curAction = "index"
        }else{
            pathStr := requestPath[1:]
            if strings.HasSuffix(requestPath, "/"){
                pathStr = requestPath[1:len(requestPath)-1]
            }
            pathArr := strings.SplitN(pathStr, "/", 2)
            curSection = pathArr[0]
            if len(pathArr) == 1 { //only have section
                curAction = "index"
            }else { //have section and action
                curAction = pathArr[1]      
            }
        }
        
        ctx.Params().Set("$controller", curSection)
        ctx.Params().Set("$action", curAction)
        pageHandler = getHandlerThouActionsMap(curSection, curAction)
    }
    
    
    if pageHandler != nil { //handler found
        err := safelyCall(*pageHandler, ctx)
        if err != nil {
            //there was an error or panic while calling the handler
            ctx.Abort(500, "Internal Server Error")
        }        
    }else{ //no handler found
        handlePageNotFound(ctx)
    }
}

func safelyCall(handler func(WebContext), ctx WebContext) (e interface{}) {
    defer func() {
        if err := recover(); err != nil {
            e = err
            
            errStr := fmt.Sprintf("------------\r\nHandler crashed with error: %+v\r\n", err)
                errStr += fmt.Sprintf("Params: %+v", ctx.Params())
            for i := 1; ; i += 1 {
                pc, file, line, ok := runtime.Caller(i)
                if !ok {
                    break
                }
                funcName := runtime.FuncForPC(pc).Name()
                errStr += fmt.Sprintf("\r\nfunc: %s, file: %s, line: %d", funcName, file, line)
            }
            relog.Warning(errStr)
        } 
    }()
    
    handler(ctx)
    //handle session close
    session := ctx.Session()
    if session!=nil {
        session.Close()
    }
    
    return nil
}

func handlePageNotFound(ctx WebContext){
    isOK := false
    handler := getHandlerThouActionsMap("error", "pagenotfound")
    if handler != nil{
        err := safelyCall(*handler, ctx)
        if err == nil {
            isOK = true
        }
    }
    
    if !isOK {
        ctx.Abort(404, "Page not found")
    }
}

func getHandlerThouActionsMap(section, action string) *func(WebContext) {
    
        secVal, secExist := ActionsMap[section]
        if !secExist {
            return nil
        }
        
        handler, exist := secVal[action]
        if !exist {
            return nil
        }
        
        return &handler
}

func fileExists(filePath string) bool {
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return false
    }

    return !fileInfo.IsDir()
}

type routeStruct struct {
    r       string
    cr      *regexp.Regexp
    handler func(WebContext)
}

var routes []routeStruct

func AddRoute(r string, handler func(WebContext)) {
    cr, err := regexp.Compile(r)
    if err != nil {
        relog.Warning("AddRoute: route (" + r + ") regex compile error:" + err.Error())
        return
    }

    routes = append(routes, routeStruct{r, cr, handler})
}






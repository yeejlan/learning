
package kiss

import(
    "net/http"  
    "net/url"
    "time"
    "fmt"
    "strings"
)

type WebContext map[string]interface{}

func NewWebContext(request *http.Request, response http.ResponseWriter) WebContext {
    return WebContext{
        "kiss.request" : request,
        "kiss.response" : response,
        "kiss.params" : url.Values{},
        "kiss.cache" : map[string]interface{}{},
    }
}

func (this WebContext) Request() *http.Request{
    return this["kiss.request"].(*http.Request)
}

func (this WebContext) Response() http.ResponseWriter{
    return this["kiss.response"].(http.ResponseWriter)
}

func (this WebContext) Params() url.Values{
    return this["kiss.params"].(url.Values)
}

func (this WebContext) Session() ISessionStorage {
    s := this["kiss.session"]
    if s!=nil {
        return s.(ISessionStorage)
    }
    
    return nil
}

func (this WebContext) Cache() map[string]interface{}{
    return this["kiss.cache"].(map[string]interface{})
}

func (this WebContext) SetHeader(key string, val string) {
    this.Response().Header().Set(key, val)
}

func (this WebContext) AddHeader(key string, val string) {
    this.Response().Header().Add(key, val)
}

//set cookie. ageSeconds: 0 = forever, -1 = current opened window
func (this WebContext) SetCookie(name string, value string, ageSeconds int64) {
    var utctime time.Time
    if ageSeconds == 0 {
        // 2^31 - 1 seconds (roughly 2038)
        utctime = time.Unix(2147483647, 0)
    } else if ageSeconds > 0 {
        utctime = time.Unix(time.Now().Unix()+ageSeconds, 0)
    }
    
    var cookieStr string
    if ageSeconds < 0 {
        cookieStr = fmt.Sprintf("%s=%s", name, value)
    }else{
        cookieStr = fmt.Sprintf("%s=%s; expires=%s", name, value, webTime(utctime))
    }
    
    if App.Domain() !="" {
        cookieStr += fmt.Sprintf("; domain=%s", App.Domain())
    }
    
    this.AddHeader("Set-Cookie", cookieStr)
}

func (this WebContext) GetCookie(name string) string {
    cookie, err := this.Request().Cookie(name)
    if err!=nil {
        return ""
    }
    
    return cookie.Value
}

func (this WebContext) Write(content []byte) (int, error) {
    s := this.Session()
    if s!= nil {
        s.Write()
    }
    return this.Response().Write(content)
}

func (this WebContext) WriteString(content string) {
    this.Write([]byte(content))
}

func (this WebContext) WriteHeader(status int) {
    this.Response().WriteHeader(status)
}

func (this WebContext) Abort(status int, body string) {
    this.Response().WriteHeader(status)
    this.Response().Write([]byte(body))
}

func (this WebContext) Redirect(url string) {
    this.Response().Header().Set("Location", url)
    this.Response().WriteHeader(302)
    this.Response().Write([]byte("Redirecting to: " + url))
}

func (this WebContext) Render(tplFile string, data interface{}) {
    Render(this, tplFile, data)
}

func webTime(t time.Time) string {
    ftime := t.Format(time.RFC1123)
    if strings.HasSuffix(ftime, "UTC") {
        ftime = ftime[0:len(ftime)-3] + "GMT"
    }
    return ftime
}

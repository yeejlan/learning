
package controller

import(
    "../kiss"
)

func init(){
    var actions = map[string]func(kiss.WebContext){
        "set" : TestSet,
        "get" : TestGet,
    }
    kiss.ExposeActions("test", actions)
}


func TestSet(ctx kiss.WebContext){
    ctx.Session().Set("hello", "hihi~~")
}

func TestGet(ctx kiss.WebContext) {
    ctx.WriteString("session: " + ctx.Session().Get("hello"))
}









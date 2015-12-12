
package controller

import(
    "../kiss"
    "fmt"
)

func init(){
    var actions = map[string]func(kiss.WebContext){
        "list" : UserList,
        "hello" : UserHello,
    }
    kiss.ExposeActions("user", actions)
}


func UserList(ctx kiss.WebContext){
    var str = fmt.Sprintf("%#v", ctx)
    ctx.WriteString(str)    
    ctx.WriteString("<br />\r\nuser list")
}

func UserHello(ctx kiss.WebContext) {
    ctx.WriteString("user hello~~~")
}







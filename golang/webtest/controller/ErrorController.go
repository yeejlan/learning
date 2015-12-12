
package controller

import(
    "../kiss"
)


func init(){
    kiss.ExposeAction("error", "pagenotfound", ErrorPageNotFound)
}

func ErrorPageNotFound(ctx kiss.WebContext){
    ctx.WriteHeader(404)
    ctx.WriteString("page not found~~~")
}


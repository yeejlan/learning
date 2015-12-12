
package controller

import(
    "../kiss"
)


func init(){
    kiss.ExposeAction("home", "index", HomeIndex)
}

func HomeIndex(ctx kiss.WebContext){
    ctx.WriteString("home index~~")
}


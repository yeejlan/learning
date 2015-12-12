
package main

import (
    "./kiss"
    _ "./controller"
    //"fmt"
)


func main() {
	
	kiss.AddRoute("/hello(.*)", hello)
	kiss.AddRoute("/hi(.*)", hi)
	kiss.StartWebServer(":8080", "http")

	kiss.App.Shutdown()
}

func hello(ctx kiss.WebContext){

    haha := ctx.GetCookie("haha") 
    ctx.WriteString(haha)       
    ctx.Render("user/userlist.html", 12345)
}

func hi(ctx kiss.WebContext){

    
    ctx.SetCookie("aa", "aaa", -1)
    ctx.SetCookie("cc", "ccc", -1)
    
    ctx.Session().Set("hello", "kitty~")   

    c := ctx.GetCookie(kiss.GetSessionName())
    ctx.WriteString(c)
    ctx.WriteString("<br>\r\nh!~~~~~~~~~~")
    
}
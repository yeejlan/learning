
package main

import (
    "./kiss"
    "os"
    "fmt"
    _ "./controller"
)


func main() {
    kiss.AddRoute("/hi(.*)", hi)
    kiss.ServeFcgi("127.0.0.1:8000")
}

func hi(ctx kiss.WebContext){

    pid := os.Getpid()
    ctx.WriteString(fmt.Sprintf("%d called\r\n", pid))
    //kiss.Render(ctx, "user/userlist.html", 12345)
}
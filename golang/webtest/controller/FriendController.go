
package controller

import(
    "../kiss"
)


func init(){
    kiss.ExposeAction("friend", "list", FriendList)
    kiss.ExposeAction("friend", "error", FriendError)
}

func FriendList(ctx kiss.WebContext){
    ctx.WriteString("friend list")
}


func FriendError(ctx kiss.WebContext){
    ctx.WriteString("friend error")
}

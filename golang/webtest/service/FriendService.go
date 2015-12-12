
package service

import(
    "../kiss"
    "fmt"
)

type Friend struct{}

func NewFriendService() *Friend {
	return &Friend{}
}

func (this Friend) Hello(){
	fmt.Println("friendservice:hello", kiss.App.Env)
	
    userServ := NewUserService()
    fmt.Println("userServ:add from friendServ", userServ.Add(5,6))
}
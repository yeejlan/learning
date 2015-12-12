
package service

type User struct{}

func NewUserService() *User{
	return &User{}
}

func (this User) Add(x, y int) int {
	return x+y
}
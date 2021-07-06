package chapter3

//go:generate ../../array_helper . User
type User struct {
	ID   uint64
	Name string
	Age  int32
}

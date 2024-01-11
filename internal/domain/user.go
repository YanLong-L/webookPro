package domain

type User struct {
	Id       int64
	Email    string
	Password string
	Utime    int64
	Ctime    int64
}

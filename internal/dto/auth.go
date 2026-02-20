package dto

type RegisterIn struct {
	Email    string
	Password string
	Role     string
}

type LoginIn struct {
	Email    string
	Password string
}

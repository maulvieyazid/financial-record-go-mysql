package entities

type User struct {
	Id       string
	Name     string `validate:"required" label:"Nama"`
	Email    string `validate:"required,email"`
	Password string `validate:"omitempty,min=6"`
	Photo    *string
}

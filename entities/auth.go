package entities

type Register struct {
	Id              string
	Name            string `validate:"required" label:"Nama"`
	Email           string `validate:"required,email,isunique=users-email"`
	Password        string `validate:"required,min=6"`
	ConfirmPassword string `validate:"required,min=6,eqfield=Password" label:"Konfirmasi Password"`
}

type Auth struct {
	Id       string
	Name     string
	Email    string `validate:"required"`
	Password string `validate:"required,min=6"`
}



package models

import (
	"database/sql"
	"financial-record/entities"
)

type AuthModel struct {
	db *sql.DB
}

func NewAuthModel(db *sql.DB) *AuthModel {
	return &AuthModel{
		db: db,
	}
}

func (model AuthModel) Register(user entities.Register) error {

	_, err := model.db.Exec(
		"INSERT INTO users (id, name, email, password) VALUES (?,?,?,?)",
		user.Id, user.Name, user.Email, user.Password,
	)

	return err
}

func (model AuthModel) Login(email string) (entities.Auth, error) {

	var user entities.Auth
	query := "SELECT id, email, name, password FROM users WHERE email = ?"

	err := model.db.QueryRow(query, email).Scan(
		&user.Id,
		&user.Email,
		&user.Name,
		&user.Password,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}

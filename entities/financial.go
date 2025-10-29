package entities

import "time"

type AddFinancial struct {
	Id          int16
	UserId      string
	Date        time.Time `validate:"required" label:"Tanggal"`
	Type        string    `validate:"required"`
	Nominal     int64     `validate:"required,numeric"`
	Category    string    `validate:"required" label:"Kategori"`
	Description *string
	Attachment  *string
}

type Financial struct {
	Id          int16
	UserId      string
	Date        time.Time
	Type        string
	Nominal     int64
	Category    string
	Description *string
	Attachment  *string
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

package models

import (
	"database/sql"
	"financial-record/entities"
	"time"
)

type FinancialModel struct {
	db *sql.DB
}

func NewFinancalModel(db *sql.DB) *FinancialModel {
	return &FinancialModel{
		db: db,
	}
}

func (model FinancialModel) AddFinacialRecord(data entities.AddFinancial) error {

	query := `
		INSERT INTO record (user_id, date, type, category, nominal, description, attachment)
		VALUES (?,?,?,?,?,?,?)
	`

	_, err := model.db.Exec(
		query,
		data.UserId,
		data.Date,
		data.Type,
		data.Category,
		data.Nominal,
		data.Description,
		data.Attachment,
	)

	return err
}

func (model FinancialModel) GetFinancialTotalNominal(user_id string, monthYear string, pemasukanOnly bool, pengeluaranOnly bool) (total_pemasukan int64, total_pengeluaran int64, err error) {

	parsedDate, _ := time.Parse("January 2006", monthYear)
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'pemasukan' THEN nominal ELSE 0 END), 0) AS total_pemasukan,
			COALESCE(SUM(CASE WHEN type = 'pengeluaran' THEN nominal ELSE 0 END), 0) AS total_pengeluaran
		FROM record
		WHERE user_id = ?
		AND MONTH(date) = ?
	    AND YEAR(date) = ?
	`

	if pemasukanOnly {
		query += " AND type = 'pemasukan'"
	}
	if pengeluaranOnly {
		query += " AND type = 'pengeluaran'"
	}

	err = model.db.QueryRow(query, user_id, parsedDate.Month(), parsedDate.Year()).Scan(&total_pemasukan, &total_pengeluaran)

	if err != nil {
		return 0, 0, err
	}

	return total_pemasukan, total_pengeluaran, nil
}

func (model FinancialModel) FindAllFinancial(user_id string, monthYear string, pemasukanOnly bool, pengeluaranOnly bool) ([]entities.Financial, error) {

	// mysql
	parsedDate, _ := time.Parse("January 2006", monthYear)
	query := `
	    SELECT id, date, type, category, nominal, description, attachment
	    FROM record
	    WHERE user_id = ?
	    AND MONTH(date) = ?
	    AND YEAR(date) = ?
	`

	if pemasukanOnly {
		query += " AND type = 'pemasukan'"
	}
	if pengeluaranOnly {
		query += " AND type = 'pengeluaran'"
	}

	rows, err := model.db.Query(query, user_id, parsedDate.Month(), parsedDate.Year())
	if err != nil {
		return []entities.Financial{}, err
	}

	defer rows.Close()

	var financials []entities.Financial
	for rows.Next() {
		var financial entities.Financial
		err := rows.Scan(
			&financial.Id,
			&financial.Date,
			&financial.Type,
			&financial.Category,
			&financial.Nominal,
			&financial.Description,
			&financial.Attachment,
		)
		if err != nil {
			return []entities.Financial{}, err
		}
		financials = append(financials, financial)
	}

	return financials, rows.Err()
}

func (model FinancialModel) DeleteFinancialRecord(id int16) error {

	query := "DELETE FROM record WHERE id = ?"

	_, err := model.db.Exec(query, id)

	return err
}

func (model FinancialModel) FindFinancialById(id int16) (*entities.Financial, error) {

	financial := &entities.Financial{}

	query := `
		SELECT id, date, type, category, nominal, description, attachment
		FROM record WHERE id = ?
	`

	err := model.db.QueryRow(query, id).Scan(
		&financial.Id,
		&financial.Date,
		&financial.Type,
		&financial.Category,
		&financial.Nominal,
		&financial.Description,
		&financial.Attachment,
	)

	if err != nil {
		return nil, err
	}
	return financial, nil
}

func (model FinancialModel) EditFinancialRecord(data entities.AddFinancial) error {

	query := `
		UPDATE record SET 
		date = ?, 
		type = ?, 
		category = ?, 
		nominal = ?, 
		description = ?, 
		attachment = ?, 
		updated_at = ? 
		WHERE id = ?
	`

	_, err := model.db.Exec(
		query,
		data.Date,
		data.Type,
		data.Category,
		data.Nominal,
		data.Description,
		data.Attachment,
		time.Now(),
		data.Id,
	)

	return err
}

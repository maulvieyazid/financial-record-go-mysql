package helpers

import (
	"database/sql"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type Validation struct {
	db *sql.DB
}

func NewValidator(db *sql.DB) *Validation {
	return &Validation{
		db: db,
	}
}

func initValidator(validation *Validation) (*validator.Validate, ut.Translator) {

	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()

	// register translation
	en_translations.RegisterDefaultTranslations(validate, trans)

	// set label field
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		labelName := field.Tag.Get("label")
		if labelName == "" {
			return field.Name
		}
		return labelName
	})

	// custom translate required
	validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} tidak boleh kosong", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})

	// custom translate email
	validate.RegisterTranslation("email", trans, func(ut ut.Translator) error {
		return ut.Add("email", "{0} harus berupa email yang valid", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("email", fe.Field())
		return t
	})

	// custom translate min
	validate.RegisterTranslation("min", trans, func(ut ut.Translator) error {
		return ut.Add("min", "{0} minimal {1} karakter", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("min", fe.Field(), fe.Param())
		return t
	})

	// custom translate eqfield
	validate.RegisterTranslation("eqfield", trans, func(ut ut.Translator) error {
		return ut.Add("eqfield", "{0} harus sama dengan {1}", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("eqfield", fe.Field(), fe.Param())
		return t
	})

	// register isunique
	validate.RegisterValidation("isunique", func(fl validator.FieldLevel) bool {
		param := fl.Param()
		splitParam := strings.Split(param, "-")

		tableName := splitParam[0]
		fieldName := splitParam[1]
		fieldValue := fl.Field().String()

		return validation.checkIsUnique(tableName, fieldName, fieldValue)
	})

	// custom translate isunique
	validate.RegisterTranslation("isunique", trans, func(ut ut.Translator) error {
		return ut.Add("isunique", "{0} sudah digunakan", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("isunique", fe.Field())
		return t
	})

	return validate, trans
}

func (validation *Validation) checkIsUnique(tableName, fieldName, fieldValue string) bool{

	query := "SELECT " + fieldName + " FROM " + tableName + " WHERE " + fieldName + " = ?"
	row := validation.db.QueryRow(query, fieldValue)

	var result string
	err := row.Scan(&result)
	return err == sql.ErrNoRows
}

func (validation *Validation) Struct(s interface{}) interface{} {

	validate, trans := initValidator(validation)
	var vErrors = make(map[string]interface{})

	if err := validate.Struct(s); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			vErrors[e.StructField()] = e.Translate(trans)
		}
	}

	if len(vErrors) > 0 {
		return vErrors
	}

	return nil
}

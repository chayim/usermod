package usermod

import "database/sql"

func CreateAllTables(db *sql.DB) []error {
	u := User{db: db}
	uot := UserOperationToken{db: db}

	var errors []error
	err := u.CreateTable()
	errors = append(errors, err)

	err = uot.CreateTable()
	errors = append(errors, err)
	return errors
}

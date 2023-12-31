package user

import (
	"database/sql"
	"fmt"

	t "github.com/sikozonpc/notebase/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(user t.User) error {
	_, err := s.db.Exec("INSERT INTO users (firstName, lastName, email, password) VALUES (?, ?, ?, ?)", user.FirstName, user.LastName, user.Email, user.Password)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetUserByEmail(email string) (*t.User, error) {
	rows, err := s.db.Query("SELECT * FROM users WHERE email = ?", email)
	if err != nil {
		return nil, err
	}

	u := new(t.User)
	for rows.Next() {
		u, err = scanRowsIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}

func (s *Store) GetUserByID(id int) (*t.User, error) {
	rows, err := s.db.Query("SELECT * FROM users WHERE id = ? AND isActive = 1", id)
	if err != nil {
		return nil, err
	}

	u := new(t.User)
	for rows.Next() {
		u, err = scanRowsIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}

	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return u, nil
}

func (s *Store) GetUsers() ([]*t.User, error) {
	rows, err := s.db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}

	users := make([]*t.User, 0)
	for rows.Next() {
		u, err := scanRowsIntoUser(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}

func (s *Store) UpdateUser(user t.User) error {
	_, err := s.db.Exec("UPDATE users SET firstName = ?, lastName = ?, email = ?, password = ?, isActive = ? WHERE id = ?", user.FirstName, user.LastName, user.Email, user.Password, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func scanRowsIntoUser(rows *sql.Rows) (*t.User, error) {
	user := new(t.User)

	err := rows.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

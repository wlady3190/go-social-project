package store

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  Password `json:"-"`
	CreatedAt string   `json:"created_at"`
}

type Password struct {
	text *string //* se asigna puntero para que pueda ser null
	hash []byte
}

func (p *Password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &text
	p.hash = hash

	return nil

}

type UserStore struct {
	db *sql.DB
}

// * implementando transaccion
func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `INSERT INTO users (username, password, email) 
			VALUES ($1, $2, $3)
			RETURNING id, created_at`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	// err := s.db.QueryRowContext(ctx, query,
	err := tx.QueryRowContext(ctx, query,
		user.Username,
		user.Password,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique contraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique contraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err

		}
	}
	return nil
}

func (s *UserStore) GetById(ctx context.Context, userID int64) (*User, error) {
	query := `
	SELECT id, username, email, password, created_at
	from users
	WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil

}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	//transaction wrapper
	//* se está capturando la funcion de Create al agregar la tx
	return withTX(s.db, ctx, func(tx *sql.Tx) error {
		//* create user
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		//! create user invitation, private method, no one can access to this method outside

		if err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}
		return nil
	})

}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	query := `
	INSERT INTO user_invitations 
	(token, user_id, expiry) values ($1, $2, $3)
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))

	if err != nil {
		return err
	}
	return nil
}

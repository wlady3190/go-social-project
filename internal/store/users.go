package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  Password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
	Role      Role     `json:"role"`
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

func (p *Password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

type UserStore struct {
	db *sql.DB
}

// * implementando transaccion
func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `INSERT INTO users (username, password, email, role_id) 
			VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4 ) )
			RETURNING id, created_at`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	//! User default
	role := user.Role.Name
	if role == "" {
		role = "user"
	}
	// err := s.db.QueryRowContext(ctx, query,
	err := tx.QueryRowContext(ctx, query,
		user.Username,
		//! va el password hasheado
		user.Password.hash,
		user.Email,
		// user.RoleID,
		role,
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
	SELECT users.id, username, email, password, created_at, roles.*
	FROM users	
	JOIN roles ON (users.role_id=roles.id)
	WHERE users.id = $1
	AND is_active = true
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
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

func (s *UserStore) Activate(ctx context.Context, token string) error {
	return withTX(s.db, ctx, func(tx *sql.Tx) error {
		// find th user than the token belongs to
		user, err := s.getUserFromInvitation(ctx, tx, token)

		if err != nil {
			return err
		}
		// update user -> is_actrive = true
		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		// clean invitation
		if err := s.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
	SELECT u.id, u.username, u.email, u.created_at, u.is_active
	FROM users u
	JOIN user_invitations ui 
	ON u.id = ui.user_id
	WHERE ui.token = $1 AND ui.expiry > $2;	
	`

	hash := sha256.Sum256([]byte(token))

	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
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

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		UPDATE users SET username = $1,
						 email = $2,
						 is_active = $3
		WHERE id = $4;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.IsActive,
		user.ID,
	)
	if err != nil {
		return err
	}

	return nil

}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil

}

func (s *UserStore) Delete(ctx context.Context, userID int64) error {
	return withTX(s.db, ctx, func(tx *sql.Tx) error {

		if err := s.delete(ctx, tx, userID); err != nil {
			return err
		}

		if err := s.deleteUserInvitations(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})

}

func (s *UserStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {

	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil

}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {

	query := `
	SELECT id, username, email, password, created_at  FROM users
	WHERE email = $1 AND  is_active=true
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	user := &User{}

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
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

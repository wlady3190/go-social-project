package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"` //esto se vincula desde el query de GetPostByID, tb se añade al Post
}

type CommentStore struct {
	db *sql.DB
}

// !Esta funcion se debe añadir al Repository.
// * Luego en post, se puede incluir esta función para traer el post y los comentarios. Ir a PostHandler
func (s *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	query := ` SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.username, users.id
				FROM comments c
	 			JOIN users ON users.id = c.user_id
				WHERE c.post_id = $1
				ORDER BY c.created_at DESC`
	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var c Comment
		c.User = User{}
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.User.Username, &c.User.ID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil

}

func (s *CommentStore) Create(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (post_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	// ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	// defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.UserID,
		comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

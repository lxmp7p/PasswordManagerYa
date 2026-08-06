package repository

import (
	"context"
)

type userRepository struct {
	db DBinterface
}

func NewUserRepository(db DBinterface) UserRepositoryInterface {
	return &userRepository{
		db: db,
	}
}

var _ UserRepositoryInterface = (*userRepository)(nil)

func (r *userRepository) Create(
	ctx context.Context,
	login string,
	passwordHash string,
) error {

	_, err := r.db.Exec(
		ctx,
		`INSERT INTO users(login, password_hash)
		 VALUES($1, $2)`,
		login,
		passwordHash,
	)

	return err
}

func (r *userRepository) FindByLogin(
	ctx context.Context,
	login string,
) (*User, error) {

	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`SELECT id, login, password_hash
		 FROM users
		 WHERE login=$1`,
		login,
	).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

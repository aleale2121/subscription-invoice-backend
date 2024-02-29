package db

import (
	"context"
	"database/sql"
	"log"
	"subscription-service/internal/constants/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserPersistence struct {
	db *sql.DB
}

// NewUsersPersistence is the function used to create an instance of the UserPersistence.
func NewUsersPersistence(dbPool *sql.DB) UserPersistence {
	return UserPersistence{db: dbPool}
}

// GetAll returns a slice of all users, sorted by last name
func (u *UserPersistence) GetAll() ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, active, created_at, updated_at
	from users order by last_name`

	rows, err := u.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User

	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Password,
			&user.Active,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning", err)
			return nil, err
		}

		users = append(users, &user)
	}

	return users, nil
}

// GetByEmail returns one user by email
func (u *UserPersistence) GetByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, active, created_at, updated_at from users where email = $1`

	var user models.User
	row := u.db.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetOne returns one user by id
func (u *UserPersistence) GetOne(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, active, created_at, updated_at from users where id = $1`

	var user models.User
	row := u.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update updates one user in the database, using the information
// stored in the receiver u
func (u *UserPersistence) Update(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `update users set
		email = $1,
		first_name = $2,
		last_name = $3,
		active = $4,
		updated_at = $5
		where id = $6
	`

	_, err := u.db.ExecContext(ctx, stmt,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Active,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes one user from the database, by User.ID
func (u *UserPersistence) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from users where id = $1`

	_, err := u.db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new user into the database, and returns the ID of the newly inserted row
func (u *UserPersistence) AddUser(user models.User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return 0, err
	}

	var newID int
	stmt := `insert into users (email, first_name, last_name, password, active, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7) returning id`

	err = u.db.QueryRowContext(ctx, stmt,
		user.Email,
		user.FirstName,
		user.LastName,
		hashedPassword,
		user.Active,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// ResetPassword is the method we will use to change a user's password.
func (u *UserPersistence) ResetPassword(password string, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `update users set password = $1 where id = $2`
	_, err = u.db.ExecContext(ctx, stmt, hashedPassword, userID)
	if err != nil {
		return err
	}

	return nil
}

// AddBillingAddress inserts a new billing address into the database
func (u *UserPersistence) AddBillingAddress(address models.Address) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newID int
	stmt := `INSERT INTO billing_address (user_id, address, address_2, postal_code, city, country)
        VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
    `

	err := u.db.QueryRowContext(ctx, stmt, address.UserID, address.Address, address.Address2, address.PostalCode, address.PostalCode, address.PostalCode).Scan(&newID)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

// GetBillingAddressByUserID returns the billing address for the given user ID
func (u *UserPersistence) GetBillingAddressByUserID(userID int) (*models.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, address, address_2, postal_code, city, country FROM billing_address WHERE user_id = $1`

	var billingAddress models.Address
	row := u.db.QueryRowContext(ctx, query, userID)

	err := row.Scan(&billingAddress.ID, &billingAddress.Address, &billingAddress.Address2, &billingAddress.PostalCode, &billingAddress.City, &billingAddress.Country)
	if err != nil {
		return nil, err
	}

	return &billingAddress, nil
}

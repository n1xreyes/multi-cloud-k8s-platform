package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Config contains the PostgreSQL connection parameters
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Client represents a PostgreSQL client
type Client struct {
	db *sql.DB
}

// NewClient creates a new PostgreSQL client
func NewClient(ctx context.Context, config Config) (*Client, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Verify the connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &Client{db: db}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// User represents a user in the database
type User struct {
	ID        int       `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// CreateUser creates a new user in the database
func (c *Client) CreateUser(ctx context.Context, user *User, passwordHash string) error {
	query := `
		INSERT INTO users (username, email, password_hash, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	row := c.db.QueryRowContext(
		ctx, query,
		user.Username, user.Email, passwordHash, user.FirstName, user.LastName,
	)

	return row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetUserByUsername retrieves a user by username
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, email, first_name, last_name, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &User{}
	err := c.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return user, nil
}

// APIKey represents an API key in the database
type APIKey struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	KeyHash   string    `db:"key_hash"`
	Name      string    `db:"name"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// CreateAPIKey creates a new API key for a user
func (c *Client) CreateAPIKey(ctx context.Context, apiKey *APIKey) error {
	query := `
		INSERT INTO api_keys (user_id, key_hash, name, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	row := c.db.QueryRowContext(
		ctx, query,
		apiKey.UserID, apiKey.KeyHash, apiKey.Name, apiKey.ExpiresAt,
	)

	return row.Scan(&apiKey.ID, &apiKey.CreatedAt, &apiKey.UpdatedAt)
}

// GetAPIKeyByHash retrieves an API key by its hash
func (c *Client) GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, expires_at, created_at, updated_at
		FROM api_keys
		WHERE key_hash = $1
	`

	apiKey := &APIKey{}
	err := c.db.QueryRowContext(ctx, query, keyHash).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.KeyHash, &apiKey.Name, &apiKey.ExpiresAt,
		&apiKey.CreatedAt, &apiKey.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, err
	}

	return apiKey, nil
}

// CloudCredential represents cloud provider credentials in the database
type CloudCredential struct {
	ID          int       `db:"id"`
	UserID      int       `db:"user_id"`
	Provider    string    `db:"provider"`
	Name        string    `db:"name"`
	Credentials string    `db:"credentials"` // JSON string
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// CreateCloudCredential creates new cloud provider credentials
func (c *Client) CreateCloudCredential(ctx context.Context, cred *CloudCredential) error {
	query := `
		INSERT INTO cloud_credentials (user_id, provider, name, credentials)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	row := c.db.QueryRowContext(
		ctx, query,
		cred.UserID, cred.Provider, cred.Name, cred.Credentials,
	)

	return row.Scan(&cred.ID, &cred.CreatedAt, &cred.UpdatedAt)
}

// GetCloudCredentialsByUserID retrieves all cloud credentials for a user
func (c *Client) GetCloudCredentialsByUserID(ctx context.Context, userID int) ([]CloudCredential, error) {
	query := `
		SELECT id, user_id, provider, name, credentials, created_at, updated_at
		FROM cloud_credentials
		WHERE user_id = $1
	`

	rows, err := c.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []CloudCredential
	for rows.Next() {
		var cred CloudCredential
		err := rows.Scan(
			&cred.ID, &cred.UserID, &cred.Provider, &cred.Name, &cred.Credentials,
			&cred.CreatedAt, &cred.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, cred)
	}

	return credentials, rows.Err()
}

// LogAuditEvent logs an audit event to the database
func (c *Client) LogAuditEvent(ctx context.Context, userID int, action, resourceType, resourceName, namespace string, requestData string, status, message, clientIP string) error {
	query := `
		INSERT INTO audit_logs (user_id, action, resource_type, resource_name, namespace, request_data, status, message, client_ip)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := c.db.ExecContext(
		ctx, query,
		userID, action, resourceType, resourceName, namespace, requestData, status, message, clientIP,
	)
	return err
}

// ExecuteInTransaction executes the provided function within a transaction
func (c *Client) ExecuteInTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			_ = tx.Rollback()
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			// Rollback on error
			_ = tx.Rollback()
		} else {
			// Commit if no error
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

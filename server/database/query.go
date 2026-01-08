package database

import (
	"context"
)

// User operations

func (db *DB) CreateUser(ctx context.Context) (string, error) {
	query := `INSERT INTO users (api_key, user_id, created_at, last_accessed) 
	          VALUES (gen_random_uuid(), gen_random_uuid(), NOW(), NOW()) 
	          RETURNING api_key`

	var apiKey string
	err := db.Pool.QueryRow(ctx, query).Scan(&apiKey)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func (db *DB) GetUserID(ctx context.Context, apiKey string) (string, error) {
	var userID string
	query := "SELECT user_id FROM users WHERE api_key = $1"
	err := db.Pool.QueryRow(ctx, query, apiKey).Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (db *DB) GetAPIKey(ctx context.Context, userID string) (string, error) {
	var apiKey string
	query := "SELECT api_key FROM users WHERE user_id = $1"
	err := db.Pool.QueryRow(ctx, query, userID).Scan(&apiKey)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func (db *DB) DeleteUser(ctx context.Context, apiKey string) error {
	query := "DELETE FROM users WHERE api_key = $1"
	_, err := db.Pool.Exec(ctx, query, apiKey)
	return err
}

func (db *DB) UpdateLastAccessed(ctx context.Context, apiKey string) error {
	query := "UPDATE users SET last_accessed = NOW() WHERE api_key = $1"
	_, err := db.Pool.Exec(ctx, query, apiKey)
	return err
}

func (db *DB) UpdateLastAccessedByUserID(ctx context.Context, userID string) error {
	query := "UPDATE users SET last_accessed = NOW() WHERE user_id = $1"
	_, err := db.Pool.Exec(ctx, query, userID)
	return err
}

// Delete operations

func (db *DB) DeleteRequests(ctx context.Context, apiKey string) error {
	query := "DELETE FROM requests WHERE api_key = $1"
	_, err := db.Pool.Exec(ctx, query, apiKey)
	return err
}

func (db *DB) DeleteMonitors(ctx context.Context, apiKey string) error {
	query := "DELETE FROM monitor WHERE api_key = $1"
	_, err := db.Pool.Exec(ctx, query, apiKey)
	return err
}

func (db *DB) DeletePings(ctx context.Context, apiKey string) error {
	query := "DELETE FROM pings WHERE api_key = $1"
	_, err := db.Pool.Exec(ctx, query, apiKey)
	return err
}

func (db *DB) DeleteURLMonitor(ctx context.Context, apiKey string, url string) error {
	query := "DELETE FROM monitor WHERE api_key = $1 AND url = $2"
	_, err := db.Pool.Exec(ctx, query, apiKey, url)
	return err
}

func (db *DB) DeleteURLPings(ctx context.Context, apiKey string, url string) error {
	query := "DELETE FROM pings WHERE api_key = $1 AND url = $2"
	_, err := db.Pool.Exec(ctx, query, apiKey, url)
	return err
}

// Transaction-based delete operations

func (db *DB) DeleteAllData(ctx context.Context, apiKey string) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM requests WHERE api_key = $1", apiKey); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM monitor WHERE api_key = $1", apiKey); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM pings WHERE api_key = $1", apiKey); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (db *DB) DeleteUserAccount(ctx context.Context, apiKey string) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete all related data
	if _, err := tx.Exec(ctx, "DELETE FROM requests WHERE api_key = $1", apiKey); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM monitor WHERE api_key = $1", apiKey); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM pings WHERE api_key = $1", apiKey); err != nil {
		return err
	}

	// Delete the user
	if _, err := tx.Exec(ctx, "DELETE FROM users WHERE api_key = $1", apiKey); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Connection check

func (db *DB) CheckConnection(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// User agent operations

func (db *DB) GetUserAgent(ctx context.Context, userAgentID int) (string, error) {
	var userAgent string
	query := "SELECT user_agent FROM user_agents WHERE id = $1"
	err := db.Pool.QueryRow(ctx, query, userAgentID).Scan(&userAgent)
	if err != nil {
		return "", err
	}

	return userAgent, nil
}

func (db *DB) GetUserAgents(ctx context.Context, userAgentIDs map[int]struct{}) (map[int]string, error) {
	userAgents := make(map[int]string)
	if len(userAgentIDs) == 0 {
		return userAgents, nil
	}

	// Convert map keys to slice
	ids := make([]int, 0, len(userAgentIDs))
	for id := range userAgentIDs {
		ids = append(ids, id)
	}

	query := "SELECT id, user_agent FROM user_agents WHERE id = ANY($1)"
	rows, err := db.Pool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var userAgent string
		if err := rows.Scan(&id, &userAgent); err != nil {
			return nil, err
		}
		userAgents[id] = userAgent
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userAgents, nil
}
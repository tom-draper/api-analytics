package database

import (
	"context"
	"github.com/jackc/pgx/v5"
)

func CreateUser() (string, error) {
	conn, err := NewConnection()
	if err != nil {
		return "", err
	}
	defer conn.Close(context.Background())

	return CreateUserWithConnection(context.Background(), conn)
}

func CreateUserWithConnection(ctx context.Context, conn *pgx.Conn) (string, error) {
	query := "INSERT INTO users (api_key, user_id, created_at, last_accessed) VALUES (gen_random_uuid(), gen_random_uuid(), NOW(), NOW()) RETURNING api_key;"

	var apiKey string
	err := conn.QueryRow(ctx, query).Scan(&apiKey)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func GetUserID(apiKey string) (string, error) {
	conn, err := NewConnection()
	if err != nil {
		return "", err
	}
	defer conn.Close(context.Background())

	return GetUserIDWithConnection(context.Background(), conn, apiKey)
}

func GetUserIDWithConnection(ctx context.Context, conn *pgx.Conn, apiKey string) (string, error) {
	// Fetch user ID associated with API key
	var userID string
	query := "SELECT user_id FROM users WHERE api_key = $1;"
	err := conn.QueryRow(ctx, query, apiKey).Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func GetAPIKey(userID string) (string, error) {
	conn, err := NewConnection()
	if err != nil {
		return "", err
	}
	defer conn.Close(context.Background())

	return GetAPIKeyWithConnection(context.Background(), conn, userID)
}

func GetAPIKeyWithConnection(ctx context.Context, conn *pgx.Conn, userID string) (string, error) {
	// Fetch API key associated with user ID
	var apiKey string
	query := "SELECT api_key FROM users WHERE user_id = $1;"
	err := conn.QueryRow(ctx, query, userID).Scan(&apiKey)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func DeleteUser(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return DeleteUserWithConnection(context.Background(), conn, apiKey)
}

func DeleteUserWithConnection(ctx context.Context, conn *pgx.Conn, apiKey string) error {
	query := "DELETE FROM users WHERE api_key = $1;"
	_, err := conn.Exec(ctx, query, apiKey)
	return err
}

func DeleteRequests(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return DeleteRequestsWithConnection(context.Background(), conn, apiKey)
}

func DeleteRequestsWithConnection(ctx context.Context, conn *pgx.Conn, apiKey string) error {
	query := "DELETE FROM requests WHERE api_key = $1;"
	_, err := conn.Exec(ctx, query, apiKey)
	return err
}

func DeleteMonitors(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return DeleteMonitorsWithConnection(context.Background(), conn, apiKey)
}

func DeleteMonitorsWithConnection(ctx context.Context, conn *pgx.Conn, apiKey string) error {
	query := "DELETE FROM monitor WHERE api_key = $1;"
	_, err := conn.Exec(ctx, query, apiKey)
	return err
}

func DeletePings(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return DeletePingsWithConnection(context.Background(), conn, apiKey)
}

func DeletePingsWithConnection(ctx context.Context, conn *pgx.Conn, apiKey string) error {
	query := "DELETE FROM pings WHERE api_key = $1;"
	_, err := conn.Exec(ctx, query, apiKey)
	return err
}

func DeleteURLMonitor(apiKey string, url string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return DeleteURLMonitorWithConnection(context.Background(), conn, apiKey, url)
}

func DeleteURLMonitorWithConnection(ctx context.Context, conn *pgx.Conn, apiKey string, url string) error {
	query := "DELETE FROM monitor WHERE api_key = $1 AND url = $2;"
	_, err := conn.Exec(ctx, query, apiKey, url)
	return err
}

func DeleteURLPings(apiKey string, url string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return DeleteURLPingsWithConnection(context.Background(), conn, apiKey, url)
}

func DeleteURLPingsWithConnection(ctx context.Context, conn *pgx.Conn, apiKey string, url string) error {
	query := "DELETE FROM pings WHERE api_key = $1 AND url = $2;"
	_, err := conn.Exec(ctx, query, apiKey, url)
	return err
}

func DeleteAllData(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return DeleteAllDataWithConnection(context.Background(), conn, apiKey)
}

func DeleteAllDataWithConnection(ctx context.Context, conn *pgx.Conn, apiKey string) error {
	err := DeleteRequestsWithConnection(ctx, conn, apiKey)
	if err != nil {
		return err
	}

	err = DeleteMonitorsWithConnection(ctx, conn, apiKey)
	if err != nil {
		return err
	}

	err = DeletePingsWithConnection(ctx, conn, apiKey)
	if err != nil {
		return err
	}

	return nil
}

func DeleteUserAccount(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return DeleteUserAccountWithConnection(context.Background(), conn, apiKey)
}

func DeleteUserAccountWithConnection(ctx context.Context, conn *pgx.Conn, apiKey string) error {
	err := DeleteAllDataWithConnection(ctx, conn, apiKey)
	if err != nil {
		return err
	}

	err = DeleteUserWithConnection(ctx, conn, apiKey)
	if err != nil {
		return err
	}

	return nil
}

func CheckConnection() error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return CheckConnectionWithContext(context.Background(), conn)
}

func CheckConnectionWithContext(ctx context.Context, conn *pgx.Conn) error {
	// Simple query to check connection
	err := conn.Ping(ctx)
	if err != nil {
		return err
	}

	return nil
}

func UpdateLastAccessed(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}

	defer conn.Close(context.Background())

	return UpdateLastAccessedWithConnection(context.Background(), conn, apiKey)
}

func UpdateLastAccessedWithConnection(ctx context.Context, conn *pgx.Conn, apiKey string) error {
	query := "UPDATE users SET last_accessed = NOW() WHERE api_key = $1;"
	_, err := conn.Exec(context.Background(), query, apiKey)
	return err
}

func UpdateLastAccessedByUserID(userID string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}

	defer conn.Close(context.Background())

	return UpdateLastAccessedByUserIDWithConnection(context.Background(), conn, userID)
}

func UpdateLastAccessedByUserIDWithConnection(ctx context.Context, conn *pgx.Conn, userID string) error {
	query := "UPDATE users SET last_accessed = NOW() WHERE user_id = $1;"
	_, err := conn.Exec(context.Background(), query, userID)
	return err
}

func GetUserAgent(userAgentID int) (string, error) {
	conn, err := NewConnection()
	if err != nil {
		return "", err
	}
	defer conn.Close(context.Background())

	return GetUserAgentWithConnection(context.Background(), conn, userAgentID)
}

func GetUserAgentWithConnection(ctx context.Context, conn *pgx.Conn, userAgentID int) (string, error) {
	var userAgent string
	query := "SELECT user_agent FROM user_agents WHERE id = $1;"
	err := conn.QueryRow(ctx, query, userAgentID).Scan(&userAgent)
	if err != nil {
		return "", err
	}

	return userAgent, nil
}

func GetUserAgents(userAgentIDs map[int]struct{}) (map[int]string, error) {
	conn, err := NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	return GetUserAgentsWithConnection(context.Background(), conn, userAgentIDs)
}

func GetUserAgentsWithConnection(ctx context.Context, conn *pgx.Conn, userAgentIDs map[int]struct{}) (map[int]string, error) {
	userAgents := make(map[int]string)
	if len(userAgentIDs) == 0 {
		return userAgents, nil
	}

	// Convert map keys to slice for easier handling
	ids := make([]int, 0, len(userAgentIDs))
	for id := range userAgentIDs {
		ids = append(ids, id)
	}

	// Use pgx's built-in placeholder generation
	query := "SELECT id, user_agent FROM user_agents WHERE id = ANY($1)"

	rows, err := conn.Query(ctx, query, ids)
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

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userAgents, nil
}

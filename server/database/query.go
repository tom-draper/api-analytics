package database

import "context"

func CreateUser() (string, error) {
	connection, err := NewConnection()
	if err != nil {
		return "", err
	}
	defer connection.Close(context.Background())

	query := "INSERT INTO users (api_key, user_id, created_at, last_accessed) VALUES (gen_random_uuid(), gen_random_uuid(), NOW(), NOW()) RETURNING api_key;"

	var apiKey string
	err = connection.QueryRow(context.Background(), query).Scan(&apiKey)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func GetUserID(apiKey string) (string, error) {
	connection, err := NewConnection()
	if err != nil {
		return "", err
	}
	defer connection.Close(context.Background())

	// Fetch user ID associated with API key
	var userID string
	query := "SELECT user_id FROM users WHERE api_key = $1;"
	err = connection.QueryRow(context.Background(), query, apiKey).Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func GetAPIKey(userID string) (string, error) {
	connection, err := NewConnection()
	if err != nil {
		return "", err
	}
	defer connection.Close(context.Background())

	// Fetch API key associated with user ID
	var apiKey string
	query := "SELECT api_key FROM users WHERE user_id = $1;"
	err = connection.QueryRow(context.Background(), query, userID).Scan(&apiKey)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func DeleteUser(apiKey string) error {
	connection, err := NewConnection()
	if err != nil {
		return err
	}
	defer connection.Close(context.Background())

	query := "DELETE FROM users WHERE api_key = $1;"
	_, err = connection.Exec(context.Background(), query, apiKey)
	return err
}

func DeleteRequests(apiKey string) error {
	connection, err := NewConnection()
	if err != nil {
		return err
	}
	defer connection.Close(context.Background())

	query := "DELETE FROM requests WHERE api_key = $1;"
	_, err = connection.Exec(context.Background(), query, apiKey)
	return err
}

func DeleteMonitors(apiKey string) error {
	connection, err := NewConnection()
	if err != nil {
		return err
	}
	defer connection.Close(context.Background())

	query := "DELETE FROM monitor WHERE api_key = $1;"
	_, err = connection.Exec(context.Background(), query, apiKey)
	return err
}

func DeletePings(apiKey string) error {
	connection, err := NewConnection()
	if err != nil {
		return err
	}
	defer connection.Close(context.Background())

	query := "DELETE FROM pings WHERE api_key = $1;"
	_, err = connection.Exec(context.Background(), query, apiKey)
	return err
}

func CheckConnection() error {
	connection, err := NewConnection()
	if err != nil {
		return err
	}
	defer connection.Close(context.Background())

	// Simple query to check connection
	err = connection.Ping(context.Background())
	if err != nil {
		return err
	}

	return nil
}

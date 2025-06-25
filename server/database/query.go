package database

import "context"

func DeleteUser(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := "DELETE FROM users WHERE api_key = $1;"
	_, err = conn.Exec(context.Background(), query, apiKey)
	return err
}

func DeleteRequests(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := "DELETE FROM requests WHERE api_key = $1;"
	_, err = conn.Exec(context.Background(), query, apiKey)
	return err
}

func DeleteMonitors(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := "DELETE FROM monitor WHERE api_key = $1;"
	_, err = conn.Exec(context.Background(), query, apiKey)
	return err
}

func DeletePings(apiKey string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := "DELETE FROM pings WHERE api_key = $1;"
	_, err = conn.Exec(context.Background(), query, apiKey)
	return err
}

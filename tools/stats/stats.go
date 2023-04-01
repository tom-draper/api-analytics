package main

func UserUsage() {
	db := OpenDBConnection()

	query := "SELECT *, COUNT(*) OVER(PARTITION BY api_key) count FROM requests;"
	_, err := db.Query(query)
	if err != nil {
		panic(err)
	}
}

func TopUsers() {

}

func DeleteUser() {

}

func DeleteUserRequests() {

}

func DailyUsage() {

}

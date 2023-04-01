package lib

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"
)

func getSupabaseLogin() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	return supabaseURL, supabaseKey
}

type SupabaseRequestRow struct {
	RequestRow
	RequestID int       `json:"request_id"`
	CreatedAt time.Time `json:"created_at"`
}

func readRequests() []SupabaseRequestRow {
	f, err := os.Open("supabase_tyirpladmhanzkwhmspj_New Query (1).csv")
	if err != nil {
		log.Fatal("Unable to read input file ", err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.LazyQuotes = true
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for ", err)
	}

	// "user_agent","method","created_at","path","api_key","status","response_time","request_id","framework","hostname","ip_address","location"
	result := make([]SupabaseRequestRow, 0)
	for _, record := range records {
		r := new(SupabaseRequestRow)
		r.APIKey = record[0]
		method, _ := strconv.Atoi(record[1])
		r.Method = int16(method)
		r.CreatedAt, err = time.Parse("2006-01-02 15:04:05.000000+00", record[2])
		if err != nil {
			r.CreatedAt, err = time.Parse("2006-01-02 15:04:05.00000+00", record[2])
			if err != nil {
				r.CreatedAt, err = time.Parse("2006-01-02 15:04:05.0000+00", record[2])
				if err != nil {
					r.CreatedAt, err = time.Parse("2006-01-02 15:04:05.000+00", record[2])
					if err != nil {
						r.CreatedAt, err = time.Parse("2006-01-02 15:04:05.00+00", record[2])
						if err != nil {
							r.CreatedAt, err = time.Parse("2006-01-02 15:04:05.00+00", record[2])
							if err != nil {
								r.CreatedAt, err = time.Parse("2006-01-02 15:04:05.0+00", record[2])
								if err != nil {
									fmt.Println(record[2], r.CreatedAt)
								}
							}
						}
					}
				}
			}
		}

		r.Path = record[3]
		r.APIKey = record[4]
		status, _ := strconv.Atoi(record[5])
		r.Status = int16(status)
		responseTime, _ := strconv.Atoi(record[6])
		r.ResponseTime = int16(responseTime)
		r.RequestID, _ = strconv.Atoi(record[7])
		framework, _ := strconv.Atoi(record[8])
		r.Framework = int16(framework)
		r.Hostname = record[9]
		r.IPAddress = record[10]
		r.Location = record[11]
		result = append(result, *r)
	}

	return result
}

func migrateSupabaseRequests() {
	// var result []SupabaseRequestRow
	// err := supabase.DB.From("Requests").Select("*").Execute(&result)
	// if err != nil {
	// 	panic(err)
	// }

	result := readRequests()

	var db *sql.DB
	var query bytes.Buffer
	i := 0
	for _, request := range result {
		if i > 0 {
			query.WriteString(",")
		} else {
			query = bytes.Buffer{}
			query.WriteString("INSERT INTO requests (api_key, method, created_at, path, status, response_time, framework, hostname, ip_address, location) VALUES")
		}
		if request.IPAddress == "" || request.IPAddress == "testclient" {
			query.WriteString(fmt.Sprintf(" ('%s', %d, '%s', '%s', %d, %d, %d, '%s', NULL, '%s')", request.APIKey, request.Method, request.CreatedAt.UTC().Format(time.RFC3339), request.Path, request.Status, request.ResponseTime, request.Framework, request.Hostname, request.Location))
		} else {
			query.WriteString(fmt.Sprintf(" ('%s', %d, '%s', '%s', %d, %d, %d, '%s', '%s', '%s')", request.APIKey, request.Method, request.CreatedAt.UTC().Format(time.RFC3339), request.Path, request.Status, request.ResponseTime, request.Framework, request.Hostname, request.IPAddress, request.Location))
		}

		i++
		if i == 10000 {
			query.WriteString(";")

			fmt.Println("Write to database")

			db = OpenDBConnection()
			_, err := db.Query(query.String())
			if err != nil {
				panic(err)
			}
			i = 0
			db.Close()
			time.Sleep(10 * time.Second)
		}
	}

	db = OpenDBConnection()
	query.WriteString(";")

	fmt.Println("Final write to database")

	_, err := db.Query(query.String())
	if err != nil {
		panic(err)
	}
	fmt.Println("Complete")
}

type SupabaseUsersRow struct {
	User
}

func migrateSupabaseUsers(db *sql.DB, supabase *supa.Client) {
	var result []SupabaseUsersRow
	err := supabase.DB.From("Users").Select("*").Execute(&result)
	if err != nil {
		panic(err)
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO users (api_key, user_id, created_at) VALUES")
	for i, user := range result {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', '%s')", user.APIKey, user.UserID, user.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err = db.Query(query.String())
	if err != nil {
		panic(err)
	}
}

type SupabaseMonitorRow struct {
	MonitorRow
	CreatedAt time.Time `json:"created_at"`
}

func migrateSupabaseMonitors(db *sql.DB, supabase *supa.Client) {
	var result []SupabaseMonitorRow
	err := supabase.DB.From("Monitor").Select("*").Execute(&result)
	if err != nil {
		panic(err)
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO monitor (api_key, url, secure, ping, created_at) VALUES")
	for i, monitor := range result {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', %t, %t, '%s')", monitor.APIKey, monitor.URL, monitor.Secure, monitor.Ping, monitor.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err = db.Query(query.String())
	if err != nil {
		panic(err)
	}
}

type SupabasePingsRow struct {
	PublicPingsRow
	APIKey string `json:"api_key"`
}

func migrateSupabasePings(db *sql.DB, supabase *supa.Client) {
	var result []SupabasePingsRow
	err := supabase.DB.From("Pings").Select("*").Execute(&result)
	if err != nil {
		panic(err)
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO pings (api_key, url, response_time, status, created_at) VALUES")
	for i, monitor := range result {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', %d, %d, '%s')", monitor.APIKey, monitor.URL, monitor.ResponseTime, monitor.Status, monitor.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err = db.Query(query.String())
	if err != nil {
		panic(err)
	}
}

func MigrateSupabaseData() {
	// supabaseURL, supabaseKey := getSupabaseLogin()
	// supabase := supa.CreateClient(supabaseURL, supabaseKey)

	// db := OpenDBConnection()

	migrateSupabaseRequests()
}

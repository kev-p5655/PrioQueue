package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const JOB_TABLE_NAME string = "jobs"

type Job struct {
	id          int
	description string
}

func createDb() (*sql.DB, error) {
	const file string = "foo.db"
	db, err := sql.Open("sqlite3", file)
	return db, err
}

func createTables(db *sql.DB) error {
	var init string = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
		id INTEGER NOT NULL PRIMARY KEY,
		description TEXT
		);`,
		JOB_TABLE_NAME,
	)
	_, err := db.Exec(init)
	return err
}

func initDb() (*sql.DB, error) {
	db, err := createDb()
	if err != nil {
		return nil, err
	}
	err = createTables(db)
	if err != nil {
		return nil, err
	}
	return db, err
}

func createJobInsertQuery() (query string) {
	query = fmt.Sprintf(`
		insert into %s
		values (1, "hello"), (5, "goodbye")
		`,
		JOB_TABLE_NAME,
	)
	return
}

func getAllJobs(db *sql.DB) (jobs []Job, err error) {
	query := fmt.Sprintf(`
		select id, description
		from %s
		`,
		JOB_TABLE_NAME,
	)
	rows, err := db.Query(query)
	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.id, &job.description); err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}
	return
}

func exec(db *sql.DB, query string) (err error) {
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return err
}

func main() {
	db, err := initDb()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = exec(db, createJobInsertQuery())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	jobs, err := getAllJobs(db)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, job := range jobs {
		fmt.Println(job)
	}
}

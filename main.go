package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const JOB_TABLE_NAME string = "jobs"

type Job struct {
	id          int
	description string
	priority    int
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
		description TEXT,
		priority INTEGER NOT NULL
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

func createJobInsertQuery(db *sql.DB) (query string, err error) {
	// Create the elements to insert
	descriptions := []string{"hello", "goodbye"}
	currPrio, err := getCurrPrio(db)
	if err != nil {
		return
	}

	currPrio++
	fmt.Println(descriptions)
	items := []string{}
	for i, description := range descriptions {
		fmt.Println(description)
		items = append(items,
			fmt.Sprintf(`("%s", %d)`, description, currPrio+i),
		)
	}

	// Create the query
	query = fmt.Sprintf(`
		INSERT INTO %s (description, priority)
		VALUES %s
		;`,
		JOB_TABLE_NAME,
		strings.Join(items, ","),
	)
	return
}

func getCurrPrio(db *sql.DB) (prio int, err error) {
	query := fmt.Sprintf(`
		SELECT priority
		FROM %s
		ORDER BY priority DESC
		LIMIT 1
		;`,
		JOB_TABLE_NAME,
	)
	err = db.QueryRow(query).Scan(&prio)
	if err == sql.ErrNoRows {
		// Handle ErrNoRows
		err = nil
		prio = 0
	}
	return
}

func getAllJobs(db *sql.DB) (jobs []Job, err error) {
	query := fmt.Sprintf(`
		SELECT id, description, priority
		FROM %s
		ORDER by priority ASC
		`,
		JOB_TABLE_NAME,
	)
	rows, err := db.Query(query)
	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.id, &job.description, &job.priority); err != nil {
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
	// Setup
	db, err := initDb()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()

	// Queries
	query, err := createJobInsertQuery(db)
	fmt.Println(query)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = exec(db, query)
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

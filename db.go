package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const JOB_TABLE_NAME string = "jobs"

type Job struct {
	Id          int        `json:"id"`
	Description string     `json:"description"`
	Priority    int        `json:"priority"`
	FinishedAt  *time.Time `json:"finished_at"`
}

func createDb() (*sql.DB, error) {
	const file string = "jobs.db"
	db, err := sql.Open("sqlite3", file)
	return db, err
}

func createTables(db *sql.DB) error {
	var init string = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
		id INTEGER NOT NULL PRIMARY KEY,
		description TEXT,
		priority INTEGER NOT NULL,
		finished_at DATETIME
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

func createJobInsertQuery(db *sql.DB, descriptions []string) (query string, err error) {
	currPrio, err := getCurrPrio(db)
	if err != nil {
		return
	}
	currPrio++

	items := []string{}
	for i, description := range descriptions {
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

func createJobs(db *sql.DB, descriptions []string) (jobs []Job, err error) {
	// TODO: Make this return the jobs correctly
	query, err := createJobInsertQuery(db, descriptions)
	if err != nil {
		return
	}
	_, err = exec(db, query)
	return
}

func updateJobPriority(db *sql.DB, id int, priority int) (job Job, err error) {
	// TODO: Make this return the jobs correctly
	query := fmt.Sprintf(`
		UPDATE %s
		SET priority = %d
		WHERE id = %d
		;`,
		JOB_TABLE_NAME,
		priority,
		id,
	)

	_, err = exec(db, query)

	// Both of these are 0 <nil> if no rows were updated.
	// fmt.Println(result.LastInsertId())
	// fmt.Println(result.RowsAffected())
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
	// Define to be empty array, instead of nil.
	jobs = []Job{}
	query := fmt.Sprintf(`
		SELECT id, description, priority, finished_at
		FROM %s
		ORDER by priority ASC
		`,
		JOB_TABLE_NAME,
	)
	rows, err := db.Query(query)
	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.Id, &job.Description, &job.Priority, &job.FinishedAt); err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}
	return
}

func exec(db *sql.DB, query string) (result sql.Result, err error) {
	stmt, err := db.Prepare(query)
	if err != nil {
		return
	}
	result, err = stmt.Exec()
	return
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const JOB_TABLE_NAME string = "jobs"

// TODO: Could add most of these functions to an interface, so it's easier to swap the sqlite implementation with something else.
type Job struct {
	id          int
	description string
	priority    int
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

// Http related code. This probably should get moved to another module?
func handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "hello world\n")
	}
}

func addRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/hello", handleHello())
}

func run(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	stdin io.Reader,
	stdout, strerr io.Writer,
) error {
	mux := http.NewServeMux()
	addRoutes(mux)
	err := http.ListenAndServe(":8080", mux)
	return err
}

func setupLogging() *slog.Logger {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// logger := log.New(os.Stdout, "my:", log.LstdFlags)
	return logger
}

func main() {
	// Setup
	logger := setupLogging()

	// Kinda feels like a lot of the db stuff should be moved elsewhere. Like into the "run" function.
	//	I guess if I actually setup the http stuff. This should be moved to run, also handlers should be setup with a middleware that connects to the db instead of needing to pass in the connection?
	db, err := initDb()
	defer db.Close()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Queries
	query, err := createJobInsertQuery(db)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	err = exec(db, query)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	jobs, err := getAllJobs(db)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	for _, job := range jobs {
		logger.Info("Obtained job", slog.Any("job", job))
	}

	// // Run/setup http server
	// ctx := context.Background()
	// err = run(ctx, nil, nil, nil, nil, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(2)
	// }
}

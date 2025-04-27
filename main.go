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
	"time"

	docs "PrioQueue/docs"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const JOB_TABLE_NAME string = "jobs"

// TODO: Could add most of these functions to an interface, so it's easier to swap the sqlite implementation with something else.
type Job struct {
	Id          int       `json:"id"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	FinishedAt  time.Time `json:"finished_at"`
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
		SELECT id, description, priority
		FROM %s
		ORDER by priority ASC
		`,
		JOB_TABLE_NAME,
	)
	rows, err := db.Query(query)
	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.Id, &job.Description, &job.Priority); err != nil {
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

// Http related code. This probably should get moved to another module?

//	@BasePath	/api/v1
//	@Summary	Create jobs
//	@Schemes
//	@Description	Creates jobs
//	@Tags			Jobs
//	@Accept			json
//	@Produce		json
//	@Param			jobDescriptions	body	[]string	true	"An array of Job descriptions"
//	@Success		200				{array}	Job
//	@Router			/jobs [post]
func handleCreateJobs(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newJobs []string
		if err := c.BindJSON(&newJobs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		query, err := createJobInsertQuery(db, newJobs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		_, err = exec(db, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		fmt.Println(newJobs)
		c.JSON(http.StatusOK, gin.H{
			"message": "Need to return the jobs that are created.",
		})
		// TODO: Need to get the result from the execution of the query, and return all the new job records.
		//		This kinda shows how the structure of the code currently sucks. B/c a lot of this logic should be handled by some "service" that handles all the interaction with the database. Instead of in the handler function.
	}
}

//	@BasePath	/api/v1
//	@Summary	Get all jobs
//	@Schemes
//	@Description	Gets all jobs ordered by priority
//	@Tags			Jobs
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	Job
//	@Router			/jobs [get]
func handleListJobs(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobs, err := getAllJobs(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, jobs)
	}
}

func addRoutes(r *gin.RouterGroup, db *sql.DB, logger *slog.Logger) {
	r.GET("/jobs", handleListJobs(db))
	r.POST("/jobs", handleCreateJobs(db))
}

func run(
	ctx context.Context,
	logger *slog.Logger,
	args []string,
	getenv func(string) string,
	stdin io.Reader,
	stdout, strerr io.Writer,
) error {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	docs.SwaggerInfo.Title = "http Priority Queue"
	docs.SwaggerInfo.Description = "This api has endpoints for managing a priority queue"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"

	db, err := initDb()
	if err != nil {
		// This would be a big problem
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close() // Some of the db stuff could get moved elsewhere, but I don't know how to correctly defer closing it if it's in another function?

	v1 := r.Group("/api/v1")
	addRoutes(v1, db, logger)

	err = r.Run("localhost:8080")
	return err
}

func setupLogging() *slog.Logger {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return logger
}

func main() {
	// Setup
	logger := setupLogging()
	ctx := context.Background()

	// Run http server
	err := run(ctx, logger, nil, nil, nil, nil, nil)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(2)
	}
}

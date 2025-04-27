package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @BasePath	/api/v1
// @Summary	Create jobs
// @Schemes
// @Description	Creates jobs
// @Tags			Jobs
// @Accept			json
// @Produce		json
// @Param			jobDescriptions	body	[]string	true	"An array of Job descriptions"
// @Success		200				{array}	Job
// @Router			/jobs [post]
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

// @BasePath	/api/v1
// @Summary	Get all jobs
// @Schemes
// @Description	Gets all jobs ordered by priority
// @Tags			Jobs
// @Accept			json
// @Produce		json
// @Success		200	{array}	Job
// @Router			/jobs [get]
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

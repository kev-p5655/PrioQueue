package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func checkBadRequestError(c *gin.Context, err error) bool {
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return true
	}
	return false
}

func checkInternalError(c *gin.Context, err error) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return true
	}
	return false
}

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
func handleCreateJobs(db *sql.DB, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newJobs []string
		if checkBadRequestError(c, c.BindJSON(&newJobs)) {
			return
		}

		_, err := createJobs(db, newJobs)
		if checkInternalError(c, err) {
			logger.Error("Error creating jobs", slog.String("error", err.Error()))
			return
		}

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
		if checkInternalError(c, err) {
			return
		}
		c.JSON(http.StatusOK, jobs)
	}
}

// @BasePath	/api/v1
// @Summary	Get job by id
// @Schemes
// @Description	Gets a job by it's ID
// @Tags			Jobs
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"Id of job"
// @Success		200	{object}	Job
// @Router			/jobs/{id} [get]
func handleGetJobById(
	db *sql.DB,
	logger *slog.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if checkBadRequestError(c, err) {
			logger.Error("Error parsing param id", slog.String("error", err.Error()))
			return
		}

		job, err := getJobById(db, id)
		if checkInternalError(c, err) {
			logger.Error("Error when getting job by id", slog.String("error", err.Error()))
			return
		}

		if len(job) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No job found with id",
			})
			return
		}

		c.JSON(http.StatusOK, job[0])
	}
}

// @BasePath	/api/v1
// @Summary	Update job priority
// @Schemes
// @Description	Updates a jobs priority
// @Tags			Jobs
// @Accept			json
// @Produce		json
// @Param			id			path		int	true	"The id of the job being updated"
// @Param			priority	query		int	true	"The new priority"
// @Success		200			{object}	Job
// @Router			/jobs/{id} [patch]
func handleUpdatePriority(
	db *sql.DB,
	logger *slog.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if checkBadRequestError(c, err) {
			logger.Error("Error parsing param id", slog.String("error", err.Error()))
			return
		}

		priority, err := strconv.Atoi(c.Query("priority"))
		if checkBadRequestError(c, err) {
			logger.Error("Error parsing param priority", slog.String("error", err.Error()))
			return
		}

		_, err = updateJobPriority(db, id, priority)
		if checkInternalError(c, err) {
			logger.Error("Error updating job priority", slog.String("error", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Need to return the updated job struct",
		})
	}
}

func addRoutes(r *gin.RouterGroup, db *sql.DB, logger *slog.Logger) {
	r.GET("/jobs", handleListJobs(db))
	r.GET("/jobs/:id", handleGetJobById(db, logger))
	r.POST("/jobs", handleCreateJobs(db, logger))
	r.PATCH("/jobs/:id", handleUpdatePriority(db, logger))
}

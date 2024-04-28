package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"log/slog"
	"os"
	"rms/internals/auth"
	"rms/internals/handlers"
	"rms/internals/middlewares"
	"rms/internals/services"
)

func main() {
	db, err := sql.Open("mysql", "root:chirag@tcp(localhost:3306)/rms")

	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("database Connected")
	defer db.Close()

	// Set up user service
	us, err := services.NewUserService(db)
	if err != nil {
		log.Panic(err)
	}
	js := services.NewJobService(db)

	rs, err := services.NewResumeService(db)
	if err != nil {
		log.Panic(err)
	}

	// Setup authentication using RSA keys
	privatePem, err := os.ReadFile("private.pem")
	if err != nil {
		log.Panic(err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePem)
	if err != nil {
		log.Panic(err)
	}

	publicPEM, err := os.ReadFile("pubkey.pem")
	if err != nil {
		log.Panic("not able to read pem file")
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicPEM)
	if err != nil {
		log.Panic(err)
	}

	a, err := auth.NewAuth(publicKey, privateKey)
	if err != nil {
		log.Panic(err)
	}

	// Setup middleware using the authentication service
	m, err := middlewares.NewMid(a)
	if err != nil {
		log.Panic(err)
	}

	usersC, err := handlers.NewUsers(us, a)
	if err != nil {
		log.Panic(err)
	}

	jobC := handlers.NewJobHandler(js, a)

	resumeC, err := handlers.NewResume(rs, a)
	if err != nil {
		log.Panic(err)
	}
	setupSlog()
	r := gin.Default()

	r.POST("/api/signup", usersC.CreateUser)
	r.POST("/api/login", usersC.ProcessLoginIn)
	r.POST("/api/admin/job", m.JWTMiddleware(jobC.CreateJob, auth.Admin))
	r.POST("/api/upload", m.JWTMiddleware(resumeC.UploadResume, auth.Applicant))
	r.GET("/api/jobs", m.JWTMiddleware(jobC.GetJobs, auth.Applicant))
	r.GET("/api/jobs/apply/:job_id", m.JWTMiddleware(jobC.JobApplyHandler, auth.Applicant))
	r.GET("/api/admin/job/:job_id", m.JWTMiddleware(jobC.GetJobAndApplicantsHandler, auth.Admin))
	r.GET("/api/admin/applicants", m.JWTMiddleware(usersC.GetApplicantsHandler, auth.Admin))
	r.GET("/api/admin/applicant/:applicant_id", m.JWTMiddleware(usersC.GetApplicantsHandler, auth.Admin))

	r.Run(":8080")

}

func setupSlog() {
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		//AddSource: true: This will cause the source dir and line number of the log message to be included in the output
		AddSource: true,
	})

	logger := slog.New(logHandler)
	//SetDefault makes l the default Logger. in our case we would be doing structured logging
	slog.SetDefault(logger)
}

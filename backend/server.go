package backend

import (
	"context"
	"flag"
	"log"
	"main/backend/api"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func Run() {
	host := flag.String("host", "localhost", "host to run the server on")
	port := flag.String("port", "8080", "port to run the server on")
	docker := flag.Bool("docker", false, "whether to run in docker mode")
	flag.Parse()

	if err := api.InitDatabase(); err != nil {
		log.Fatalln(err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/healthcheck", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		apiGroup.POST("/contact", api.PostContact)
	}

	// Static files
	r.Use(static.Serve("/", static.LocalFile("./build", true)))
	r.NoRoute(func(c *gin.Context) {
		c.File("./build/index.html")
	})

	var serverPath string
	if *docker {
		serverPath = "0.0.0.0:8080"
		log.Println("Server running in docker mode on port 8080")
	} else {
		serverPath = *host + ":" + *port
		log.Println("Server running at http://" + serverPath)
	}

	server := &http.Server{
		Addr:         serverPath,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}

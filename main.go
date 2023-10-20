package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

const (
	FORTUNE              = "FORTUNE"
	PORT                 = "PORT"
	PARAM_NUM_FORTUNE    = "count"
	DEFAULT_FORTUNE_FILE = "./fortune.txt"
	DEFAULT_STATIC_DIR = "./static"
	DEFAULT_PORT         = 3000
	DEFAULT_NUM_FORTUNE  = "1"
)

func loadFortunes(path string) []string {
	buff, err := ioutil.ReadFile(path)
	if nil != err {
		log.Fatalf("Error reading %s: %v\n", path, err)
	}
	lines := strings.Split(string(buff), "\n")
	return lines[:len(lines)-1]
}

func checkStaticAsset(path string) {
	if _, err := os.Stat(path); nil != err {
		log.Fatalf("Static directory '%s' error : %s ", path, err)
	}
}

func defaultFortune() string {
	value, present := os.LookupEnv(FORTUNE)
	if present {
		return value
	}
	return DEFAULT_FORTUNE_FILE
}

func defaultPort() (int, error) {
	value, present := os.LookupEnv(PORT)
	if present {
		return strconv.Atoi(value)
	}
	return DEFAULT_PORT, nil
}

func getFortunes(fortune []string, count int) []string {
	idx := rand.Perm(len(fortune))[:count]
	f := make([]string, count)
	for i := 0; i < count; i++ {
		f[i] = fortune[idx[i]]
	}
	return f
}

func mkHandler(fortunes []string) func(*gin.Context) {
	return func(c *gin.Context) {

		count, err := strconv.Atoi(c.DefaultQuery(PARAM_NUM_FORTUNE, DEFAULT_NUM_FORTUNE))
		if nil != err {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}

		f := getFortunes(fortunes, count)
		t, _ := time.Now().MarshalText()
		c.JSON(http.StatusOK, gin.H{
			"timestamp": string(t),
			"fortunes":  f,
		})
	}
}

func mkMVCHandler(fortunes []string) func(*gin.Context) {
	return func(c *gin.Context) {
		f := getFortunes(fortunes, 1)
		c.HTML(http.StatusOK, "index", gin.H{ "fortuneText": f[0] })
	}
}

func healthz(c *gin.Context) {
	t, _ := time.Now().MarshalText()
	c.JSON(http.StatusOK, gin.H{"timestamp": string(t)})
}

func notFound(c *gin.Context) {

	if strings.Contains(c.GetHeader("Accept"), "text/html") {
		c.Redirect(http.StatusPermanentRedirect, "/static/404.html")
		return
	}

	t, _ := time.Now().MarshalText()
	c.JSON(http.StatusNotFound, gin.H{
		"timestamp": string(t),
		"error":     fmt.Sprintf("Resource not found: %s", c.Request.URL.String()),
	})
}

func main() {

	var fortuneFile string
	var port int
	var staticDir string
	defPort, err := defaultPort()

	if nil != err {
		log.Fatalf("Error: %v", err)
	}

	flag.StringVar(&fortuneFile, "fortune", defaultFortune(), "Fortune file")
	flag.IntVar(&port, "port", defPort, "port")
	flag.StringVar(&staticDir, "static", DEFAULT_STATIC_DIR, "Static resources directory")
	flag.Parse()

	log.Printf("fortune file: %s, static directory: %s, port: %d", fortuneFile, staticDir, port)

	checkStaticAsset(staticDir)

	fortunes := loadFortunes(fortuneFile)
	log.Printf("Loaded %s fortunes file\n", fortuneFile)

	rand.Seed(time.Now().UnixNano())

	r := gin.Default()

	r.HTMLRender = ginview.Default()

	r.GET("/", mkMVCHandler(fortunes))
	r.GET("/api/fortune", mkHandler(fortunes))

	r.GET("/healthz", healthz)

	r.Use(static.Serve("/static", static.LocalFile(staticDir, true)))

	r.Use(notFound)

	log.Printf("Starting server on port %d\n", port)
	if err := r.Run(fmt.Sprintf("0.0.0.0:%d", port)); nil != err {
		log.Panicf("Cannot start server. %v\n", err)
	}

}

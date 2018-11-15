package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var (
	forceDb       bool
	forceTemplate bool
	debug         bool
	listenPort    int
	logP          Println
	logF          Printf
)

const (
	layoutPath = "templates/layout.html"
)

func init() {
	flag.BoolVar(&debug, "d", false, "show debug information")
	flag.BoolVar(&forceDb, "fd", false, "forcefully recreate database tables if already exist")
	flag.BoolVar(&forceTemplate, "ft", false, "forcefully recreate templates if already exist")
	flag.IntVar(&listenPort, "p", 8080, "listen port")
	flag.Parse()

	initLog()

	connDB()
}

func main() {
	parsePageCfg()
	genTemplates()
	createTables()
	defer closeDB()

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Static("/static", "./static")

	router.HTMLRender = createRender()

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("session", store))

	router.GET("/logout", logout)
	router.GET("login", loginPage)
	router.POST("/login", postLogin)
	router.GET("/page/:tmpl", pageTemplate)
	router.POST("/addapi", add)
	router.POST("/delapi", del)
	router.GET("/listapi", list)
	router.POST("/updateapi", update)
	router.GET("/", home)

	addr := fmt.Sprintf(":%d", listenPort)
	if !debug {
		log.Printf("Listening and serving HTTP on %s\n", addr)
	}
	router.Run(addr)
}

func initLog() {
	if debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		logP = log.Println
		logF = log.Printf
	} else {
		logP = func(...interface{}) {}
		logF = func(string, ...interface{}) {}
	}
}

func createRender() multitemplate.Renderer {
	tmpls := loadTemplatesFiles()
	r := multitemplate.NewRenderer()
	r.AddFromFiles("login.html", "templates/login.html")
	r.AddFromFiles("index.html", layoutPath, "templates/index.html")

	for _, v := range tmpls {
		r.AddFromFiles(filepath.Base(v), layoutPath, v)
	}
	return r
}

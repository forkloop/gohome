package main

import (
	"fmt"
	"flag"
	"os/exec"
	"path/filepath"
	"path"
	"os"
	"time"
	"log"
	"io/ioutil"
	"html/template"
	"github.com/gin-gonic/gin"
)

var (
	uploadDir string
	username string
	password string
)


func main() {
	port := flag.Int("p", 4000, "port")
	flag.Parse()

	r := gin.Default()

	r.Static("/public", "public")

	r.GET("/", func(c *gin.Context) {
		r.SetHTMLTemplate(template.Must(template.ParseFiles("templates/base.tmpl", "templates/home.tmpl")))
		c.HTML(200, "base", nil)
	})

	r.GET("/control", func(c *gin.Context) {
		r.SetHTMLTemplate(template.Must(template.ParseFiles("templates/base.tmpl", "templates/control.tmpl")))
		c.HTML(200, "base", nil)
	})

	r.POST("/control", gin.BasicAuth(gin.Accounts{username: password}), func(c *gin.Context) {
		err := c.Request.ParseForm()
		if err != nil {
			log.Printf("Error while parsing control request: %v", err)
			c.Fail(500, err)
			return
		}

		mode := c.Request.Form["mode"]

		log.Printf("Request to turn [%v] motion.", mode)
		var command *exec.Cmd
		if len(mode) > 0 && mode[0] == "on" {
			//command = exec.Command("service", "motion", "start")
			command = exec.Command("ls")
		} else {
			//command = exec.Command("service", "motion", "stop")
			command = exec.Command("ls")
		}

		err = command.Run()
		if err != nil {
			log.Printf("Error while %v motion, %v", mode, err)
			c.Fail(500, err)
			return
		}

		if accept := c.Request.Header["Accept"]; len(accept) > 0 && accept[0] == "application/json" {
			c.JSON(200, gin.H{"mode": mode})
		} else {
			c.Redirect(301, "/control")
		}
	})

	r.GET("/tartan/*time", gin.BasicAuth(gin.Accounts{username: password}), func(c *gin.Context) {
		r.SetHTMLTemplate(template.Must(template.ParseFiles("templates/base.tmpl", "templates/tartan.tmpl")))
		log.Println(c.Params.ByName("time"))
		t := c.Params.ByName("time")
		if t == "/" {
			t = time.Now().Format("2006-01-02")
		}

		log.Printf("query %v", t)
		photos := []string{}
		for _, p := range listPhotos(path.Join(uploadDir, t)) {
			photos = append(photos, path.Join(t, path.Base(p)))
		}
		c.HTML(200, "base", gin.H{"photos": photos})
	})

	r.GET("/photos/:date/:name", gin.BasicAuth(gin.Accounts{username: password}), func(c *gin.Context) {
		p := path.Join(uploadDir, c.Params.ByName("date"), c.Params.ByName("name"))
		log.Printf("request photo %v", p)
		c.File(p)
	})

	r.POST("/tartan", func(c *gin.Context) {
		defer c.Request.Body.Close()
		err := c.Request.ParseMultipartForm(5242880) // 5MB
		if err != nil {
			log.Printf("Error while parsing uploaded file: %v", err.Error())
			c.Fail(500, err)
			return
		}

		// save file
		form := c.Request.MultipartForm
		if form != nil {
			for _, fileHeader := range form.File {
				file, err := fileHeader[0].Open()
				if err != nil {
					log.Printf("%v", err)
					continue
				}
				data, err := ioutil.ReadAll(file)
				if err != nil {
					continue
				}
				saveUpload(data, fileHeader[0].Filename)
			}
		}
		c.String(200, "")
	})

	r.Run(fmt.Sprintf(":%v", *port))
}

func init() {
	uploadDir = os.Getenv("DROPCAM_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = os.TempDir()
		log.Printf("Use default dir: %v", uploadDir)
	}

	username = os.Getenv("USER")
	password = os.Getenv("HOME_PASSWORD")
	if password == "" {
		panic("Must provide a HOME_PASSWORD for authorization.")
	}
}

func listPhotos(p string) []string {
	log.Printf("listing file %v", p)
	matches, err := filepath.Glob(path.Join(p, "*.jpg"))
	if err != nil {
		log.Printf("%v", err)
	}
	log.Printf("matches %v", matches)
	return matches
}

func saveUpload(data []byte, name string) {
	dir := getUploadDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0700)
	}
	fullPath := path.Join(getUploadDir(), name)
	err := ioutil.WriteFile(fullPath, data, 0644)
	if err != nil {
		log.Printf("Error while saving [%s] upload [%s]", fullPath, err.Error())
	}
}

// sort the uploaded files by the date
func getUploadDir() string {
	return path.Join(uploadDir, time.Now().Format("2006-01-02"))
}

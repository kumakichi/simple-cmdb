package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func postLogin(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("username") != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	username := c.PostForm("username")
	password := c.PostForm("password")

	if strings.Trim(username, " ") == "" || strings.Trim(password, " ") == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Parameters can't be empty"})
		return
	}

	sql := fmt.Sprintf("select * from user where `username` = \"%s\" and `password` = \"%s\"", username, password)
	result := sqlQuery(sql)
	if len(result) == 1 {
		session.Set("username", username) //In real world usage you'd set this to the users ID
		err := session.Save()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session token"})
		} else {
			c.JSON(http.StatusOK, gin.H{"result": 0})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"result": 1})
	}
}

func loadTemplatesFiles() []string {
	var tmpls []string
	err := filepath.Walk("./templates/page", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			if err != nil {
				log.Println(err)
			}
			tmpls = append(tmpls, path)
		}
		return err
	})

	checkError(err)

	return tmpls
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("username")
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete token"})
	} else {
		c.Redirect(http.StatusFound, "/login")
	}
}

func loginPage(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("username") != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func pageTemplate(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username != nil {
		name := c.Param("tmpl")

		bb, err := ioutil.ReadFile(pageConfig)
		checkError(err)
		var dat gin.H
		err = json.Unmarshal(bb, &dat)
		checkError(err)
		dat["username"] = username.(string)

		c.HTML(http.StatusOK, name+".html", dat)
	} else {
		c.Redirect(http.StatusFound, "/login")
	}
}

func add(c *gin.Context) {
	c.Request.ParseForm()
	f := c.Request.Form
	table := f["action_type"][0]
	delete(f, "action_type")

	var keys, vals []string
	for k, v := range f {
		keys = append(keys, k)
		vals = append(vals, v[0])
	}

	sql := fmt.Sprintf("insert into `%s` (`%s`) values (\"%s\")", table, strings.Join(keys, "`, `"), strings.Join(vals, "\", \""))
	dbExec(sql)
	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func del(c *gin.Context) {
	dbExec(fmt.Sprintf("delete from `%s` where `id`=%s", c.PostForm("action_type"), c.PostForm("id")))
	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func list(c *gin.Context) {
	listResult := sqlQuery("select * from " + c.Query("action_type"))
	if len(listResult) == 0 {
		hack := make([]string, 0)
		c.JSON(http.StatusOK, gin.H{"result": hack})
	} else {
		c.JSON(http.StatusOK, gin.H{"result": listResult}) // gin will return empty array to null, see: https://github.com/gin-gonic/gin/issues/125
	}

}

func update(c *gin.Context) {
	c.Request.ParseForm()
	f := c.Request.Form
	table := f["action_type"][0]
	table_id := f["id"][0]
	delete(f, "action_type")
	delete(f, "id")

	var arr []string
	for k, v := range f {
		arr = append(arr, fmt.Sprintf("`%s` = \"%s\"", k, v[0]))
	}

	sql := fmt.Sprintf("update `%s` set %s where `id` = %s", table, strings.Join(arr, ", "), table_id)
	dbExec(sql)
	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func home(c *gin.Context) {
	c.Redirect(http.StatusFound, "/page/user")
}

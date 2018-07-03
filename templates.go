package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/bitly/go-simplejson"
)

type pageStruct struct {
	Name  string      `json:"name"`
	Title string      `json:"title"`
	Data  interface{} `json:"data"`
}

var (
	pageConfigs []interface{}
)

const (
	pageConfig = "page_config.json"
	head       = `{{ template "layout.html" .}}

{{ define "js" }}
    <script>	
        $(function() {
            $.rebootOps(
`

	foot = `            )
        })
    </script>
{{end}}
`
)

func parsePageCfg() []interface{} {
	pageContent, err := ioutil.ReadFile(pageConfig)
	checkError(err)
	c, err := simplejson.NewJson(pageContent)
	checkError(err)

	arr, err := c.Get("menu").Array()
	checkError(err)

	for _, v := range arr {
		vv := toMap(v)
		if sub, ok := vv["sub"]; ok {
			if s, ok := sub.([]interface{}); ok {
				pageConfigs = append(pageConfigs, s...)
			}
		} else {
			pageConfigs = append(pageConfigs, v)
		}
	}

	return pageConfigs
}

func genTemplates() {
	wd, err := os.Getwd()
	checkError(err)
	pagePath := wd + "/" + "templates/page/"

	for _, v := range pageConfigs {
		byt, err := json.Marshal(v)
		checkError(err)
		pc := pageStruct{}
		json.Unmarshal(byt, &pc)

		filePath := pagePath + pc.Name + ".html"
		if forceTemplate {
			os.Remove(filePath)
		}

		file, err := os.Create(filePath)
		checkError(err)
		defer file.Close()

		_, err = file.WriteString(head)
		checkError(err)
		_, err = file.WriteString("              " + string(byt))
		checkError(err)
		_, err = file.WriteString(foot)
		checkError(err)
		logP("[created template]:'" + pc.Name + "'")
	}
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const rootPath = "openapi/"
const indexPath = "public/"
const ymlFile = "openapi.yml"
const jsonFile = "openapi.json"
const indexFile = "index.html"

func main() {
	clearDir(getWdFile(""))
	initFile()
	writeFile(fmt.Sprintf("%sbasic.yml", rootPath), getWdFile(ymlFile))
	walkMatch(rootPath, "*.yml", getWdFile(ymlFile))

	yamlToJson(getWdFile(ymlFile), getWdFile(jsonFile))
	makeHTML(getWdFile(indexFile), getWdFile(jsonFile))
}

func clearDir(dir string) error {
	names, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entery := range names {
		os.RemoveAll(path.Join([]string{dir, entery.Name()}...))
	}
	return nil
}

func walkMatch(root, pattern, ymlFile string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path != fmt.Sprintf("%sbasic.yml", root) {
			if info.IsDir() && info.Name() != root {
				if str := strings.Replace(info.Name(), strings.TrimRight(root, "/"), "", 1); str != "" {
					s := fmt.Sprintf("%s:\n", getSpace(path)+str)
					writeString(s, ymlFile)
				}
			} else {
				if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
					return err
				} else if matched {
					writeFile(path, ymlFile)
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func initFile() {
	_, err := os.Stat(indexPath)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(indexPath, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}
	file, err := os.OpenFile(getWdFile(ymlFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
}

// 获取文件的绝对路径
func getWdFile(filename string) string {
	wd, _ := os.Getwd()
	return fmt.Sprintf("%s/%s%s", wd, indexPath, filename)
}

func readFile(file string) []byte {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func writeFile(inputFile, outputFile string) error {
	lines, err := readLines(inputFile)
	if err != nil {
		log.Println(err)
	}

	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		w.WriteString(getSpace(inputFile) + line)
		w.WriteByte('\n')
	}

	w.WriteByte('\n')
	return w.Flush()
}

func getSpace(path string) string {
	space := ""
	count := strings.Count(path, "/")
	for i := 1; i < count; i++ {
		space += "  "
	}
	return space
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeString(str, outputFile string) {
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	file.WriteString(str)
}

func yamlToJson(ymlFile, jsonFile string) {
	b := readFile(ymlFile)
	var body interface{}
	if err := yaml.Unmarshal(b, &body); err != nil {
		log.Fatal(err)
	}

	body = convert(body)

	if b, err := json.Marshal(body); err != nil {
		log.Fatal(err)
	} else {
		ioutil.WriteFile(jsonFile, b, 0644)
	}
}

func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}

func makeHTML(filename, jsonFile string) {
	template := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link href="https://fonts.googleapis.com/css?family=Open+Sans:400,700|Source+Code+Pro:300,600|Titillium+Web:400,600,700" rel="stylesheet">
  <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.24.2/swagger-ui.css" >
  <style>
    html
    {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
      overflow-y: scroll;
    }
    *,
    *:before,
    *:after
    {
      box-sizing: inherit;
    }
    body {
      margin:0;
      background: #fafafa;
    }
  </style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.24.2/swagger-ui-bundle.js"> </script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.24.2/swagger-ui-standalone-preset.js"> </script>
<script>
window.onload = function() {
  var spec = %s
  // Build a system
  const ui = SwaggerUIBundle({
    spec: spec,
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  })
  window.ui = ui
}
</script>
</body>
</html>
`
	json := string(readFile(jsonFile))
	writeString(fmt.Sprintf(template, json), filename)
}

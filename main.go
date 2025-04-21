/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

/*
import "venomouswolf/minlog/cmd"

func main() {
	cmd.Execute()

}*/

import (
	"fmt"
	"html/template"
	"os"
)

func main() {
	data := map[string][]int{
		"group1": {1, 2, 3},
		"group2": {4, 5, 6},
	}
	tmplStr := `{{- range $groupName, $nums := .}}Group: {{$groupName}}{{"\n"}}{{- range $nums}}Number: {{.}}{{"\n"}}{{- end}}{{- end}}`
	tmpl, err := template.New("example").Parse(tmplStr)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		fmt.Println("Error executing template:", err)
	}

}

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	nexus "github.com/tinyzimmer/nexus3-go"
)

func executeScript() {
	var scriptContent string
	if *scriptFile != "" {
		file, err := os.Open(*scriptFile)
		checkErr(err)
		scriptBytes, err := ioutil.ReadAll(file)
		checkErr(err)
		scriptContent = string(scriptBytes)
	} else if len(*scriptArgs) > 0 {
		log.Println(*scriptArgs)
		scriptContent = strings.Join(*scriptArgs, " ")
	} else {
		checkErr(errors.New("You must provide either a script file or a command to execute"))
	}
	client, err := nexus.New(*host, *username, *password)
	checkErr(err)
	script := client.NewEphemeralScript(&nexus.Script{
		Type:    nexus.ScriptTypeGroovy,
		Content: nexus.String(scriptContent),
	})
	res, err := script.Execute(nil)
	checkErr(err)
	fmt.Println(*res.Result)
}

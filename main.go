package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

func main() {
	// disable showing time with log failures
	log.SetFlags(0)

	scripts := loadConfig()

	script, shoulRun := selectScript(scripts)

	if shoulRun {
		runNpmScript(script.Name)
	}
}

func runNpmScript(scriptName string) {
	fmt.Printf("> npm run %v\n", scriptName)
	cmd := exec.Command("npm", "run", scriptName)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

type NpmScript struct {
	Name    string
	Command string
}

type PackageJson struct {
	Scripts map[string]string `json:"scripts"`
}

func loadConfig() []NpmScript {
	scripts := make([]NpmScript, 0)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting working directory path\n", err)
	}

	configFilePath := path.Join(cwd, "package.json")
	packageJsonBytes, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		log.Fatal("Error reading package.json file\n", err)
	}

	packageJson := PackageJson{
		Scripts: make(map[string]string),
	}
	// load config
	err = json.Unmarshal(packageJsonBytes, &packageJson)

	if err != nil {
		log.Fatal("Error parsing package.json file\n", err)
	}

	for name, command := range packageJson.Scripts {
		scripts = append(scripts, NpmScript{
			Name:    name,
			Command: command,
		})
	}

	return scripts
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
)

func main() {
	// disable showing time with log failures
	log.SetFlags(0)

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal("Unable to find working directory: ", err)
	}

	scripts := loadConfig()

	// last commands are stored per-path, so you can just hit enter immediately to rerun the last command
	cache, err := loadCache()

	if err == nil && cache[cwd] != "" {
		// find and prepend the last script used for this path
		lastRunScript := cache[cwd]

		for _, script := range scripts {
			if script.Name == lastRunScript {
				suggestionScript := NpmScript{
					Name:        script.Name,
					Command:     script.Command,
					Description: "last run",
				}
				// prepend the last script that was run, so they can just hit
				// enter immediately to run the last thing they ran again
				// without having to find it in the list
				scripts = append([]NpmScript{suggestionScript}, scripts...)
			}
		}
	}

	script, shouldRun := selectScript(scripts)

	if shouldRun {
		cache[cwd] = script.Name
		saveCache(cache)

		runNpmScript(script.Name)
	}
}

func saveCache(cache map[string]string) {
	cacheBytes, err := json.Marshal(cache)

	if err != nil {
		log.Fatal("Failed to marshal last run cache: ", err)
	}

	err = os.MkdirAll(getCachePath(), 0700)

	if err != nil {
		log.Fatal("Failed to create cache path: ", err)
	}

	err = os.WriteFile(getCacheFilePath(), cacheBytes, 0600)

	if err != nil {
		log.Fatal("Failed to write cache: ", err)
	}
}

func getCachePath() string {
	cacheDir, err := os.UserCacheDir()

	if err != nil {
		log.Fatalf("Failed to get cache path")
	}

	return path.Join(cacheDir, "sheodox", "nr")
}

func getCacheFilePath() string {
	return path.Join(getCachePath(), "last_commands.json")
}

func loadCache() (map[string]string, error) {
	// cached commands is a map of the cwd path and the command name
	cached := make(map[string]string)

	cachePath := getCacheFilePath()

	cacheBytes, err := os.ReadFile(cachePath)

	if err != nil {
		return cached, err
	}

	err = json.Unmarshal(cacheBytes, &cached)

	return cached, err
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
	Name        string
	Command     string
	Description string
}

type PackageJson struct {
	Scripts map[string]string `json:"scripts"`
}

func loadConfig() []NpmScript {
	scripts := make([]NpmScript, 0)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed getting working directory path: ", err)
	}

	configFilePath := path.Join(cwd, "package.json")
	packageJsonBytes, err := os.ReadFile(configFilePath)

	if err != nil {
		log.Fatal("Failed to read package.json file: ", err)
	}

	packageJson := PackageJson{
		Scripts: make(map[string]string),
	}

	err = json.Unmarshal(packageJsonBytes, &packageJson)

	if err != nil {
		log.Fatal("Failed to parsing package.json file: ", err)
	}

	for name, command := range packageJson.Scripts {
		scripts = append(scripts, NpmScript{
			Name:    name,
			Command: command,
		})
	}

	// sort scripts alphabetically by name
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].Name < scripts[j].Name
	})

	return scripts
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	latestUrl    string = "https://api.github.com/repos/brave/brave-browser/releases/latest"
	templateFile string = "template"
	baseUrl      string = "https://github.com/brave/brave-browser/releases/download/v"
)

func getSha(version string) (string, error) {
	shaUrl := fmt.Sprintf("%s%s/brave-browser_%s_amd64.deb.sha256", baseUrl, version, version)
	data, err := getContent(shaUrl)
	if err != nil {
		return "", err
	}

	return strings.Split(string(data), " ")[0], nil
}

// Get the version of the latest stable release
func getVersion() (string, error) {
	data, err := getContent(latestUrl)
	if err != nil {
		return "", err
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		return "", err
	}

	return jsonMap["tag_name"].(string)[1:], nil
}

func getContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %v", err)
	}

	return data, nil
}

func main() {
	version, err := getVersion()
	if err != nil {
		log.Printf("Failed to get version: %v", err)
		os.Exit(1)
	}

	sha256, err := getSha(version)
	if err != nil {
		log.Printf("Failed to get SHA256: %v", err)
		os.Exit(1)
	}

	// Replace version and checksum lines
	input, err := os.ReadFile(templateFile)
	if err != nil {
		log.Printf("Failed to read template file: %v", err)
		os.Exit(1)
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, "version=") {
			lines[i] = fmt.Sprintf("version=%s", version)
		} else if strings.Contains(line, "checksum=") {
			lines[i] = fmt.Sprintf("checksum=%s", sha256)
		}
	}

	// Write new template to file
	output := strings.Join(lines, "\n")
	err = os.WriteFile(templateFile, []byte(output), 0644)
	if err != nil {
		log.Printf("Failed to write template file: %v", err)
		os.Exit(1)
	}

	fmt.Println("Successfully updated template file")
}

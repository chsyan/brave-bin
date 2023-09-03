package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
    "flag"
    "os/exec"
)

const (
	latestUrl    string = "https://api.github.com/repos/brave/brave-browser/releases/latest"
	templateFile string = "template"
	baseUrl      string = "https://github.com/brave/brave-browser/releases/download/v"
    voidTemplateFile string = "srcpkgs/brave-bin/template"
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

func updateTemplate() {
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

func updatePackage(voidDir string) {
	input, err := os.ReadFile(templateFile)
	if err != nil {
		log.Printf("Failed to read template file: %v", err)
		os.Exit(1)
	}
	lines := strings.Split(string(input), "\n")

    err = os.Chdir(voidDir)
    if err != nil {
        log.Printf("Failed to change directory: %v", err)
        os.Exit(1)
    }

	// Write template to file
	output := strings.Join(lines, "\n")
	err = os.WriteFile(voidTemplateFile, []byte(output), 0644)
	if err != nil {
		log.Printf("Failed to write template file: %v", err)
		os.Exit(1)
	}

	fmt.Println("Successfully copied template file")
	fmt.Println("Building brave-bin")

    cmd := exec.Command("./xbps-src", "pkg", "brave-bin")
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    out, err := cmd.Output()
    if err != nil {
        fmt.Println(string(out))
        os.Exit(1)
    } else {
	    fmt.Println("Finished building brave-bin")
    }

    fmt.Println("Installing brave-bin")
    cmd = exec.Command("xi", "brave-bin")
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    out, err = cmd.Output()
}

func main() {
    templatePtr := flag.Bool("t", false, "update template")
    updatePtr := flag.Bool("u", false, "update void-packages source")
    home, _ := os.UserHomeDir()
    defaultPath := fmt.Sprintf("%s/src/void-packages", home)
    pathPtr := flag.String("p", defaultPath, "path to void-packages directory")

    flag.Parse()

    if !*templatePtr && !*updatePtr {
        fmt.Println("Please specify one of -t or -u")
        os.Exit(1)
    }

    if *templatePtr {
        updateTemplate()
    }

    if *updatePtr {
        updatePackage(*pathPtr)
    }
}

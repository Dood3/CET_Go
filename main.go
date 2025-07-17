// Run:
// go run main.go
//
// Compile:
// GOOS=windows GOARCH=amd64 go build -o myprogram.exe main.go
// GOOS=linux GOARCH=amd64 go build -o myprogram main.go
// GOOS=darwin GOARCH=arm64 go build -o myprogram main.go  # macOS ARM
//
// Compile binary smaller and more optimized:
// go build -ldflags="-s -w" -o myprogram main.go

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("=== GoSt - Go Stager ===\n")

	scanner := bufio.NewScanner(os.Stdin)

	// Step 0: Create output directory
	outputDir := "output"
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Step 1: Get user input
	targetOS := getTargetOS(scanner)
	fmt.Printf("Selected target: %s\n\n", targetOS)

	command := getCommand(scanner)
	fmt.Printf("Command to embed: %s\n\n", command)

	outputName := getOutputName(scanner, targetOS)
	fmt.Printf("Output binary: %s\n\n", outputName)

	// Step 2: Generate payload source code
	// Generate source filename based on output name
	baseName := strings.TrimSuffix(outputName, ".exe")
	payloadSourceFile := filepath.Join(outputDir, baseName+".go")
	outputBinaryPath := filepath.Join(outputDir, outputName)

	err = generatePayloadSource(command, targetOS, payloadSourceFile)
	if err != nil {
		fmt.Printf("Error generating payload source: %v\n", err)
		return
	}
	fmt.Printf("Payload source code generated as %s\n", payloadSourceFile)

	// Step 3: Compile the payload
	err = compilePayload(payloadSourceFile, outputBinaryPath, targetOS)
	if err != nil {
		fmt.Printf("Error compiling payload: %v\n", err)
		return
	}
	fmt.Printf("Compilation successful! Files generated:\n- %s\n- %s\n", payloadSourceFile, outputBinaryPath)
}

// Prompt for target OS
func getTargetOS(scanner *bufio.Scanner) string {
	for {
		fmt.Print("Choose target OS - [W]indows / [L]inux: ")
		scanner.Scan()
		choice := strings.ToLower(strings.TrimSpace(scanner.Text()))

		switch choice {
		case "w", "windows":
			return "windows"
		case "l", "linux":
			return "linux"
		default:
			fmt.Println("Invalid choice. Please enter 'W' for Windows or 'L' for Linux.")
		}
	}
}

// Prompt for command to embed
func getCommand(scanner *bufio.Scanner) string {
	for {
		fmt.Print("Enter command to embed: ")
		scanner.Scan()
		command := strings.TrimSpace(scanner.Text())

		if command == "" {
			fmt.Println("Command cannot be empty. Please try again.")
			continue
		}

		return command
	}
}

// Prompt for output filename
func getOutputName(scanner *bufio.Scanner, targetOS string) string {
	extension := ""
	if targetOS == "windows" {
		extension = ".exe"
	}

	fmt.Printf("Enter output filename (default: payload%s): ", extension)
	scanner.Scan()
	name := strings.TrimSpace(scanner.Text())

	if name == "" {
		return "payload" + extension
	}

	// Add extension if not present and targeting Windows
	if targetOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
		name += ".exe"
	}

	return name
}

// Generate the payload Go source file with the embedded command
func generatePayloadSource(command, targetOS, filename string) error {
	var template string
	if targetOS == "windows" {
		template = `package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := %q
	c := exec.Command("cmd", "/C", cmd)
	out, err := c.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command: %%v\n", err)
	}
	fmt.Print(string(out))
}
`
	} else {
		template = `package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := %q
	c := exec.Command("sh", "-c", cmd)
	out, err := c.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command: %%v\n", err)
	}
	fmt.Print(string(out))
}
`
	}

	source := fmt.Sprintf(template, command)
	return os.WriteFile(filename, []byte(source), 0644)
}

// Compile the payload Go file for the selected OS
func compilePayload(sourceFile, outputFile, targetOS string) error {
	var goos string
	var goarch = "amd64" // 64-bit by default

	if targetOS == "windows" {
		goos = "windows"
	} else {
		goos = "linux"
	}

	cmd := exec.Command("go", "build", "-o", outputFile, sourceFile)
	cmd.Env = append(os.Environ(),
		"GOOS="+goos,
		"GOARCH="+goarch,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

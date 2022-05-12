package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"time"
)

var (
	useHelp bool

	actor  string
	repo   string
	token  string
	dbName string
)

const (
	apiUrl = "https://api.github.com"
)

type PutRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
}

func main() {

	date := time.Now()
	formattedDate := date.Format(time.RFC3339)

	info, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	if info.Username != "postgres" {
		log.Fatal("This command needs to be launched by 'postgres' user.")
	}

	flag.BoolVar(&useHelp, "help", false, "Show this help menu.")
	flag.StringVar(&actor, "actor", "", "The GitHub repository actor.")
	flag.StringVar(&repo, "repo", "", "The GitHub repository name.")
	flag.StringVar(&token, "token", "", "The GitHub Personal Access Token.")
	flag.StringVar(&dbName, "db", "", "The name of the database to backup.")

	flag.Parse()

	if useHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if len(actor) < 1 {
		log.Fatal("Actor cannot be empty, use --help.")
	}

	if len(repo) < 1 {
		log.Fatal("Repo cannot be empty, use --help.")
	}

	if len(token) < 1 {
		log.Fatal("Token cannot be empty, use --help.")
	}

	if len(dbName) < 1 {
		log.Fatal("Database Name cannot be empty, use --help.")
	}

	log.Printf("Creating backup from database '%s'...\n", dbName)

	tempDir := os.TempDir()
	fileName := fmt.Sprintf("dump-%s-%s.backup", dbName, formattedDate)
	destination := fmt.Sprintf("%s/%s", tempDir, fileName)

	cmd := exec.Command("pg_dump", "-Z5", "-Fc", dbName, "-f", destination)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Backup created successfully on %s.\n", destination)
	log.Printf("Encoding file to Base64...\n")

	content, err := os.ReadFile(destination)
	if err != nil {
		log.Fatal(err)
	}

	encodedFile := base64.StdEncoding.EncodeToString(content)

	log.Printf("Uploading encoded backup to GitHub Repository %s/%s...\n", actor, repo)

	body := PutRequest{
		Message: fmt.Sprintf("Upload backup from %s at %s", dbName, formattedDate),
		Content: encodedFile,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/repos/%s/%s/contents/%s", apiUrl, actor, repo, fileName), &buf)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		log.Fatal(resp.Status)
	}

	log.Printf("Successfully uploaded encoded backup to GitHub.\n\n")

}

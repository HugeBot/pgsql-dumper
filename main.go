package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"gopkg.in/yaml.v2"
)

var (
	config      *Config
	filePath    string
	date        time.Time
	baseCommand string
	containerId string
)

type Config struct {
	S3 struct {
		Endpoint        string `yaml:"endpoint"`
		Bucket          string `yaml:"bucket"`
		AccessKeyId     string `yaml:"accessKeyId"`
		AccessKeySecret string `yaml:"accessKeySecret"`
		Region          string `yaml:"region"`
	} `yaml:"s3"`
	Database struct {
		Name     string `yaml:"name"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"database"`
}

type PutRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
}

func init() {
	date = time.Now()

	var useHelp bool
	flag.BoolVar(&useHelp, "help", false, "Show this help menu.")

	flag.StringVar(&filePath, "config", "./config.yml", "Select where is located config file.")

	flag.StringVar(&containerId, "container", "", "Specific the ID (or name) of the container in which the instance of the database is running, this will avoid the requirement that the command is executed by the postgre user.")

	flag.Parse()

	if useHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	file, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatal(err)
	}

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	if config.Database.Name == "" {
		config.Database.Name = "all"
		baseCommand = "pg_dumpall"
	} else {
		baseCommand = "pg_dump"
	}

	if config.S3.Region == "" {
		log.Fatalf("s3 region not defined on %s", file)
	}

	if config.S3.Bucket == "" {
		log.Fatalf("s3 bucket not defined on %s", file)
	}

	if config.S3.AccessKeyId == "" {
		log.Fatalf("s3 accessKeyId not defined on %s", file)
	}

	if config.S3.AccessKeySecret == "" {
		log.Fatalf("s3 secretAccessKey not defined on %s", file)
	}
}

func printBanner() {
	fmt.Println(`
    ┌───────────────────────────────────────────────────┐
    │                    PSQL DUMPER                    │
    │                                                   │
    │       https://github.com/HugeBot/psql-dumper      │
    └───────────────────────────────────────────────────┘
	`)
}

func prepareS3Connection() *s3manager.Uploader {
	awsConfig := aws.NewConfig()

	awsConfig.WithRegion(config.S3.Region)
	awsConfig.WithEndpoint(config.S3.Endpoint)
	awsConfig.WithCredentials(credentials.NewStaticCredentials(config.S3.AccessKeyId, config.S3.AccessKeySecret, ""))

	sess := session.Must(session.NewSession(awsConfig))

	log.Println("successfully conected with S3 bucket")

	return s3manager.NewUploader(sess)
}

func buildCommand(destination string) *exec.Cmd {
	if containerId == "" {
		return exec.Command(baseCommand, "-Z5", "-Fc", config.Database.Name, "-f", destination)
	} else {
		cmdArray := []string{
			"docker",
			"exec",
			"-i",
			containerId,
			"/bin/bash",
			"-c",
		}

		if config.Database.Password != "" {
			cmdArray = append(cmdArray, fmt.Sprintf("\"PGPASSWORD=%s", config.Database.Password), baseCommand)
		} else {
			cmdArray = append(cmdArray, fmt.Sprintf("\"%s", baseCommand))
		}

		if config.Database.Username != "" {
			cmdArray = append(cmdArray, "--username", config.Database.Username)
		} else {
			cmdArray = append(cmdArray, "--username", "postgres")
		}

		if config.Database.Name != "all" {
			cmdArray = append(cmdArray, fmt.Sprintf("%s\"", config.Database.Name), ">", destination)
		} else {
			cmdArray = append(cmdArray, "\"", ">", destination)
		}

		return exec.Command(cmdArray[0], cmdArray[1:]...)
	}
}

func main() {
	printBanner()

	formattedDate := date.Format(time.RFC3339)

	if containerId == "" {
		info, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}

		if info.Username != "postgres" {
			log.Fatal("this command needs to be launched by 'postgres' user or with '--container <container name or id>' flag.")
		}
	}

	log.Printf("creating backup from database '%s'...\n", config.Database.Name)

	tempDir := os.TempDir()
	fileName := fmt.Sprintf("dump-%s-%s.backup", config.Database.Name, formattedDate)
	destination := fmt.Sprintf("%s/%s", tempDir, fileName)

	cmd := buildCommand(destination)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Running command %v\n", cmd.Args)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Printf("backup created successfully on %s.\n", destination)

	content, err := os.ReadFile(destination)
	if err != nil {
		log.Fatal(err)
	}

	result, err := prepareS3Connection().Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.S3.Bucket),
		Key:    aws.String(fileName),
		Body:   aws.ReadSeekCloser(bytes.NewReader(content)),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("file uploaded to, %s\n", aws.StringValue(&result.Location))
}

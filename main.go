package main

import (
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
	config       *Config
	filePath     string
	date         time.Time
	baseCommand  string
	containerId  string
	containerCLI string
	allDatabases bool

	compressLevel int

	Version = "unknown"
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
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
	} `yaml:"database"`
}

type PutRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
}

func (c *Config) init() {
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
		if !allDatabases {
			log.Fatalf("database name not defined on %s", file)
		}
		config.Database.Name = "all"
		baseCommand = "pg_dumpall"
	} else {
		baseCommand = "pg_dump"
	}

	if config.Database.Host == "" {
		config.Database.Host = "127.0.0.1"
	}

	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}

	if config.Database.Username == "" {
		log.Fatalf("database username not defined on %s", file)
	}

	if config.Database.Password == "" {
		log.Fatalf("database password not defined on %s", file)
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

func init() {
	date = time.Now()

	var useHelp bool
	flag.BoolVar(&useHelp, "help", false, "Show this help menu.")
	flag.BoolVar(&allDatabases, "all", false, "If defined will be dumped all the databases (pg_dumpall instead of pg_dump)")

	flag.StringVar(&filePath, "config", "./config.yml", "Select where is located config file.")

	flag.StringVar(&containerCLI, "cli", "docker", "Determine runtime command like docker (default), nerdctl, podman... must be a docker compatible CLI.")
	flag.StringVar(&containerId, "cid", "", "Specific the ID (or name) of the container in which the instance of the database is running, this will avoid the requirement that the command is executed by the postgre user.")
	flag.IntVar(&compressLevel, "compress", 5, "The compress level (default to 5)")

	flag.Parse()

	if useHelp {
		printBanner()
		flag.PrintDefaults()
		os.Exit(0)
	}

	if compressLevel < 0 && compressLevel > 9 {
		log.Fatalln("the compression level must be between 0 and 9 (both inclusive)")
	}

	config.init()
}

func printBanner() {
	fmt.Printf(`
    ┌───────────────────────────────────────────────────┐
    │                  PSQL DUMPER v%s               │
    │                                                   │
    │       https://github.com/HugeBot/psql-dumper      │
    └───────────────────────────────────────────────────┘
	
	`, Version)
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

// pg_dump -Z5 -Fc --dbname=postgresql://postgres:postgres@127.0.0.1:5432/hugebot
func buildCommand(destination string) *exec.Cmd {
	if containerId == "" {
		return exec.Command(baseCommand, "-Z5", "-Fc")
	} else if allDatabases {
		return exec.Command(containerCLI, "exec", containerId, baseCommand, fmt.Sprintf("--dbname=postgresql://%s:%s@%s:%d", config.Database.Username, config.Database.Password, config.Database.Host, config.Database.Port))
	} else {
		return exec.Command(containerCLI, "exec", containerId, baseCommand, fmt.Sprintf("-Z%d", compressLevel), "-Fc", fmt.Sprintf("--dbname=postgresql://%s:%s@%s:%d/%s", config.Database.Username, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name))
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
	cmd.Stderr = os.Stderr

	log.Printf("Running command %v\n", cmd.Args)
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	result, err := prepareS3Connection().Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.S3.Bucket),
		Key:    aws.String(fileName),
		Body:   aws.ReadSeekCloser(out),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("file uploaded to, %s\n", aws.StringValue(&result.Location))
}

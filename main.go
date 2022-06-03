package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gurkankaymak/hocon"
)

var (
	config   *Config
	filePath string
)

type DatabaseConfig struct {
	Name string
}

type S3Config struct {
	Endpoint        string
	Bucket          string
	SecretAccessKey string
	AccessKeyId     string
	Region          string
}

type Config struct {
	S3       S3Config
	Database DatabaseConfig
}

type PutRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
}

func init() {
	var useHelp bool
	flag.BoolVar(&useHelp, "help", false, "Show this help menu.")

	flag.StringVar(&filePath, "file", "./psql.conf", "Select where is located config file.")

	flag.Parse()

	if useHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	file, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatal(err)
	}

	conf, err := hocon.ParseResource(file)
	if err != nil {
		log.Fatal(err)
	}

	config = &Config{
		Database: DatabaseConfig{
			Name: conf.GetString("database.db_name"),
		},
		S3: S3Config{
			Endpoint:        conf.GetString("s3.endpoint"),
			Bucket:          conf.GetString("s3.bucket"),
			Region:          conf.GetString("s3.region"),
			AccessKeyId:     conf.GetString("s3.access_key_id"),
			SecretAccessKey: conf.GetString("s3.secret_access_key"),
		},
	}

	if config.Database.Name == "" {
		config.Database.Name = "postgres"
	}

	if config.S3.Region == "" {
		config.S3.Region = "auto"
	}

	if config.S3.Endpoint == "" {
		log.Fatalf("s3 endpoint not defined on %s", file)
	}

	if config.S3.Bucket == "" {
		log.Fatalf("s3 bucket not defined on %s", file)
	}

	if config.S3.AccessKeyId == "" {
		log.Fatalf("s3 accessKeyId not defined on %s", file)
	}

	if config.S3.SecretAccessKey == "" {
		log.Fatalf("s3 secretAccessKey not defined on %s", file)
	}

	log.Printf("%v#", config)
}

func prepareS3Connection() *s3manager.Uploader {
	awsConfig := aws.NewConfig()

	awsConfig.WithRegion(config.S3.Region)
	awsConfig.WithEndpoint(config.S3.Endpoint)
	awsConfig.WithCredentials(credentials.NewCredentials(&credentials.StaticProvider{
		Value: credentials.Value{
			AccessKeyID:     config.S3.AccessKeyId,
			SecretAccessKey: config.S3.SecretAccessKey,
		},
	}))

	sess := session.Must(session.NewSession(awsConfig))

	return s3manager.NewUploader(sess)
}

func main() {

	date := time.Now()
	formattedDate := date.Format(time.RFC3339)

	info, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	if info.Username != "postgres" {
		log.Fatal("this command needs to be launched by 'postgres' user.")
	}

	log.Printf("creating backup from database '%s'...\n", config.Database.Name)

	tempDir := os.TempDir()
	fileName := fmt.Sprintf("dump-%s-%s.backup", config.Database.Name, formattedDate)
	destination := fmt.Sprintf("%s/%s", tempDir, fileName)

	cmd := exec.Command("pg_dump", "-Z5", "-Fc", config.Database.Name, "-f", destination)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Printf("backup created successfully on %s.\n", destination)

	content, err := os.ReadFile(destination)
	if err != nil {
		log.Fatal(err)
	}

	result, err := prepareS3Connection().Upload(&s3manager.UploadInput{
		Bucket: &config.S3.Bucket,
		Key:    &fileName,
		Body:   strings.NewReader(string(content)),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("file uploaded to, %s\n", aws.StringValue(&result.Location))
}

package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var AccessKeyID string
var SecretAccessKey string
var MyRegion string
var MyBucket string
var filepath string

//GetEnvWithKey : get env value
func GetEnvWithKey(key string) string {
	return os.Getenv(key)
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		os.Exit(1)
	}
}

func ConnectAws() *session.Session {
	AccessKeyID = GetEnvWithKey("AWS_ACCESS_KEY_ID")
	SecretAccessKey = GetEnvWithKey("AWS_SECRET_ACCESS_KEY")
	MyRegion = GetEnvWithKey("AWS_REGION")

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(MyRegion),
			Credentials: credentials.NewStaticCredentials(
				AccessKeyID,
				SecretAccessKey,
				"", // a token will be created when the session it's used.
			),
		})

	if err != nil {
		panic(err)
	}

	return sess
}

func UploadFile(c *gin.Context) {
	sess := c.MustGet("sess").(*session.Session)
	uploader := s3manager.NewUploader(sess)

	MyBucket = GetEnvWithKey("BUCKET_NAME")

	file, header, err := c.Request.FormFile("file")
	filename := header.Filename
	up, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(MyBucket),
		ACL:    aws.String("public-read"),
		Key:    aws.String(filename),
		Body:   file,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Failed to upload file",
			"uploader": up,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"file successfully added ": filename ,
	})
}

func UploadFiles(c *gin.Context) {

	MyBucket = GetEnvWithKey("BUCKET_NAME")
	sess := c.MustGet("sess").(*session.Session)
	svc:= s3manager.NewUploader(sess)
	file,header, _ := c.Request.FormFile("file")
	filename := header.Filename
	files := []s3manager.BatchUploadObject{
		{
			Object:	&s3manager.UploadInput {
				Bucket: aws.String(MyBucket),
				Key: aws.String(filename),
				Body:   file,
			},
		},
	}

	iter := &s3manager.UploadObjectsIterator{Objects: files}
	if err := svc.UploadWithIterator(aws.BackgroundContext(), iter); err != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"file successfully added ": filename ,
	})
}

func DeleteFile(c *gin.Context) {
	svc := s3.New(session.New())
	MyBucket = GetEnvWithKey("BUCKET_NAME")
	Key := c.Request.FormValue("Key")
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(MyBucket),
		Key:    aws.String(Key),
	}
	result, err := svc.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
	c.JSON(http.StatusOK, gin.H{
		"delete successful ": Key ,
	})
	return

}

func DownloadFiles(c *gin.Context)  {
	sess := session.Must(session.NewSession())
	downloader := s3manager.NewDownloader(sess)
	MyBucket = GetEnvWithKey("BUCKET_NAME")
	Key := c.Request.FormValue("Key")
	f, err := os.Create(Key)
	n, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(MyBucket),
		Key:    aws.String(Key),
	})
	if err != nil {
		return
	}
	fmt.Println(n)
	c.JSON(http.StatusOK, gin.H{
		"download successful ": Key ,
	})
}

func main() {
	LoadEnv()

	sess := ConnectAws()
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("sess", sess)
		c.Next()
	})

	router.POST("/upload", UploadFile)
	router.POST("/uploads", UploadFiles)
	router.DELETE("/delete", DeleteFile)
	router.GET("/download",DownloadFiles)

	_ = router.Run(":3030")
}

package controller

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"main.go/config"
	"net/http"
	"os"
)

var MyBucket string

func UploadFile(c *gin.Context) {
	sess := c.MustGet("sess").(*session.Session)
	uploader := s3manager.NewUploader(sess)

	MyBucket = config.GetEnvWithKey("BUCKET_NAME")

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

	MyBucket = config.GetEnvWithKey("BUCKET_NAME")
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
	MyBucket = config.GetEnvWithKey("BUCKET_NAME")
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
	MyBucket = config.GetEnvWithKey("BUCKET_NAME")
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

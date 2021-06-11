package services

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/disintegration/imaging"
	"github.com/sentrionic/OlympusGin/config"
	"github.com/sentrionic/OlympusGin/utils"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"mime/multipart"
	"regexp"
	"strings"
)

const DimMax = 1080
const DimMin = 320

type FileService interface {
	UploadAvatar(image *multipart.FileHeader, directory string) (string, error)
	UploadImage(image *multipart.FileHeader, directory string) (string, error)
	DeleteImage(key string) error
}

type fileService struct {
	sess   *session.Session
	bucket string
}

func NewFileService(c *config.Config) FileService {

	cfg := c.Get()
	accessKey := cfg.GetString("aws.access_key")
	secretKey := cfg.GetString("aws.secret_access_key")
	region := cfg.GetString("aws.region")
	bucket := cfg.GetString("aws.storage_bucket_name")

	sess, err := session.NewSession(
		&aws.Config{
			Credentials: credentials.NewStaticCredentials(
				accessKey,
				secretKey,
				"",
			),
			Region: aws.String(region),
		},
	)

	if err != nil {
		panic("error initializing s3")
	}

	return &fileService{
		sess:   sess,
		bucket: bucket,
	}
}

func (fs *fileService) UploadAvatar(header *multipart.FileHeader, directory string) (string, error) {
	uploader := s3manager.NewUploader(fs.sess)

	key := fmt.Sprintf("files/%s/avatar.jpeg", directory)

	file, err := header.Open()

	if err != nil {
		return "", err
	}

	src, _, err := image.Decode(file)

	if err != nil {
		return "", err
	}

	img := imaging.Resize(src, 150, 0, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 75})

	if err != nil {
		return "", err
	}

	up, err := uploader.Upload(&s3manager.UploadInput{
		Body:        buf,
		Bucket:      aws.String(fs.bucket),
		ContentType: aws.String("image/jpeg"),
		Key:         aws.String(key),
	})

	if err != nil {
		return "", err
	}

	return up.Location, nil
}

func (fs *fileService) UploadImage(header *multipart.FileHeader, directory string) (string, error) {
	uploader := s3manager.NewUploader(fs.sess)

	key := fmt.Sprintf("files/%s/%s", directory, formatName(header.Filename))

	file, err := header.Open()

	if err != nil {
		return "", err
	}

	src, _, err := image.Decode(file)

	if err != nil {
		return "", err
	}

	b := src.Bounds()
	width := b.Dx()
	height := b.Dy()

	var img *image.NRGBA
	if height < DimMin || width < DimMin {
		img = imaging.Resize(src, DimMin, 0, imaging.Lanczos)
	} else if height > DimMax && height > width {
		img = imaging.Fit(src, width, DimMax, imaging.Lanczos)
	} else if width > DimMax && width > height {
		img = imaging.Fit(src, DimMax, height, imaging.Lanczos)
	} else {
		img = imaging.Fill(src, DimMax, DimMax, imaging.Center, imaging.Lanczos)
	}

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 75})

	if err != nil {
		return "", err
	}

	up, err := uploader.Upload(&s3manager.UploadInput{
		Body:        buf,
		Bucket:      aws.String(fs.bucket),
		ContentType: aws.String("image/jpeg"),
		Key:         aws.String(key),
	})

	if err != nil {
		return "", err
	}

	return up.Location, nil
}

var re = regexp.MustCompile(`/[^a-z0-9]/g`)

func formatName(filename string) string {
	pre := utils.RandomString(5)
	index := strings.LastIndex(filename, ".")
	filename = filename[:index]
	filename = strings.ToLower(filename)
	filename = re.ReplaceAllString(filename, "-")
	return fmt.Sprintf("%s-%s.jpeg", pre, filename)
}

func (fs *fileService) DeleteImage(key string) error {
	srv := s3.New(fs.sess)
	_, err := srv.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(fs.bucket),
		Key:    aws.String(key),
	})

	return err
}

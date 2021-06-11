# OlympusGin

OlympusGin is a backend for the [OlympusBlog](https://github.com/sentrionic/OlympusBlog) stack using [Gin-Gonic](https://github.com/gin-gonic/gin).

## Stack
- [Gorm](https://gorm.io/) as the DB ORM
- [Validator](https://github.com/go-playground/validator) for validation
- [Imaging](https://github.com/disintegration/imaging) for image resizing

## Getting started

0. Install Golang
1. Clone this repository
2. Run `go mod tidy` to get all the dependencies.
3. Rename `local.example.yaml` in `config` to `local.yaml`
   and fill out the values. AWS is only required if you want file upload,
   GMail if you want to send reset emails.
4. Run `go build github.com/sentrionic/OlympusGin`
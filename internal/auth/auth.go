package auth

import "os"

type Credentials struct {
	Username string
	Password string
}

func GetCredentials() Credentials {
	return Credentials{
		Username: os.Getenv("ON_DOWNLOADER_USERNAME"),
		Password: os.Getenv("ON_DOWNLOADER_PASSWORD"),
	}
}

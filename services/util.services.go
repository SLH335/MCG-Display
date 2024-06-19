package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
)

func RenderComponent(component templ.Component) string {
	buf := new(bytes.Buffer)
	component.Render(context.Background(), buf)

	return buf.String()
}

func GetCredentials(n int) (username, password string, err error) {
	err = godotenv.Load()
	if err != nil {
		return "", "", err
	}

	suffix := fmt.Sprintf("_%d", n)

	username = os.Getenv("WEBUNTIS_USERNAME" + suffix)
	password = os.Getenv("WEBUNTIS_PASSWORD" + suffix)

	if username == "" && password == "" {
		return "", "", errors.New("error: no credentials found in .env")
	}

	return username, password, nil
}

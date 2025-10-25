package tools

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/auth/repo"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"golang.org/x/term"
)

var (
	ErrEmailMissing     = errors.New("email missing")
	ErrEmailNotUnique   = errors.New("user with specified email already exists")
	ErrPasswordMismatch = errors.New("passwords do not match")
)

func CreateUser(ctx context.Context, store *store.Store) error {
	authRepo := repo.New(store)

	email, err := getValidEmail(ctx, authRepo)
	if err != nil {
		return err
	}

	password, err := getValidPassword()
	if err != nil {
		return err
	}

	user, err := createUser(ctx, authRepo, email, password)
	if err != nil {
		return err
	}

	logUserCreated(user)
	return nil
}

func getValidEmail(ctx context.Context, authRepo repo.Repo) (string, error) {
	email, err := getEmailString()
	if err != nil {
		return "", err
	}
	if err := assertEmailUniqueness(ctx, authRepo, email); err != nil {
		return "", err
	}

	return email, nil
}

func getEmailString() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter email: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	email := strings.TrimSpace(input)
	if email == "" {
		return "", ErrEmailMissing
	}

	return email, nil
}

func assertEmailUniqueness(ctx context.Context, authRepo repo.Repo, email string) error {
	_, err := authRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return ErrEmailNotUnique
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	return err
}

func getValidPassword() (string, error) {
	password1, err := getPasswordString("Enter password: ")
	if err != nil {
		return "", err
	}

	password2, err := getPasswordString("Enter password again: ")
	if err != nil {
		return "", err
	}

	if password1 != password2 {
		return "", ErrPasswordMismatch
	}

	return password1, nil
}

func getPasswordString(msg string) (string, error) {
	fmt.Print(msg)

	bytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	input := string(bytes)
	password := strings.TrimSpace(input)

	return password, nil
}

func createUser(ctx context.Context, authRepo repo.Repo, email string, password string) (*domain.UserWithHash, error) {
	hash, err := utils.GetHash(password)
	if err != nil {
		return nil, err
	}

	return authRepo.CreateUser(ctx, email, hash)
}

func logUserCreated(user *domain.UserWithHash) {
	log.Print("User created")
	log.Print("User ID: ", user.ID)
	log.Print("User Email: ", user.Email)
}

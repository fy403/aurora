package mongo_test

import (
	"aurora/internal/config"
	"aurora/internal/repo/iface"
	"aurora/internal/repo/mongo"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewRepo() (iface.Repo, error) {
	cnf := &config.Config{
		ResultBackend:   os.Getenv("MONGODB_URL"),
		ResultsExpireIn: 30,
	}
	repo, err := mongo.New(cnf)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func TestNew(t *testing.T) {

	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	repo, err := NewRepo()
	if assert.NoError(t, err) {
		assert.NotNil(t, repo)
	}
}

func TestUploadFile(t *testing.T) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	repo, err := NewRepo()
	if err != nil {
		t.Fatal(err)
	}

	err = repo.UploadFile("test.txt", []byte("test content"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateFile(t *testing.T) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	repo, err := NewRepo()
	if err != nil {
		t.Fatal(err)
	}

	err = repo.UpdateFile("test.txt", []byte("new content"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDownloadFile(t *testing.T) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	repo, err := NewRepo()
	if err != nil {
		t.Fatal(err)
	}

	fileContent, err := repo.DownloadFile("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("content: %s", fileContent)
}

func TestDeleteFile(t *testing.T) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	repo, err := NewRepo()
	if err != nil {
		t.Fatal(err)
	}

	err = repo.DeleteFile("test.txt")
	if err != nil {
		t.Fatal(err)
	}
}

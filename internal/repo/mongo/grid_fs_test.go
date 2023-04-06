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

func TestRepo(t *testing.T) {
	t.Parallel()
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	repo, err := NewRepo()
	if assert.NoError(t, err) {
		assert.NotNil(t, repo)
	}

	fileID := UploadFile(t)
	DownloadFile(t, fileID)
	UpdateFile(t, fileID)
	DownloadFile(t, fileID)
	DeleteFile(t, fileID)
}

func UploadFile(t *testing.T) string {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	repo, err := NewRepo()
	if err != nil {
		t.Fatal(err)
	}

	fileID, err := repo.UploadFile("test.txt", []byte("test content"))

	if err != nil {
		t.Fatal(err)
	}
	return fileID
}

func UpdateFile(t *testing.T, fileID string) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	repo, err := NewRepo()
	if err != nil {
		t.Fatal(err)
	}

	err = repo.UpdateFile(fileID, "test.txt", []byte("new content"))
	if err != nil {
		t.Fatal(err)
	}
}

func DownloadFile(t *testing.T, fileID string) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	repo, err := NewRepo()
	if err != nil {
		t.Fatal(err)
	}

	fileContent, err := repo.DownloadFile(fileID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("content: %s", fileContent)
}

func DeleteFile(t *testing.T, fileID string) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	repo, err := NewRepo()
	if err != nil {
		t.Fatal(err)
	}

	err = repo.DeleteFile(fileID)
	if err != nil {
		t.Fatal(err)
	}
}

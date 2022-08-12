package pdfinfo

import (
	"aurora/internal/log"
	"aurora/internal/model"
	"math"
	"os"

	gtbg "github.com/fy403/gotenberg-client-go"
	"github.com/google/uuid"
)

func init() {
	model.ExtantTaskMap["pdf_pages"] = PdfPages
}

// pdf-pages
func PdfPages(args ...string) (int64, error) {
	url, ok := os.LookupEnv("GOTENBERG_URL")
	if !ok {
		url = "http://localhost:3000/forms/libreoffice/convert"
		log.Runtime().Infof("Can`t find GOTENBERG_URL in env, use default url: %s", url)
	}
	formKey := "files"
	extraParams := map[string]string{
		"merge":     "true",
		"pdfFormat": "PDF/A-1a",
	}
	saveDirName := os.TempDir()
	saveFileName := "merge_" + uuid.NewString()

	gotenberg := gtbg.NewGotenberg(url)
	request, err := gotenberg.NewRequest(extraParams, formKey, args...)
	if err != nil {
		return math.MaxInt64, err
	}
	savePdfPath, err := gotenberg.Send(request, saveDirName, saveFileName)
	if err != nil {
		return math.MaxInt64, err
	}
	pages, err := gotenberg.Pdfpages(savePdfPath)
	if err != nil {
		return math.MaxInt64, err
	}
	return int64(pages), nil
}

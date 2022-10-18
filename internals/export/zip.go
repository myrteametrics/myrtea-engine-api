package export

import (
	"bytes"
	"io"

	"github.com/alexmullins/zip"
	"go.uber.org/zap"
)

func CreatePasswordProtectedZipFile(zipFileName string, contents []byte) ([]byte, error) {
	// create a buffer to write our archive to
	buf := new(bytes.Buffer)

	// create a new zip archive writer with password
	zipw := zip.NewWriter(buf)
	w, err := zipw.Encrypt(zipFileName, "5URETE758570?")
	if err != nil {
		zap.L().Error("Error Creating Zip File", zap.Error(err))
		return contents, err
	}
	_, err = io.Copy(w, bytes.NewReader(contents))
	if err != nil {
		zap.L().Error("Error Creating Zip File", zap.Error(err))
		return contents, err
	}
	zipw.Close()

	return buf.Bytes(), nil
}

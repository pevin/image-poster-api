package request

import (
	"bytes"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/grokify/go-awslambda"
)

type MultipartValues struct {
	Body          io.Reader
	Filename      string
	FileExtension string
	ContentType   string
	Caption       string
}

func GetMultipartValues(req events.APIGatewayProxyRequest, fileFieldName string) (mv MultipartValues, err error) {
	r, err := awslambda.NewReaderMultipart(req)
	if err != nil {
		log.Printf("Error in creating reader multipart: %s", err)
		return
	}

	for {
		p, fErr := r.NextPart()
		if fErr == io.EOF {
			break
		}
		if fErr != nil {
			return mv, fErr
		}
		if p.FormName() == fileFieldName {
			pC, fErr := io.ReadAll(p)
			if fErr != nil {
				return mv, fErr
			}
			mv.Body = bytes.NewReader(pC)
			mv.Filename = p.FileName()
			split := strings.Split(mv.Filename, ".")
			mv.FileExtension = split[len(split)-1]
			mv.ContentType = p.Header.Get("Content-Type")
		}
		if p.FormName() == "caption" {
			pC, fErr := io.ReadAll(p)
			if fErr != nil {
				return mv, fErr
			}
			mv.Caption = string(pC)
		}
	}
	return
}

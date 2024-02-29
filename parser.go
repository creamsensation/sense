package sense

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"

	"github.com/creamsensation/form"
)

type ParseContext interface {
	File(filename string) (form.Multipart, error)
	Files(filesname ...string) ([]form.Multipart, error)
	Json(target any) error
	Text() (string, error)
	Xml(target any) error
}

type parser struct {
	req   *http.Request
	bytes []byte
	limit int64
}

func (p *parser) Text() (string, error) {
	if len(p.bytes) > 0 {
		return string(p.bytes), nil
	}
	bytes, err := io.ReadAll(p.req.Body)
	return string(bytes), err
}

func (p *parser) Json(value any) error {
	if len(p.bytes) > 0 {
		return json.Unmarshal(p.bytes, value)
	}
	return json.NewDecoder(p.req.Body).Decode(value)
}

func (p *parser) Xml(value any) error {
	if len(p.bytes) > 0 {
		return xml.Unmarshal(p.bytes, value)
	}
	return xml.NewDecoder(p.req.Body).Decode(value)
}

func (p *parser) File(filename string) (form.Multipart, error) {
	if len(p.bytes) > 0 {
		return form.Multipart{}, nil
	}
	err := p.parseMultipartForm()
	if err != nil {
		return form.Multipart{}, err
	}
	multiparts, err := p.createMultiparts(filename)
	if err != nil {
		return form.Multipart{}, err
	}
	if len(multiparts) == 0 {
		return form.Multipart{}, nil
	}
	return multiparts[0], nil
}

func (p *parser) Files(filesname ...string) ([]form.Multipart, error) {
	if len(p.bytes) > 0 {
		return []form.Multipart{}, nil
	}
	err := p.parseMultipartForm()
	if err != nil {
		return []form.Multipart{}, err
	}
	multiparts, err := p.createMultiparts(filesname...)
	if err != nil {
		return []form.Multipart{}, err
	}
	return multiparts, nil
}

func (p *parser) createMultiparts(filename ...string) ([]form.Multipart, error) {
	var fn string
	if len(filename) > 0 {
		fn = filename[0]
	}
	fnLen := len(fn)
	result := make([]form.Multipart, 0)
	for name, files := range p.req.MultipartForm.File {
		if fnLen > 0 && name != fn {
			continue
		}
		for _, file := range files {
			f, err := file.Open()
			if err != nil {
				return result, errors.Join(ErrorOpenFile, err)
			}
			data, err := io.ReadAll(f)
			if err != nil {
				return result, errors.Join(ErrorReadData, err)
			}
			result = append(
				result, form.Multipart{
					Key:    name,
					Name:   file.Filename,
					Type:   http.DetectContentType(data),
					Suffix: getFileSuffixFromName(file.Filename),
					Data:   data,
				},
			)
		}
	}
	return result, nil
}

func (p *parser) parseMultipartForm() error {
	if !isRequestMultipart(p.req) {
		return ErrorInvalidMultipart
	}
	return p.req.ParseMultipartForm(p.limit << 20)
}

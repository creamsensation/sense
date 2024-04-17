package sense

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"reflect"
	
	"github.com/creamsensation/form"
)

type ParseContext interface {
	QueryParam(key string, target any) error
	PathValue(key string, target any) error
	File(filename string) (form.Multipart, error)
	Files(filesnames ...string) ([]form.Multipart, error)
	Json(target any) error
	Text() (string, error)
	Xml(target any) error
	Url(target any) error
	
	MustQueryParam(key string, target any)
	MustPathValue(key string, target any)
	MustFile(filename string) form.Multipart
	MustFiles(filesnames ...string) []form.Multipart
	MustJson(target any)
	MustText() string
	MustXml(target any)
	MustUrl(target any)
}

type parser struct {
	req   *http.Request
	bytes []byte
	limit int64
}

func (p *parser) QueryParam(key string, target any) error {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Ptr {
		return ErrorPointerTarget
	}
	v := reflect.ValueOf(target)
	if !p.req.URL.Query().Has(key) {
		return ErrorQueryParamMissing
	}
	stringValue := p.req.URL.Query().Get(key)
	setValueToReflected(t.Elem().Kind(), v.Elem(), stringValue)
	return nil
}

func (p *parser) MustQueryParam(key string, target any) {
	err := p.QueryParam(key, target)
	if err != nil {
		panic(err)
	}
}

func (p *parser) PathValue(key string, target any) error {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Ptr {
		return ErrorPointerTarget
	}
	v := reflect.ValueOf(target)
	pathValue := p.req.PathValue(key)
	if len(pathValue) == 0 {
		return ErrorPathValueMissing
	}
	setValueToReflected(t.Elem().Kind(), v.Elem(), pathValue)
	return nil
}

func (p *parser) MustPathValue(key string, target any) {
	err := p.PathValue(key, target)
	if err != nil {
		panic(err)
	}
}

func (p *parser) Url(target any) error {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer")
	}
	v := reflect.ValueOf(target)
	for i := 0; i < t.Elem().NumField(); i++ {
		queryKey := t.Elem().Field(i).Tag.Get("query")
		if len(queryKey) > 0 && p.req.URL.Query().Has(queryKey) {
			queryParam := p.req.URL.Query().Get(queryKey)
			setValueToReflected(t.Elem().Field(i).Type.Kind(), v.Elem().Field(i), queryParam)
		}
		pathKey := t.Elem().Field(i).Tag.Get("path")
		pathValue := p.req.PathValue(pathKey)
		if len(pathKey) > 0 && len(pathValue) > 0 {
			setValueToReflected(t.Elem().Field(i).Type.Kind(), v.Elem().Field(i), pathValue)
		}
	}
	return nil
}

func (p *parser) MustUrl(target any) {
	err := p.Url(target)
	if err != nil {
		panic(err)
	}
}

func (p *parser) Text() (string, error) {
	if len(p.bytes) > 0 {
		return string(p.bytes), nil
	}
	if p.req.Body == nil {
		return "", nil
	}
	bytes, err := io.ReadAll(p.req.Body)
	return string(bytes), err
}

func (p *parser) MustText() string {
	r, err := p.Text()
	if err != nil {
		panic(err)
	}
	return r
}

func (p *parser) Json(target any) error {
	if len(p.bytes) > 0 {
		return json.Unmarshal(p.bytes, target)
	}
	if p.req.Body == nil {
		return nil
	}
	return json.NewDecoder(p.req.Body).Decode(target)
}

func (p *parser) MustJson(target any) {
	err := p.Json(target)
	if err != nil {
		panic(err)
	}
}

func (p *parser) Xml(value any) error {
	if len(p.bytes) > 0 {
		return xml.Unmarshal(p.bytes, value)
	}
	if p.req.Body == nil {
		return nil
	}
	return xml.NewDecoder(p.req.Body).Decode(value)
}

func (p *parser) MustXml(target any) {
	err := p.Xml(target)
	if err != nil {
		panic(err)
	}
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

func (p *parser) MustFile(filename string) form.Multipart {
	file, err := p.File(filename)
	if err != nil {
		panic(err)
	}
	return file
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

func (p *parser) MustFiles(filesnames ...string) []form.Multipart {
	files, err := p.Files(filesnames...)
	if err != nil {
		panic(err)
	}
	return files
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

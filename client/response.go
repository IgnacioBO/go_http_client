package client

//Los Get, Post del build devuelve un response

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
)

// Struct que tiene response, error y body
type Response struct {
	*http.Response
	Err      error
	byteBody []byte
}

func (r *Response) String() string {
	return string(r.Bytes())
}

func (r *Response) Bytes() []byte {
	return r.byteBody
}

func (r *Response) SetBytes(bytes []byte) {
	r.byteBody = bytes
}

// Metodo que permite pasarle una interfaz y rellenar la interfaz con LA RESPUESTA en JSON
// Osea decodifica el JSON/XML que recibi cono Respone a una struct X y LLENA un STRUCT
func (r *Response) FillUp(fill interface{}) error {
	ctypeJSON := "application/json"
	ctypeXML := "application/xml"

	ctype := strings.ToLower(r.Header.Get("Content-Type"))

	for i := 0; i < 2; i++ {

		switch {
		case strings.Contains(ctype, ctypeJSON):
			return json.Unmarshal(r.byteBody, fill)
		case strings.Contains(ctype, ctypeXML):
			return xml.Unmarshal(r.byteBody, fill)
		case i == 0:
			ctype = http.DetectContentType(r.byteBody)
		}

	}

	return errors.New("response format neither JSON nor XML")
}

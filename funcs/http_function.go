package funcs
import (
	"net/http"
	"bytes"
	"fmt"
)
// ResponseWriterWrapper struct is used to log the response
type ResponseWriterWrapper struct {
    W          *http.ResponseWriter
    Body       *bytes.Buffer
    StatusCode *int
}

// NewResponseWriterWrapper static function creates a wrapper for the http.ResponseWriter
func NewResponseWriterWrapper(w http.ResponseWriter) ResponseWriterWrapper {
    var buf bytes.Buffer
    var statusCode int = 200
    return ResponseWriterWrapper{
        W:          &w,
        Body:       &buf,
        StatusCode: &statusCode,
    }
}

func (rww ResponseWriterWrapper) Write(buf []byte) (int, error) {
    rww.Body.Write(buf)
    return (*rww.W).Write([]byte{})
}

// Header function overwrites the http.ResponseWriter Header() function
func (rww ResponseWriterWrapper) Header() http.Header {
    return (*rww.W).Header()
}

// WriteHeader function overwrites the http.ResponseWriter WriteHeader() function
func (rww ResponseWriterWrapper) WriteHeader(statusCode int) {
    (*rww.StatusCode) = statusCode
    (*rww.W).WriteHeader(statusCode)
}

func (rww ResponseWriterWrapper) String() string {
    var buf bytes.Buffer
    buf.WriteString("Response:")
    buf.WriteString("Headers:")
    for k, v := range (*rww.W).Header() {
        buf.WriteString(fmt.Sprintf("%s: %v", k, v))
    }

    buf.WriteString(fmt.Sprintf(" Status Code: %d", *(rww.StatusCode)))

    buf.WriteString("Body")
	fmt.Println("haha!",rww.Body.String())
    buf.WriteString(rww.Body.String())
    return buf.String()
}
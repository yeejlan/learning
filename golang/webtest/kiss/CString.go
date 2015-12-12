
package kiss

import( 
    "bufio"
    "io"
    "fmt"
    "bytes"
)


//ascii strings + "\0" is called a c string
func ReadCString(r io.Reader) []byte {
    if r==nil {
        return nil
    }
    
    reader := bufio.NewReader(r)
	b, err := reader.ReadBytes(byte(0))
	if err != nil {
	    return nil
	}
	return b[:len(b)-1]
}

func WriteCString(w io.Writer, b []byte) (int64, error) {
    if w == nil {
        return 0, NewCStringError("Writer is nil")
    }
    var buf bytes.Buffer
    buf.Write(b)
    buf.Write([]byte{byte(0)})
    
    return buf.WriteTo(w)
}

type CStringError struct {
	Message string
}

func NewCStringError(format string, args ...interface{}) CStringError {
	return CStringError{fmt.Sprintf(format, args...)}
}

func (this CStringError) Error() string {
	return this.Message
}


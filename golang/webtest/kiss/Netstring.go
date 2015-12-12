
package kiss

import( 
    "bufio"
    "io"
    "strconv"
    "fmt"
    "bytes"
)

//[len]":"[string]"," is called a netstring.i.e., "12:hello world!,". The empty string is encoded as "0:,".

func ReadNetstring(r io.Reader) ([]byte, error ) {
    if r==nil {
        return nil, NewNetstringError("Reader is nil")
    }
    
    reader := bufio.NewReader(r)
    
    lenstr, err := reader.ReadString(':') 
    if err != nil {
        return nil, err
    }
    var count int
    count, err = strconv.Atoi(lenstr[0:len(lenstr)-1])
    if err != nil {
        return nil, err
    }
    //read real content
	b := make([]byte, count)
	total := 0
	for i := 0; i < 100; i++ {
		n, err := reader.Read(b[total:])
		if err != nil {
			return nil, err
		}
		total += n
		if total >= count {
			break
		}
	}
	if total != count {
		return nil, NewNetstringError("Unifinished read")
	}
	
	var endchar byte
    endchar, err = reader.ReadByte()
    if err != nil {
        return nil, err
    }
    if endchar != ',' {
        return nil, NewNetstringError("Wrong netstring ending")
    }
	return b, nil
}

func WriteNetstring(w io.Writer, b []byte) (int64, error) {
    if w == nil {
        return 0, NewNetstringError("Writer is nil")
    }
    var buf bytes.Buffer
    len := strconv.Itoa(len(b))
    buf.Write([]byte(len + ":"))
    buf.Write(b)
    buf.Write([]byte(","))
    
    return buf.WriteTo(w)
}

//additional function: just read a line 
func ReadLineString(r io.Reader) []byte {
    if r==nil {
        return nil
    }
    
    reader := bufio.NewReader(r)
	l, isPrefix, err := reader.ReadLine()
	if isPrefix || err != nil {
		return nil
	}
	return l
}

func WriteLineString(w io.Writer, b []byte) (n int64, err error) {
    if w == nil {
        return 0, NewNetstringError("Writer is nil")
    }
    var buf bytes.Buffer
    buf.Write(b)
    buf.Write([]byte("\n"))
    
    n, err = buf.WriteTo(w)
    if err != nil {
        fmt.Println("WriteLineString error: %s", err)
    }
    return
}

type NetstringError struct {
	Message string
}

func NewNetstringError(format string, args ...interface{}) NetstringError {
	return NetstringError{fmt.Sprintf(format, args...)}
}

func (this NetstringError) Error() string {
	return this.Message
}

package kiss

import(
    "os"
    "io"
    "runtime"
    "math"
    "fmt"
    "encoding/gob"
    "encoding/ascii85"
    "bytes"
)

type worker struct{
    process *os.Process
    stdin io.WriteCloser
    stdout io.ReadCloser
    status string
    version string
}

type workerMessage struct{
    Flag int
    Prefix string
    Msg string
}

const(
    workerMsg int = iota + 1
    workerStatus
    workerQuit
    workerGracefulQuit
)

func getWorkerStatusStr() string {
    pid := os.Getpid()
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    memUsed := humanSizeStr(float64(stats.Alloc))
    return fmt.Sprintf("pid: %d, memory usage: %s", pid, memUsed)
}


func humanSizeStr(size float64) string {
    var q = []string{"B", "K", "M", "G"}
    var s string
    for idx, val := range q {
        step := math.Pow(1024, float64(idx))
        if size > step { 
            s = fmt.Sprintf("%.2f%s", size / step, val)
        }
    }
    return s
}

func encodeWorkerMsg(q *workerMessage) []byte{
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    enc.Encode(&q)
    
    dst := make([]byte, ascii85.MaxEncodedLen(buf.Len()))
    dl := ascii85.Encode(dst, buf.Bytes())
        return dst[:dl]
}

func decodeWorkerMsg(msg []byte) *workerMessage{
    var q workerMessage

    dst := make([]byte, len(msg))
    dl, _, _ := ascii85.Decode(dst, msg, true)
    dec := gob.NewDecoder(bytes.NewBuffer(dst[:dl]))
    dec.Decode(&q)
    return &q
}


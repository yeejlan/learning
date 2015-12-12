
package kiss

import(
    "os"
    "time"  
    "net"
    "../lib/relog"
    "fmt"
)


var loggerWorker = &worker{}

//this function need to be called in a worker process
func processLoggerLoop(socket string) {
    stdin := os.NewFile(uintptr(3), "")
    stdout := os.NewFile(uintptr(4), "")
    
    go serveLogger(socket)
    
    go func(){
        time.Sleep(1e8)
        WriteCString(stdout, encodeWorkerMsg(&workerMessage{workerMsg, "", "ready"}))
    }()

    //manager message handler
    var running = true
    for running {
        data := ReadCString(stdin)
        if data == nil {
            continue
        }
        q := decodeWorkerMsg(data)
        switch q.Flag {
            case workerQuit:
                running = false
            case workerStatus:
                WriteCString(stdout, encodeWorkerMsg(&workerMessage{workerMsg, "", getWorkerStatusStr()}))
        }
    }

}

func serveLogger(socket string){
	var la *net.UnixAddr
	var err error
	if la, err = net.ResolveUnixAddr("unix", socket); err != nil {
		relog.Fatal("ResolveAddr error: %s", err)
	}
	var loggerListener *net.UnixListener
    loggerListener, err = net.ListenUnix("unix", la)
    if err != nil {
        relog.Fatal("listen error: %s", err)
    }    
    
    //run
    for {
        conn, err := loggerListener.Accept()
		if err != nil {
		    relog.Warning("accept error: %s", err)
		    continue
		}
        
        go func(c net.Conn) {
            data := ReadCString(c)
            q := decodeWorkerMsg(data)
            if q.Flag == workerMsg{
                logger := GetLogger(q.Prefix)
                logger.Log(q.Msg)                        
            }
            c.Close()
        }(conn)
    } 
}

func stopProcessLogger(){
    if loggerWorker.stdin == nil {
        return
    }

    //send shutdown message
    WriteCString(loggerWorker.stdin, encodeWorkerMsg(&workerMessage{Flag : workerQuit}))

    //wait logger process
    if _, err := loggerWorker.process.Wait(); err != nil {
        relog.Warning(err.Error())
    }
    
    //close goroutine logger
    LoggerCloseAll()    
}


func sendlog(prefix, format string, args ...interface{}){
    c, err := net.Dial("unix", loggerSocket)
    if err != nil {
        return
    }
    WriteCString(c, encodeWorkerMsg(&workerMessage{Flag : workerMsg, Prefix : prefix, Msg : fmt.Sprintf(format, args...)}))
    c.Close()
}

func LogDebug(format string, args ...interface{}){
    sendlog("debug", format, args ...)
}

func LogInfo(format string, args ...interface{}){
    sendlog("info", format, args ...)
}

func LogWarning(format string, args ...interface{}){
    sendlog("warning", format, args ...)
}

func LogError(format string, args ...interface{}){
    sendlog("error", format, args ...)
}





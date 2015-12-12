
package kiss

import(
    "container/list"    
	"net"
	"net/http"
	"os"
	"sync"
	"time"
	"../lib/relog"
	"fmt"
)

const maxWaitTimeBeforeQuit = 10 * time.Minute

type webWorkerList struct{
    workers *list.List
    mu sync.Mutex
}

var webWorkers = &webWorkerList{
    workers : list.New(),
}

func (this *webWorkerList) append(w *worker) {
    this.workers.PushBack(w)
}

func (this *webWorkerList) remove(e *list.Element) {
    webWorkers.mu.Lock()
    defer webWorkers.mu.Unlock()
    
    this.workers.Remove(e)
}

var stopAccepting = make(chan int)
var handlerWaitGroup = new(sync.WaitGroup)
var finishWaiting = make(chan int)

//this function need to be called in a worker process
func webWorkerLoop(l net.Listener, handler http.Handler) {
    stdin := os.NewFile(uintptr(3), "")
    stdout := os.NewFile(uintptr(4), "")
    
    go serveFcgi(l, handler)
    
    go func(){
        time.Sleep(1e8)
        WriteCString(stdout, encodeWorkerMsg(&workerMessage{workerMsg, "", "ready"}))
    }()

    //manager message handler
    var running = true
    for running {
        data := ReadCString(stdin)
        q := decodeWorkerMsg(data)
        switch q.Flag {
            case workerQuit:
                running = false
            case workerStatus:
                WriteCString(stdout, encodeWorkerMsg(&workerMessage{workerMsg, "", getWorkerStatusStr()}))
            case workerGracefulQuit:
                go func(){ //force quit when timeout
                    time.Sleep(maxWaitTimeBeforeQuit)
                    os.Exit(0)
                }()           
                stopAccepting <- 1 //graceful quit
                <- finishWaiting
                running = false
        }
    }
}

///**this funcion should be called by management process*/ 
func startWebWorker(listenerFile *os.File, version string) *worker {
    webWorkers.mu.Lock()
    defer webWorkers.mu.Unlock()
    
    mypath := os.Args[0]
    singleWorker := &worker{}
    inpr, inpw, err := os.Pipe()
    if err != nil {
        relog.Fatal(err.Error())
    }
    singleWorker.stdin = inpw
    
    stdr, stdw, err := os.Pipe()
    if err != nil {
        relog.Fatal(err.Error())
    }
    singleWorker.stdout = stdr
    
    args := []string{mypath, "startwebworker"}

    process, err := os.StartProcess(mypath, args, &os.ProcAttr{
        Dir:   "",
        Files: []*os.File{os.Stdin, os.Stdout, os.Stderr, inpr, stdw, listenerFile},
        Env:   os.Environ(),
        Sys:   nil,
    })
    if err != nil {
        relog.Fatal(err.Error())
    }
    singleWorker.process = process
    inpr.Close()
    stdw.Close()
    
    //wait for ready
    data := ReadCString(singleWorker.stdout)
    q := decodeWorkerMsg(data)
    if q.Msg != "ready" {
        relog.Fatal("Web worker start error")
    }
    
    singleWorker.status = "running"
    singleWorker.version = version
    webWorkers.append(singleWorker)
    relog.Info("Web worker(pid=%d) is ready.", singleWorker.process.Pid)
    return singleWorker
}

func reloadWebWorkers(listenerFile *os.File){
    version := GetUniqueID()
    for i:=0; i<concurrentWebWorkers; i++ {
        startWebWorker(listenerFile, version)
    }        
    
    webWorkers.mu.Lock()
    defer webWorkers.mu.Unlock()    
    for e := webWorkers.workers.Front(); e != nil; e = e.Next() {
        w := e.Value.(*worker)
        
        if w.version != version && w.status != "quitting" {
            //send graceful stop message
            WriteCString(w.stdin, encodeWorkerMsg(&workerMessage{Flag : workerGracefulQuit}))
            w.status = "quitting"
            //wait quit signal
            go func(e *list.Element){
                w.process.Wait()
                webWorkers.remove(e)
            }(e)
        }
    }    
}

func serveFcgi(l net.Listener, handler http.Handler) {
    var accepting = true
	for accepting {
	    l.(*net.TCPListener).SetDeadline(time.Now().Add(3e9))
		rw, err := l.Accept()
		if err != nil {
			if ope, ok := err.(*net.OpError); ok {
				if !(ope.Timeout() && ope.Temporary()) {
					relog.Warning("accept error: %s", ope)
				}
			} else {
				relog.Warning("accept error: %s", err)
			}
		}else{
            c := newChild(rw, handler)
            handlerWaitGroup.Add(1)
            go serve(c)
		}
		
		select {
            case <-stopAccepting:
                accepting = false
            default:
		}
	}
	// wait for handlers
	changeWorkerStatus(os.Getpid(), "waiting")
	handlerWaitGroup.Wait()
	finishWaiting <- 1
}

func serve(c *child) {
    defer handlerWaitGroup.Done()
	defer c.conn.Close()
	var rec record
	for {
		if err := rec.read(c.conn.rwc); err != nil {
			return
		}
		if err := c.handleRecord(&rec); err != nil {
			return
		}
	}
}

func changeWorkerStatus(pid int, status string){
    webWorkers.mu.Lock()
    defer webWorkers.mu.Unlock()

    for e := webWorkers.workers.Front(); e != nil; e = e.Next() {
        w := e.Value.(*worker)
        
        if w.process != nil && w.process.Pid == pid {
           w.status = status
           break
        }
    } 
}

func stopFcgiWorkers() {
    webWorkers.mu.Lock()
    defer webWorkers.mu.Unlock()
    
    for e := webWorkers.workers.Front(); e != nil; e = e.Next() {
        w := e.Value.(*worker)
        
        if w.stdin != nil {
            //send shutdown message
            WriteCString(w.stdin, encodeWorkerMsg(&workerMessage{Flag : workerQuit}))
            
            //wait exit
            if _, err := w.process.Wait(); err != nil {
                relog.Warning(err.Error())
            }
        }
    }
}

func webWorkerStatus(c net.Conn) {
    webWorkers.mu.Lock()
    defer webWorkers.mu.Unlock()
    
    cnt := 1
    for e := webWorkers.workers.Front(); e != nil; e = e.Next() {
        w := e.Value.(*worker)
        
        WriteCString(w.stdin, encodeWorkerMsg(&workerMessage{Flag : workerStatus}))
        data := ReadCString(w.stdout)
        q := decodeWorkerMsg(data)   
        fmt.Fprintf(c, "Web worker#%d: ", cnt)
        if q.Msg != "" {
            fmt.Fprintf(c, "%s, ", q.Msg)
        }
        fmt.Fprintf(c, "status: %s\r\n", w.status)
        cnt++
    }        
}

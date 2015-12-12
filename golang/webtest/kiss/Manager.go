
package kiss

import (
    "flag"
    "fmt"
    "os"  
    "os/signal"
    "syscall"
    "net"
    "../lib/relog"
    "time"
    "bufio"
    "io/ioutil"
)

const (
    concurrentWebWorkers = 2
    loggerSocket = "logger-socket"
    managerSocket = "manager-socket"
)

func ServeFcgi(addr string){
    
    var cmd = flag.String("c", "", "Management command")
    
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	
	if(*cmd != ""){
	    sendManagementCommand(*cmd, args)
	    os.Exit(0)
	}
	
	if len(args) < 1 || args[0] == "help" {
		usage()
	}	
	
	switch args[0] {
	    case "test": //call start with log redirection
	        CheckConfig()	   	    
	    case "dev": //call start with log redirection
	        devServer()	    
	    case "run": //call start with log redirection
	        startServer()
	    case "startmanager":
	        go handleManagerSignals()
	        startProcessLogger()
	        startWebWorkers(addr)
	        startInternalServer()
	    case "startlogger":
	        startLoggerWorkerLoop(loggerSocket)
	    case "startwebworker":
	        startWebWorkerLoop()
	    default:
	        fmt.Println("Unknown command")
	        usage()
	}   
}

func usage() {
    
	fmt.Fprintf(os.Stderr, "Please use \"run/dev\" to start a server.\r\n")
	fmt.Fprintf(os.Stderr, "\"-c cmd\" to send a management command.\r\n")
	os.Exit(2)
}

func sendManagementCommand(cmd string, args []string ){
    cmdClient, err := net.Dial("unix", managerSocket)
    if err != nil {
        relog.Fatal(err.Error())
    }
    prefix := ""
    if len(args) == 1 {
        prefix = args[0]
    }
    WriteCString(cmdClient, encodeWorkerMsg(&workerMessage{Flag : workerMsg, Prefix : prefix, Msg : cmd}))
    
    var ret []byte
    ret, err= ioutil.ReadAll(cmdClient)
    if err != nil {
        relog.Fatal(err.Error())
    }
    cmdClient.Close()
    fmt.Println(string(ret))    
}

func managerShutdown() {
    stopFcgiWorkers()
    stopProcessLogger()
    os.Remove(loggerSocket)
}

func startWebWorkerLoop() {
    mux := createServerMux()
    listenerFile := os.NewFile(uintptr(5), "")
    listener, err := net.FileListener(listenerFile)
    if err != nil {
		relog.Fatal(err.Error())
	}
	webWorkerLoop(listener, mux)
}


func startLoggerWorkerLoop(socket string) {
    processLoggerLoop(socket)
}

func restartLoggerProcess() {//which may not be used in normal case
    loggerWorker.process.Kill()
    loggerWorker.process.Wait()
    startProcessLogger()
}

func startServer()() {
    mypath := os.Args[0]
        
    stdr, stdw, err := os.Pipe()
	if err != nil {
		relog.Fatal(err.Error())
	}
	
    args := []string{mypath, "startmanager"}
	process, err := os.StartProcess(mypath, args, &os.ProcAttr{
		Dir:   "",
		Files: []*os.File{os.Stdin, stdw, stdw},
		Env:   os.Environ(),
		Sys:   nil,
	})
	
	stdw.Close()
	
    go func(){
        reader := bufio.NewReader(stdr)
        for{
            b, isPrefix, err := reader.ReadLine()
            if isPrefix || err != nil {
                //pass
            }else{
                writeServerLog(b)
                fmt.Println(string(b))
            }
        }
    }()
    
    process.Wait()
    stdr.Close()
}

func devServer(){
    fmt.Println("todo: create dev server")
}

var serverLog = &Logger{
    name : "server", 
    date : "0", 
}

func writeServerLog(msg []byte){
    if msg == nil {
        return
    }
    
    var t = time.Now()
    var date = t.Format("2006_01_02")

    if serverLog.file == nil || serverLog.date != date {
        var filename = GetBasePath()+"log/"+serverLog.name+"_"+date+".txt"
        f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
        if err != nil{
            return
        }
        serverLog.date = date
        
        //close opened log file
        if serverLog.file != nil{
            serverLog.file.Close()
        }
        serverLog.file = f
    }
    serverLog.file.Write(msg)
    serverLog.file.Write([]byte("\r\n"))
}

func startProcessLogger(){
    //create a logger process   
    mypath := os.Args[0]

    inpr, inpw, err := os.Pipe()
	if err != nil {
		relog.Fatal(err.Error())
	}
    loggerWorker.stdin = inpw
    
    stdr, stdw, err := os.Pipe()
	if err != nil {
		relog.Fatal(err.Error())
	}
    loggerWorker.stdout = stdr
    
	args := []string{mypath, "startlogger"}

	process, err := os.StartProcess(mypath, args, &os.ProcAttr{
		Dir:   "",
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr, inpr, stdw},
		Env:   os.Environ(),
		Sys:   nil,
	})
    if err != nil {
        relog.Fatal(err.Error())
    }
    loggerWorker.process = process
    
    inpr.Close()
    stdw.Close()
    
    //wait for ready
    data := ReadCString(loggerWorker.stdout)
    q := decodeWorkerMsg(data)
    if q.Msg != "ready" {
        relog.Fatal("Logger worker start error")
    }          
    relog.Info("Logger process(pid=%d) is ready.", loggerWorker.process.Pid)
}

var webListenerFile *os.File
func startWebWorkers(addr string){
    
    la, err := net.ResolveTCPAddr("tcp", addr)
    if err != nil {
        relog.Fatal(err.Error())
    }
    var webListener *net.TCPListener
    webListener, err = net.ListenTCP("tcp", la)
    if err != nil {
        relog.Fatal(err.Error())
    }
    webListenerFile, err = webListener.File()
    if err != nil {
        relog.Fatal(err.Error())
    }

    for i:=0; i<concurrentWebWorkers; i++ {
        startWebWorker(webListenerFile, "")    
    }
}

var internalListener net.Listener
func startInternalServer(){
	var la *net.UnixAddr
	var err error
	if la, err = net.ResolveUnixAddr("unix", managerSocket); err != nil {
		relog.Fatal("ResolveAddr error: %s", err)
	}
	var mgrListener *net.UnixListener
    mgrListener, err = net.ListenUnix("unix", la)
    if err != nil {
        relog.Fatal("listen error: %s", err)
    }    
    
    relog.Info("Manager(pid=%d) is ready.", os.Getpid())
    //run
    for {
        conn, err := mgrListener.Accept()
		if err != nil {
		    relog.Warning("accept error: %s", err)
		    continue
		}
        
        go func(c net.Conn) {
            data := ReadCString(c)
            q := decodeWorkerMsg(data)
            if q.Flag == workerMsg{
                handleMrgMessage(c, q.Msg)              
            }
            c.Close()
        }(conn)
    }
}

func managerUsage(c net.Conn){
    fmt.Fprintf(c, "Avaliable Commands(with -c option): \r\n")
    fmt.Fprintf(c, "ping: return pong\r\n")
    fmt.Fprintf(c, "status: get server status\r\n")
    fmt.Fprintf(c, "stop: shutdown server\r\n")
    fmt.Fprintf(c, "reload: reload http worker\r\n")
    fmt.Fprintf(c, "restartlogger: reload logger worker, should not used in normal case\r\n")
}

func handleMrgMessage(c net.Conn, msg string){
    switch msg {
        case "ping":
            fmt.Fprintf(c, "pong\r\n")
        case "status":
            managerStatus(c)
        case "stop":
            managerStop(c)
        case "reload":
            reloadWebWorker(c)
        case "restartlogger":
            restartloggerWorker(c)
        default:
            managerUsage(c)
    }
}

func managerStatus(c net.Conn){
    //process logger status
    WriteCString(loggerWorker.stdin, encodeWorkerMsg(&workerMessage{Flag : workerStatus}))
    data := ReadCString(loggerWorker.stdout)
    q := decodeWorkerMsg(data)   
    fmt.Fprintf(c, "Logger worker: ")
    fmt.Fprintf(c, "%s\r\n", q.Msg)
    
    //http worker status 
    webWorkerStatus(c)
}

func managerStop(c net.Conn){
    relog.Info("Stop message received, shutdown server.")
    managerShutdown()
    if internalListener != nil {    
        internalListener.Close()
    }
    
    msg := "Server stopped."
    relog.Info(msg)
    fmt.Fprintf(c, "%s\r\n", msg)
    go func(){
        time.Sleep(1e8)
        os.Remove(managerSocket)
        os.Exit(0)
    }()
}

func restartloggerWorker(c net.Conn){
    relog.Info("RestartLogger message received.")
    fmt.Fprintf(c, "Some log may lost before logger restarted.\r\n")
    restartLoggerProcess()
    fmt.Fprintf(c, "Logger started.\r\n")
}

func reloadWebWorker(c net.Conn){
    relog.Info("Reload webwork message received.")
    reloadWebWorkers(webListenerFile)
    fmt.Fprintf(c, "reloaded.\r\n")
}

func handleManagerSignals() {
	var sig os.Signal
	var sigChan = make(chan os.Signal, 16)
	signal.Notify(sigChan, syscall.SIGINT)
	for {
		sig = <-sigChan
		switch sig {
		    case syscall.SIGINT:
		        managerShutdown()
                go func(){
                    time.Sleep(1e8)
                    os.Remove(managerSocket)
                    os.Exit(0)
                }()
		}
	}
}


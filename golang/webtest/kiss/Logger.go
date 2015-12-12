
package kiss

import(
	"fmt"
	"time"
	"sync"
	"os"
)

type Logger struct{
	name string
	date string
	ch chan string
	file *os.File
	running bool
}

var loggerPool = make(map[string] *Logger)
var loggerPoolMutex sync.Mutex

func GetLogger(loggerName string) *Logger{
	loggerPoolMutex.Lock()
	defer loggerPoolMutex.Unlock()
	
	oldLogger, ok := loggerPool[loggerName]
	if ok{
		return oldLogger
	}
	
	loggerChannel := make(chan string)	
	
	newLogger := &Logger{
		name : loggerName, 
		date : "0", 
		ch : loggerChannel,
		file : nil,
		running : true,
	}
	
	loggerPool[loggerName] = newLogger
	go writeLog(newLogger)
	
	return newLogger
}

func writeLog(logger *Logger){
	for logger.running {
		var str = <- logger.ch
		if str == ""{
		    continue
		}
		
		var t = time.Now()
		var date = t.Format("2006_01_02")

		if logger.file == nil || logger.date != date {
		    var filename = GetBasePath()+"log/"+logger.name+"_"+date+".txt"
		    f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		    if err != nil{
		        continue
		    }
		    logger.date = date
		    
		    //close opened log file
		    if logger.file != nil{
		        logger.file.Close()
		    }
		    logger.file = f
		}
		
	    logStr := t.Format(time.RFC3339)+" "+str+"\r\n"
	    _, err := logger.file.Write([]byte(logStr))
        if err != nil{
            continue
        }		
		//logger.file.Sync()
	}			
}


func (this *Logger) Log(format string, args ...interface{}) {
	this.ch <- fmt.Sprintf(format, args...)
}

func (this *Logger) Flush(){
	this.ch <- ""
}

func (this *Logger) Close(){
	this.Flush();
	this.running = false
	if this.file != nil{
		this.file.Close()
	}
	this.file = nil
}

func LoggerCloseAll(){
	loggerPoolMutex.Lock()
	defer loggerPoolMutex.Unlock()
	
	for key, logger := range loggerPool{
		logger.Close()
		delete(loggerPool, key)
	}
}

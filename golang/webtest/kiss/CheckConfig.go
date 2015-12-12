
package kiss

import (
    "fmt"
    "path"
    "io/ioutil"
    "strings"
    "encoding/json"
)

func CheckConfig() {
    configDir := path.Clean(path.Join(App.BasePath(), "config"))
        
    fmt.Printf("Checking config files in %s ...\r\n", configDir);
    ok := checkConfigDir(configDir)
    
    if ok{
        fmt.Println("All OK.");
    }else{
        fmt.Println("Failed.");
    }
}

func checkConfigDir(configDir string) bool {
    fileInfoList, err := ioutil.ReadDir(configDir)
    if err!=nil {
        fmt.Println("ReadDir Error: ", err);
        return false;
    }
    
    for _, fileInfo := range fileInfoList{
        if strings.HasSuffix(fileInfo.Name(), ".json"){
            
            fmt.Printf("%s ", fileInfo.Name());
            var ok = checkConfigFile(path.Join(configDir, fileInfo.Name()))
            if ok{
                fmt.Println("[PASS]");
            }else{
                fmt.Println("[FAILED]");
                return false
            }
        }
    }
    return true
}

func checkConfigFile(configFile string) bool {
	var c interface{}
	
    data, err := ioutil.ReadFile(configFile)
    if(err != nil){
    	fmt.Printf("Read file error: %s \r\n", err)
        return false
    }
    err = json.Unmarshal(data, &c)
    if err != nil {
    	fmt.Printf("Json parse error: %s \r\n", err)
        return false
    }
    for k, v:= range c.(map[string]interface{}){
    	switch vv := v.(type) {
    		case string:
    		case float64:
    		default:
    			fmt.Printf("Key:[%s] bad value assignment: %s \r\n", k, vv)
    			return false
    	}
    }
    
    return true
}


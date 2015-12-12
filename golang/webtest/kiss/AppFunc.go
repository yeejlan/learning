
package kiss

import (
    "encoding/json"
    "io/ioutil"
    "os"
    "path"
    "strings"
    "strconv"
)

func GetBasePath() string {
    exeFile := ""
    arg0 := path.Clean(os.Args[0])
    wd, _ := os.Getwd()
    if strings.HasPrefix(arg0, "/") {
        exeFile = arg0
    } else {
        exeFile = path.Join(wd, arg0)
    }
    
    basePath, _ := path.Split(exeFile)
    return basePath
}

func GetApplicationEnv() string {
    applicationEnv := os.Getenv("APPLICATION_ENV")
    if applicationEnv == ""{
        applicationEnv = "production"
    }
    return applicationEnv
}

func ParseConfig(configFile string) (ConfigMap, error) {
    config := make(ConfigMap)
    var c interface{}
    
    data, err := ioutil.ReadFile(configFile)
    if(err != nil){
        return nil, err
    }
    err = json.Unmarshal(data, &c)
    if err != nil {
        return nil, err
    }
    
    for k, v:= range c.(map[string]interface{}){
    	var section, key, value string
    	switch v.(type) {
    		case string:
    			value = v.(string)
    		case float64:
    			value = strconv.FormatFloat(v.(float64), 'f', 0, 64)
    		default:
    			value = ""
    	}
    	if value != "" {
    		ss := strings.SplitN(k, ".", 2)
    		if len(ss) == 2 {
    			section = ss[0]
    			key = ss[1]
    		}else{
    			section = ""
    			key = k
    		}
    		_, exist := config[section]
    		if exist {
    			config[section][key] = value
    		}else{
    			config[section] = map[string]string{key : value}
    		}
    	}
    }    
    
    return config, nil
}

func MergeConfig(a, b ConfigMap) ConfigMap {
    config := make(ConfigMap)
    for k, v := range a{
        config[k] = v
    }
    
    for k, v := range b{
        _, exist := config[k]
        if !exist{
            config[k] = v
        }else{
            for subk, subv := range v{
                _, exist2 := config[k][subk]
                if !exist2{
                    config[k][subk] = subv
                }
            }
        }
    }
    
    return config
}



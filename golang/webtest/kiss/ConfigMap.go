
package kiss

import(
    "strings"
    "strconv"
)

type ConfigMap map[string]map[string]string

func NewConfigMap() ConfigMap{
    return make(ConfigMap)
}

func (this ConfigMap) GetBool(section, key string) bool{
    val, exist := this.getBool(section, key)
    if exist{
        return val  
    }
    return false
}

func (this ConfigMap) GetString(section, key string) string{
    val, exist := this.getString(section, key)
    if exist{
        return val  
    }
    return ""    
}

func (this ConfigMap) GetInt(section, key string) int{
    val, exist := this.getInt(section, key)
    if exist{
        return val
    }
    return 0
}

///*-------------------private functions------------*/
func (this ConfigMap) getBool(section, key string) (val bool, exist bool) {
	sec, ok := this[section]
	if !ok {
		return false, false
	}
	value, ok := sec[key]
	if !ok {
		return false, false
	}
	
	v := strings.ToLower(value)
	if v == "y"||v == "1"||v == "true" {
		return true, true
	}
	if v == "n"||v == "0"||v == "false" {
		return false, true
	}
	return false, false
}

func (this ConfigMap) getString(section, key string) (val string, exist bool) {
	sec, ok := this[section]
	if !ok {
		return "", false
	}
	value, ok := sec[key]
	if !ok {
		return "", false
	}
	return value, true
}

func (this ConfigMap) getInt(section, key string) (val int, exist bool) {
	sec, ok := this[section]
	if !ok {
		return 0, false
	}
	value, ok := sec[key]
	if !ok {
		return 0, false
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return i, true
}



package kiss

import(
    "crypto/rand"
    "time"
    "fmt"
    "strconv"
    "../lib/relog"
)

//create time based uique id
func GetUniqueID() string{
    var u = new([16]byte)
    _, err := rand.Read(u[0:16])
    if err != nil {
      relog.Warning("rand.Read error %s", err)
    }
    
    t:= time.Now().Unix()
        return fmt.Sprintf("%x%x%x", u[0:12], t, u[12:])
}

func GetTimeFromUniqueID(uniqueId string) string {
    if len(uniqueId)!=40 {
        return ""
    }
    
    tStr := uniqueId[24:32]
    tInt, err := strconv.ParseInt(tStr, 16, 64)
    if err != nil{
        relog.Warning("ParseInt error %s", err)
    }
    
    t := time.Unix(tInt, 0).Format(time.RFC3339)
    return t
}



package kiss

import(
    "io"
    "io/ioutil"
    "os"
)

func GetUploadFileContent(ctx WebContext, formKey string) ([]byte, error){
    file, _, err := ctx.Request().FormFile(formKey)
    if err != nil { 
        return nil, err
    }
    defer file.Close()
    
    data, err := ioutil.ReadAll(file) 
    if err != nil { 
        return nil, err
    } 
    return data, nil;
}

func SaveUploadFile(ctx WebContext, formKey string, saveToFilePath string) error{
    file, _, err := ctx.Request().FormFile(formKey) 
    if err != nil { 
        return err
    } 
    defer file.Close()
    
    saveToFile, err2 := os.Create(saveToFilePath)
    if err2 != nil {
        return err2
    }
    defer saveToFile.Close()
    
    _, err = io.Copy(saveToFile, file)
    if err != nil {
        return err
    }
    
    return nil
}


func GetUploadStreamContent(ctx WebContext) ([]byte, error) {
    defer ctx.Request().Body.Close()
    
    data, err := ioutil.ReadAll(ctx.Request().Body) 
    if err != nil { 
        return nil, err
    } 
    return data, nil;
}

func SaveUploadStream(ctx WebContext, saveToFilePath string) error{
    defer ctx.Request().Body.Close()
    
    saveToFile, err2 := os.Create(saveToFilePath)
    if err2 != nil {
        return err2
    }
    defer saveToFile.Close()
    
    _, err := io.Copy(saveToFile, ctx.Request().Body)
    if err != nil {
        return err
    }
    
    return nil
}



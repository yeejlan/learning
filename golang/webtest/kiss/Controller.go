
package kiss

var ActionsMap = map[string]map[string]func(WebContext){}

func ExposeAction(section string, action string, handler func(WebContext)){
    _, exist := ActionsMap[section]
    if !exist {
        ActionsMap[section] = map[string]func(WebContext){
            action : handler,
        }
    }else{
        ActionsMap[section][action] = handler
    }
}

func ExposeActions(section string, actions map[string]func(WebContext)){
    for action, handler := range actions {
        ExposeAction(section, action, handler)
    }
}
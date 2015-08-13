
package main

import "flag"
import "fmt"
import (
	"os/exec"
)

func main(){
	flag.Parse()

	for _, value := range flag.Args() {
		var out []byte
		var err error
		out, err = exec.Command("chown", "apache.apache", "-R", "--preserve-root", value).CombinedOutput()	
		fmt.Printf("%s", out);
		if err != nil {
			fmt.Printf("chown: %s\r\n", err)
			break
		}

                out, err = exec.Command("chmod", "g+w", "-R", "--preserve-root", value).CombinedOutput()
                fmt.Printf("%s", out);
                if err != nil {
			fmt.Printf("chmod: %s\r\n", err)
                        break
                }
		
	}
}

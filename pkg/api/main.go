package api

import (
	"net/http" 
	"fmt"
)


func TestHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Hi")
}

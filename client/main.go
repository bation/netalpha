//// main.go
package main

import (
	ethr "github.com/ethrToPkg"
	"net/http"
)

var server http.Server

func main() {
	ethr.EthrRunServer() // b bandwidth c connections/s p packets/s l 延迟
}

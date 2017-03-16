package apachelog_test

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/e-XpertSolutions/go-apachelog/apachelog"
)

var accessLogs = `127.0.0.1 - - [12/Dec/2016:10:57:30 +0100] "GET /assets/img/logo.jpg HTTP/1.1" 200 50122 "http://127.0.0.1" "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.99 Safari/537.36"
`

func main() {
	r := bytes.NewBuffer([]byte(accessLogs))

	p, err := apachelog.CombinedParser(r)
	if err != nil {
		log.Fatal(err)
	}

	for {
		logEntry, err := p.Parse()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		fmt.Println(logEntry)
	}
}

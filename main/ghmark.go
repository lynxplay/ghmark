package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"lynxplay/ghmark/pgk"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
)

func main() {
	port := 33456
	if s, ok := os.LookupEnv("GHMARK_PORT"); ok {
		if i, err := strconv.Atoi(s); err == nil {
			port = i
		} else {
			fmt.Println("found environment variable GHMARK_PORT with non-int value hence using default port")
		}
	}

	fmt.Println("internal server running on port " + strconv.Itoa(port))

	fileWG := &sync.WaitGroup{}

	compiler, err := pgk.MakeHTMLCompiler()
	if err != nil {
		log.Fatal(errors.New("could not make html compiler: " + err.Error()))
	}

	chromiumClient := pgk.NewChromiumWrapper(port)

	server := pgk.NewGHServer(port)
	server.Start()

	nextRoutineID := 0
	subRoutineIDLock := &sync.Mutex{}

	files := os.Args
	for i := 1; i < len(files); i++ {
		filePath := files[i]
		fileWG.Add(1)

		go func() {
			defer fileWG.Done()

			subRoutineIDLock.Lock()
			routineID := strconv.Itoa(nextRoutineID)
			nextRoutineID += 1
			subRoutineIDLock.Unlock()

			fileDir := path.Dir(filePath)

			fileContent, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Printf("could not read input file %s: %s\n", filePath, err.Error())
			}

			os.Geteuid()

			compile, err := compiler.Compile(fileContent)

			server.Mux.HandleFunc("/"+routineID, func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(200)
				_, _ = writer.Write(compile)
			})

			if output, err := chromiumClient.DownloadPDF(routineID, path.Base(filePath), fileDir); err != nil {
				fmt.Printf("could not download pdf for %s: %s\n", filePath, err.Error())
			} else {
				fmt.Printf("Compiled pdf file %s to %s\n", filePath, output)
			}
		}()
	}
	fileWG.Wait()

	server.Stop()
}

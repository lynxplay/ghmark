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
	port := findIntEnvValue("GHMARK_PORT", 33456)
	maximumThreads := findIntEnvValue("GHMARK_THREADS", 4)

	fmt.Println("internal server running on port " + strconv.Itoa(port))

	fileWG := &sync.WaitGroup{}
	threadMaxSignalChannel := make(chan interface{})

	compiler, err := pgk.MakeHTMLCompiler()
	if err != nil {
		log.Fatal(errors.New("could not make html compiler: " + err.Error()))
	}

	chromiumClient := pgk.NewChromiumWrapper(port)

	server := pgk.NewGHServer(port)
	server.Start()

	nextRoutineID := 0
	currentAliveThreads := 0
	subRoutineIDLock := &sync.Mutex{}

	files := os.Args
	for i := 1; i < len(files); i++ {
		filePath := files[i]
		fileWG.Add(1)

		go func() {
			subRoutineIDLock.Lock()
			if currentAliveThreads >= maximumThreads {
				subRoutineIDLock.Unlock()
				<-threadMaxSignalChannel
			} else {
				subRoutineIDLock.Unlock()
			}

			subRoutineIDLock.Lock()
			routineID := strconv.Itoa(nextRoutineID)
			nextRoutineID += 1
			currentAliveThreads += 1
			subRoutineIDLock.Unlock()

			fileDir := path.Dir(filePath)

			fileContent, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Printf("could not read input file %s: %s\n", filePath, err.Error())
			}

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

			fileWG.Done() // we do this first if no more threads are here to take from max signal channel

			subRoutineIDLock.Lock()
			currentAliveThreads -= 1
			threadMaxSignalChannel <- true
			subRoutineIDLock.Unlock()
		}()
	}

	fileWG.Wait()
	server.Stop()
}

func findIntEnvValue(key string, defaultValue int) int {
	if s, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(s); err == nil {
			defaultValue = i
		} else {
			fmt.Println("found environment variable " + key + " with non-int value hence using default " + strconv.Itoa(defaultValue))
		}
	}
	return defaultValue
}

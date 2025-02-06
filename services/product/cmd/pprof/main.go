package main

import (
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	q := make(chan os.Signal, 1)

	signal.Notify(q, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go func() {
		log.Println("pprof сервер запущен на http://localhost:8081/debug/pprof/")
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Printf("failed to start pprof server: %v", err)
		}
	}()

	<-q
}

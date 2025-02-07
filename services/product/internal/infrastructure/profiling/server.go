package profiling

import (
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/joho/godotenv"
)

func Run(addr string) {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	log.Println("pprof running on " + addr + "/debug/pprof/")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("failed to start pprof server: %v", err)
	}

}

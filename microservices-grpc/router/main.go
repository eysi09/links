package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/campoy/links/microservices-grpc/repository"
	pb "github.com/campoy/links/microservices-grpc/repository/proto"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
)

var links pb.RepositoryClient

func main() {
	var config struct {
		Address    string `default:"0.0.0.0:8085"`
		Repository string `default:"0.0.0.0:8080"`
	}
	if err := envconfig.Process("ROUTER", &config); err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.Dial(config.Repository, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	links = pb.NewRepositoryClient(conn)
	log.Printf("connecting to repository on %s", config.Repository)

	http.HandleFunc("/l/", handleVisit)
	http.HandleFunc("/s/", handleStats)
	log.Printf("listening on %s", config.Address)
	log.Fatal(http.ListenAndServe(config.Address, nil))
}

func handleVisit(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[3:]
	l, err := links.Get(r.Context(), &pb.IDRequest{Id: id})
	if err != nil {
		if err == repository.ErrNoSuchLink {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	_, err = links.CountVisit(r.Context(), &pb.IDRequest{Id: id})
	if err != nil {
		log.Printf("could not count visit: %v", err)
	}

	fmt.Fprintf(w, "<p>redirecting to %s...</p>", l.Url)
	fmt.Fprintf(w, "<script>setTimeout(function() { window.location = '%s'}, 1000)</script>", l.Url)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[3:]
	l, err := links.Get(r.Context(), &pb.IDRequest{Id: id})
	if err != nil {
		if err == repository.ErrNoSuchLink {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(l); err != nil {
		log.Printf("could not encode link information")
	}
}

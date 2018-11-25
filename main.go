// Sample logging-quickstart writes a log entry to Stackdriver Logging.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	// Imports the Stackdriver Logging client package.
	"cloud.google.com/go/logging"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

const logName = "app_logs"

var (
	projectID    string
	requestCount int
	monRes       *monitoredres.MonitoredResource
)

func main() {
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	monRes = &monitoredres.MonitoredResource{
		Labels: map[string]string{
			"module_id":  os.Getenv("GAE_SERVICE"),
			"project_id": projectID,
			"version_id": os.Getenv("GAE_VERSION"),
		},
		Type: "gae_app",
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/nolog", nolog)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func traceID(r *http.Request) string {
	return fmt.Sprintf("projects/%s/traces/%s", projectID, strings.Split(r.Header.Get("X-Cloud-Trace-Context"), "/")[0])
}

func newClient(ctx context.Context) *logging.Client {
	client, err := logging.NewClient(ctx, fmt.Sprintf("projects/%s", projectID))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return client
}

func index(w http.ResponseWriter, r *http.Request) {
	defer func() {
		requestCount += 1
	}()
	client := newClient(r.Context())
	defer client.Close()
	lg := client.Logger(logName)

	trace := traceID(r)
	t := fmt.Sprintf("[request #%d] First entry", requestCount)
	lg.Log(logging.Entry{
		Payload:  t,
		Trace:    trace,
		Resource: monRes,
		Severity: logging.Info,
	})
	fmt.Fprintf(w, "Logged: %v\n", t)
	log.Printf("log.Printf Logged: %v\n", t)
	otherFunc()

	t = fmt.Sprintf("[request #%d] A second entry here!", requestCount)
	lg.Log(logging.Entry{
		Payload:  t,
		Trace:    trace,
		Resource: monRes,
		Severity: logging.Warning,
	})
	fmt.Fprintf(w, "Logged: %v\n", t)
	log.Printf("log.Printf Logged: %v\n", t)
}

func nolog(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "No Logged: %v\n")
}

func otherFunc() {
	log.Printf("otherFunc output log")
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/itchyny/gojq"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	ctrl "sigs.k8s.io/controller-runtime"
)

type Service struct {
	Name   string
	Status bool
}

type KService struct {
	Metadata interface{}
	Spec     struct {
		Ports []map[string]interface{}
	}
}

const (
	ANNOTATION_QUERY        = ".metadata.annotations[\"kubestatus/watch\"] == \"true\""
	SERVICES_LIMIT          = 50
	ENDPOINT_ANNOTATION_KEY = "kubestatus/endpoint"
	PORT                    = "8080"
)

func ping(service string, namespace string, endpoint string, protocol string, port float64) bool {
	if endpoint == "" {
		endpoint = "/"
	}
	uri := fmt.Sprint(service, ".", namespace, ":", port, endpoint)

	if protocol == "TCP" {
		conn, err := net.Dial("tcp", uri)
		if err != nil {
			fmt.Println(err)
			return false
		}
		defer conn.Close()
		return true
	} else if protocol == "UDP" {
		add, err := net.ResolveUDPAddr("udp", uri)
		if err != nil {
			fmt.Println(err)
			return false
		}
		conn, err := net.DialUDP("udp", nil, add)
		if err != nil {
			fmt.Println(err)
			return false
		}
		defer conn.Close()
		return true

	}
	return false

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func statusHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	dynamic := dynamic.NewForConfigOrDie(config)

	resourceId := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}
	q := ANNOTATION_QUERY
	resources := make([]Service, 0, SERVICES_LIMIT)

	namespace := os.Getenv("namespace")
	if len(namespace) == 0 {
		namespace = "default"
	}

	query, err := gojq.Parse(q)
	list, err := dynamic.Resource(resourceId).Namespace(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		fmt.Println(err)
	} else {
		for _, item := range list.Items {
			var rawJson map[string]interface{}
			e := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &rawJson)

			if e != nil {
				fmt.Println(e)
			}

			iter := query.Run(rawJson)
			for {
				result, ok := iter.Next()
				if !ok {
					break
				}
				if err, ok := result.(error); ok {
					if err != nil {
						fmt.Println(err)
					}
				} else {
					boolResult, _ := result.(bool)
					if boolResult {
						var kservice KService
						annotations := item.GetAnnotations()
						jsonMarsh, err := item.MarshalJSON()
						if err != nil {
							fmt.Println(err)
						}

						jsonByte := (*json.RawMessage)(&jsonMarsh)

						e := json.Unmarshal(*jsonByte, &kservice)
						if e != nil {
							fmt.Println(err)
						}
						endpoint := annotations[ENDPOINT_ANNOTATION_KEY]
						protocol := kservice.Spec.Ports[0]["protocol"].(string)
						port := kservice.Spec.Ports[0]["port"].(float64)

						status := ping(item.GetName(), item.GetNamespace(), endpoint, protocol, port)

						resources = append(resources, Service{Name: item.GetName(), Status: status})
					}
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)

}

func main() {

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/healthy", healthHandler)
	http.HandleFunc("/ready", readinessHandler)

	log.Println("Starting Server on port " + PORT)
	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		log.Fatal(err)
	}
}

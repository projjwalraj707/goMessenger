package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"crypto/tls"
	"reflect"
)

const certFileName = "certificate.crt"
const keyFileName = "private.key"

type Message struct { //stores each of the messages along with the name of the sender
	Sender string `json: "sender"` //sender of the text message
	Text string `json: "text"` //the text message
}

var dataBase = make(map[string] []Message) //maps recipient's name to a Message slice containing all the messages he has received.

func main() {
	mux := http.NewServeMux()

	//configure TLS
	cfg := &tls.Config {
		MaxVersion: tls.VersionTLS13,
		MinVersion: tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{}, // leave it empty so that kyber is chosen whenever both client and server support it.
	}

	//handle all routes
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("POST /{user}/msg", sendMsg)
	mux.HandleFunc("GET /{user}/retrieveMsg", retrieveMsg)

	srv := &http.Server{
		Addr: ":8080",
		Handler: mux,
		TLSConfig: cfg,
	}
	fmt.Println("Server started on port: ", srv.Addr)
	srv.ListenAndServeTLS(certFileName, keyFileName)
}

func handleRoot( //handles GET request at "/"
	w http.ResponseWriter, //sends response and header
	r *http.Request, //contains the request
) {
	if r.TLS != nil {
		state := r.TLS
		fmt.Fprintf(w, "TLS Version: %x\n", state.Version)
		fmt.Fprintf(w, "Cipher Suite: %x\n", state.CipherSuite)
		fmt.Fprintf(w, "Handshake Complete?: %t\n", state.HandshakeComplete)

		// CurveID (not directly exposed, but inferred from CipherSuite)
		fmt.Fprintf(w, "Server Name: %s\n", state.ServerName)
	} else {
		fmt.Fprintf(w, "Non-TLS connection\n")
	}

	curveID, _ := getRequestCurveID(r)
	curveIDName, _ := getTlsCurveIDName(curveID)
	fmt.Fprintf(w, "Hello World\n")
	fmt.Fprintf(w, "CurveIdName is: %v\n", curveIDName)
}

func getTlsCurveIDName(curveID tls.CurveID) (string, error) {
	curveName := ""
	switch curveID {
	case tls.CurveP256:
		curveName = "P256"
	case tls.CurveP384:
		curveName = "P384"
	case tls.CurveP521:
		curveName = "P521"
	case tls.X25519:
		curveName = "X25519"
	case 0x6399:
		curveName = "X25519Kyber768Draft00"
	default:
		return "", fmt.Errorf("unknown curve ID: 0x%x", uint16(curveID))
	}
	return curveName, nil
}

func getRequestCurveID(r *http.Request) (tls.CurveID, error) {
	if r.TLS == nil {
		return 0, fmt.Errorf("the request is not a TLS connection")
	}

	// Access the private 'testingOnlyCurveID' field using reflection
	connState := reflect.ValueOf(*r.TLS)
	curveIDField := connState.FieldByName("testingOnlyCurveID")

	if !curveIDField.IsValid() {
		return 0, fmt.Errorf("the curve ID field is not found")
	}

	// Convert the reflected value to tls.CurveID
	return tls.CurveID(curveIDField.Uint()), nil
}

func sendMsg(w http.ResponseWriter, r *http.Request, ) { //handles POST request at /sender/msg to send msg
	sender := r.PathValue("user") // current user is the sender of the message
	var jsonData map[string] interface{} //to receive the JSON data from http.Request
	if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil { //check if the request contains any json and then assign it to jsonData
		http.Error(w, "Invalid Json", http.StatusBadRequest)
		return
	}
	if jsonData["recipient"] == "" || jsonData["text"] == "" { //throw error if any field is empty
		http.Error(w, "Incomplete Message", http.StatusBadRequest)
		return
	}
	text, _ := jsonData["text"].(string) //extract the text message
	recipient, _ := jsonData["recipient"].(string) //extract the recipient's name
	msg := Message {
		Sender: sender,
		Text: text,
	}
	dataBase[recipient] = append(dataBase[recipient], msg) // save the message to the data base
	fmt.Fprintf(w, "Your message has successfully been saved.\n") //acknowledge the user that their request is complete
	w.WriteHeader(http.StatusNoContent) //put the status code in the header without any content
}

func retrieveMsg(w http.ResponseWriter, r *http.Request, ) {//shows all the message that the user has received.
	user := r.PathValue("user") //user whose messages are to be retrieved
	w.Header().Set("Content-Type", "application/json")
	msgs, err := json.Marshal(dataBase[user])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError,)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(msgs)
}

package main

import (
	"fmt" // for printing and formatting related functions
	"net/http" // to create http server
	"encoding/json" //to handle JSON format data
	"crypto/tls" // to add tls suppport
	"reflect" // to work with unknown data types
)

const certFileName = "certificate.crt" //certificate file name of SSL certificate
const keyFileName = "private.key" // private key of the certificate

type Message struct { //stores each of the messages along with the name of the sender
	Sender string `json: "sender"` //sender of the text message
	Text string `json: "text"` //the text message
}

var dataBase = make(map[string] []Message) //maps recipient's name to a Message slice containing all the messages he/she has received.

func main() { //main function from where the execution starts
	mux := http.NewServeMux()

	//handle all routes
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("POST /{user}/msg", sendMsg)
	mux.HandleFunc("GET /{user}/retrieveMsg", retrieveMsg)

	srv := &http.Server{ //configure the server
		Addr: ":8080", //port number
		Handler: mux,
		TLSConfig: &tls.Config { //cofigure TLS
			MaxVersion: tls.VersionTLS13,
			MinVersion: tls.VersionTLS13,
			CurvePreferences: []tls.CurveID{}, // leave it empty so that kyber is chosen whenever both client and server support it.
		},
	}
	fmt.Println("Server started on port: ", srv.Addr)
	srv.ListenAndServeTLS(certFileName, keyFileName)
}

func handleRoot(w http.ResponseWriter, r *http.Request,) { // handle the request coming to the root i.e. "localhost:8080/"
	if r.TLS != nil {
		//state := r.TLS

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
	// Access the private 'testingOnlyCurveID' field using reflection
	connectionState := reflect.ValueOf(*r.TLS)
	//connectionState := *r.TLS
	curveIDField := connectionState.FieldByName("testingOnlyCurveID")

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

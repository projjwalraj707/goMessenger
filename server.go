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

func main() { //main function from where the execution starts
	mux := http.NewServeMux()
	srv := &http.Server{ //configure the server
		Addr: ":8080", //port number
		Handler: mux,
		TLSConfig: &tls.Config { //cofigure TLS
			MaxVersion: tls.VersionTLS13,
			MinVersion: tls.VersionTLS13,
			CurvePreferences: []tls.CurveID{
				//in GO 1.23.3 the default CurveID is X25519Kyber768Draft00 but it will be removed in upcoming version of Go (1.24).
				//Also all major browsers are ending support for X25519Kyber768Draft00. When Golang 1.24 is released (most probably in Feb 2025),
				//uncomment the next line MLKEM will automatically be enabled.
				//tls.X25519MLKEM768,
			},
		},
	}

	//handle all routes
	mux.HandleFunc("/", handleRoot) //shows details about the connection
	mux.HandleFunc("POST /{user}/msg", sendMsg) //this route can be used to send a message
	mux.HandleFunc("GET /{user}/retrieveMsg", retrieveMsg) //this route can be used to read all the messages for {user}

	fmt.Println("Server started on port: ", srv.Addr)
	srv.ListenAndServeTLS(certFileName, keyFileName)
}

func handleRoot(w http.ResponseWriter, r *http.Request,) { // handle the request coming to the root i.e. "localhost:8080/"
	if r.TLS != nil {
		//state := r.TLS
		fmt.Fprintf(w, "Hello World\n")
		curveID, _ := getRequestCurveID(r)
		if curveID == 0x6399 {
			fmt.Fprintf(w, "The TLS connection is Quantum Resistant.\n")
		} else {
			fmt.Fprintf(w, "The TLS connection is NOT Quantum Resistant.\n")
		}
		//curveIDName, _ := getTlsCurveIDName(curveID)
		//fmt.Fprintf(w, "CurveIdName is: %v\n", curveIDName)

	} else {
		fmt.Fprintf(w, "Non-TLS connection\n")
	}
}

/* delete this function
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
*/

func getRequestCurveID(r *http.Request) (tls.CurveID, error) {
	// Access the private 'testingOnlyCurveID' field using reflection
	connectionState := reflect.ValueOf(*r.TLS) //ConnectionState struct can be found at https://github.com/golang/go/blob/master/src/crypto/tls/common.go
	curveIDField := connectionState.FieldByName("testingOnlyCurveID")

	if !curveIDField.IsValid() {
		return 0, fmt.Errorf("the curve ID field is not found")
	}

	// Convert the reflected value to tls.CurveID
	return tls.CurveID(curveIDField.Uint()), nil
}







//Extra functionalities
type Message struct { //stores each of the messages along with the name of the sender
	Sender string `json: "sender"` //sender of the text message
	Text string `json: "text"` //the text message
}

var dataBase = make(map[string] []Message) //maps recipient's name to a Message slice containing all the messages he/she has received.

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

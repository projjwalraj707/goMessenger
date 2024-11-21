package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)

type Message struct {
	Author string `json: "author"` //whoever has written the text
	Text string `json: "text"` //the text message
}

var dataBase = make(map[string] []Message) //maps recipient's name to a Message slice containing all the messages he has received.

func handleRoot( //handles GET request at "/"
	w http.ResponseWriter, //sends response and header
	r *http.Request, //contains the request
) {
	fmt.Fprintf(w, "Hello World\n")
}

func sendMsg(w http.ResponseWriter, r *http.Request, ) { //handles POST request at /author/msg to send msg
	author := r.PathValue("user") // current user is the author of the message
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
		Author: author,
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

/*
func retrieveMsg(w http.ResponseWriter, r *http.Request, ) {
	author := PathValue("author")
	msg, ok := dataBase[author]
	if !ok {
		http.Error(w, "No message found for this user.")
t	
}
*/

func main() {
	mux := http.NewServeMux()

	//handle all routes
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("POST /{user}/msg", sendMsg)
	mux.HandleFunc("GET /{user}/retrieveMsg", retrieveMsg)

	fmt.Println("Server started on port: 8080")
	http.ListenAndServe(":8080", mux)
}

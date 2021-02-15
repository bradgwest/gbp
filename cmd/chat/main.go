package main

import (
	"log"
	"net/http"
)

var content = []byte(`
<html>
	<head>
		<title>Chat</title>
	</head>
	<body>
		Let's Chat!
	</body>
</html>
`)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	})

	// start the webserver
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

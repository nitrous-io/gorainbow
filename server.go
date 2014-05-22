package gorainbow

import (
	"encoding/json"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/nitrous-io/gorainbow/sockjs"
)

type (
	HttpHandler func(w http.ResponseWriter, r *http.Request) error

	config struct {
		Address string `json:"address"`
	}

	Server struct {
		Config *config
		Router *mux.Router
	}
)

var homeTemplStr = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>Chat Example</title>
<script src="http://cdn.sockjs.org/sockjs-0.3.min.js"></script>
<script type="text/javascript">
	var sock = new SockJS('http://128.199.236.54:10000/ws', null, {
		debug: true,
		devel: true,
		protocols_whitelist: ["websocket"]
	});
	sock.onopen = function() {
		console.log('open');
		setInterval(function() {
			sock.send("hello world");
		}, 10000);
	};
	sock.onmessage = function(e) {
		console.log('message', e.data);
	};
	sock.onclose = function() {
		console.log('close');
	};
</script>
</head>
<body>
</body>
</html>
`

var homeTempl = template.Must(template.New("home").Parse(homeTemplStr))

func (s *Server) getHome(w http.ResponseWriter, r *http.Request) error {
	homeTempl.Execute(w, nil)
	return nil
}

func (s *Server) getDownload(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) optionsUpload(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) postUpload(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func wsHandler(msg []byte) []interface{} {
	return []interface{}{
		"hello",
	}
}

func (s *Server) CreateRouter() {
	r := mux.NewRouter()
	// setup http handler
	m := map[string]map[string]HttpHandler{
		"GET": {
			"/":         s.getHome,
			"/download": s.getDownload,
		},
		"OPTIONS": {
			"/upload": s.optionsUpload,
		},
		"POST": {
			"/upload": s.postUpload,
		},
	}
	for method, routes := range m {
		for route, handler := range routes {
			localRoute := route
			localHandler := handler
			localMethod := method
			f := makeHttpHandler(localHandler)
			r.HandleFunc(localRoute, f).Methods(localMethod)
		}
	}
	// setup websocket handler
	wsRouter := r.PathPrefix("/ws").Subrouter()
	sockjs.New(wsRouter, wsHandler, 25*time.Second)

	// assigned the router to `Router` instance
	s.Router = r
}

func New(configData []byte) (*Server, error) {
	config := &config{}
	if err := json.Unmarshal(configData, config); err != nil {
		log.Printf("gorainbow: invalid config data %v", string(configData))
		return nil, err
	}
	server := &Server{Config: config}
	server.CreateRouter()
	return server, nil
}

func makeHttpHandler(handlerFunc HttpHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handlerFunc(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

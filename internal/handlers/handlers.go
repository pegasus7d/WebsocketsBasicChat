package handlers

import (
	
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan=make(chan WsPayload)

var clients=make(map[WebSocketConnection]string)

var views=jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection=websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {return true},

}

type WebSocketConnection struct{
	*websocket.Conn
}
type WsJsonResponse struct {
	Action      string `json:"action"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayload struct{
	Action      string `json:"action"`
	Username      string `json:"username"`
	Message     string `json:"message"`
	Conn        WebSocketConnection `json:"-"`
}

func Home(w http.ResponseWriter,r *http.Request) {

	err:=renderPage(w,"home.jet",nil)
	if err!=nil{
		log.Println(err)
	}





}

func WsEndpoint(w http.ResponseWriter,r *http.Request){
	ws,err:=upgradeConnection.Upgrade(w,r,nil)
	if err!=nil{
		log.Println(err)
	}
	log.Println("Client connected to Endpoint")
	// defer ws.Close()
	var response WsJsonResponse
	response.Message =`<em><small>Connected to server </small></em>`



	conn:=WebSocketConnection{Conn: ws}
	clients[conn]=""

	
	err=ws.WriteJSON(response)
	if err!=nil{
		log.Println(err)
	}

	go ListenForWs(&conn)




}


func ListenForWs(conn *WebSocketConnection){
	 defer func(){
		if r:=recover();r!=nil{
			log.Println(r)
		}
	 }()


	 var payload WsPayload

	 for {
		err:=conn.ReadJSON(&payload)
		if err!=nil{

		}else{
			payload.Conn=*conn
			wsChan<-payload
		}
	 }
}

func ListenToWsChannel(){
	var response WsJsonResponse
	
	for{
		e:=<-wsChan 


		switch e.Action{
		case"username":
			clients[e.Conn]=e.Username
			users:=getUserList()
			response.Action="List_users"
			response.ConnectedUsers=users
			BroadCastToAll(response)

		}
		// response.Action="Got Here"
		// response.Message=fmt.Sprintf("Some message and action was %s", e.Action)
		// BroadCastToAll(response)

	}
}

func getUserList()[]string{
	var userList []string
	for _,x:=range clients{
		userList=append(userList,x)
	}
	sort.Strings(userList)
	return userList
}

func BroadCastToAll(response WsJsonResponse){
	for client:=range clients{
		err:=client.WriteJSON(response)
		if err!=nil{
			log.Println("websocket error")
			_=client.Close()
			delete(clients,client)
		}
	}
}


func renderPage(w http.ResponseWriter,tmpl string,data jet.VarMap)error{
	view,err:=views.GetTemplate(tmpl)
	if err!=nil{
		log.Println(err)
		return err
	}

	err=view.Execute(w,data,nil)

	if err!=nil{
		log.Println(err)
		return err
	}
	return nil



}
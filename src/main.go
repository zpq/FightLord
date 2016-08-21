package main

import (
	"encoding/json"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var (
	addr       string = ":8008"
	rooms      map[int]*Room
	users      map[int]*User // key=> user id
	store      = sessions.NewCookieStore([]byte("something-very-secret"))
	cookieName = "lordCookie"
	upgrader   = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func Register(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	isCanRegister := true
	for _, v := range users {
		if username == v.Name {
			isCanRegister = false
			break
		}
	}
	res := HttpRes{
		Status:  0,
		Message: "user already exists",
	}
	if isCanRegister {
		id := len(users) + 1
		users[id] = &User{
			Id:       id,
			Name:     username,
			password: password,
			CardInfo: make(map[int]*Card, 20),
		}
		res.Status = 1
		res.Message = "register success!"
	}
	body, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err.Error())
	}
	w.Write([]byte(body))
}

func Login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	login := false
	userid := 0
	for _, v := range users {
		if username == v.Name && password == v.password {
			login = true
			userid = v.Id
			break
		}
	}
	res := HttpRes{Status: 0, Message: "login failed, please check your post"}
	if login {
		session, err := store.Get(r, cookieName)
		if err != nil {
			log.Fatal(err.Error())
		}
		session.Values["userid"] = userid
		session.Values["username"] = username
		session.Save(r, w)
		res.Status = 1
		res.Message = "login success!"
		res.Datas = append(res.Datas, users[userid])
	}
	body, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err.Error())
	}
	w.Write([]byte(body))
}

//进入房间后马上调用(用户和房间的绑定操作)
func EnterRoom(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, cookieName)
	if err != nil {
		log.Fatal(err.Error())
	}
	// arrive here, means auth success
	r.ParseForm()
	userid := session.Values["userid"].(int)
	user := users[userid]
	roomId, err := strconv.Atoi(r.PostFormValue("roomId"))
	res := HttpRes{
		Status:  0,
		Message: "invalid roomId",
	}
	if err != nil {

	} else {
		_, ok := rooms[roomId]
		if ok {
			if len(rooms[roomId].Battle.Members) == 3 { //房间满了
				res.Message = "sorry!this room is full"
			} else {
				res.Status = 1
				res.Message = "welcome " + user.Name + " enter the room " + strconv.Itoa(roomId)
				user.RoomId = roomId
			}
		} else {
			res.Message = "unknown room"
		}
	}

	body, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err.Error())
	}
	w.Write([]byte(body))
}

func Ws(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	session, err := store.Get(r, cookieName)
	if err != nil {
		// conn.WriteJSON("未登录用户")
		log.Fatal(err.Error())
	}
	// arrive here, means auth success
	userid := session.Values["userid"].(int)
	user := users[userid]
	//判断这个用户之前是否处于一个房间中，（是否断线重连）
	_, ok := rooms[user.RoomId]
	var c *Connection
	if ok { // reconnect
		c = rooms[user.RoomId].Battle.Members[userid]
		c.WS = conn
	} else { // new
		c = &Connection{
			User: user,
			WS:   conn,
			Send: make(chan []byte, 1024),
		}
	}

	rooms[user.RoomId].register <- c
	defer func() {
		c.WS = nil
	}()
	go c.write()
	go c.read()
}

func (c *Connection) read() {
	for {
		_, msg, err := c.WS.ReadMessage()
		if err != nil {
			break
		}
		boardCast := &BoardCast{
			userid:  c.User.Id,
			message: msg,
		}
		rooms[c.User.RoomId].boardCast <- boardCast
	}
	c.WS.Close()
}

func (c *Connection) write() {
	for msg := range c.Send {
		err := c.WS.WriteJSON(msg)
		if err != nil {
			break
		}
	}
}

func (room *Room) roomHandle() {
	for {
		select {
		case register := <-room.register:
			room.Battle.Members[register.User.Id] = register
			room.Battle.Position[len(room.Battle.Members)] = register.User.Id
		case runegister := <-room.unregister:
			log.Println(runegister)
		case boardCast := <-room.boardCast:
			msg := coreLogic(room, boardCast)
			for _, v := range room.Battle.Members {
				v.Send <- msg
			}
		}
	}
}

func coreLogic(room *Room, boardCast *BoardCast) []byte {
	/**
	 * 1. 是否是该用户的轮次
	 * 2. 牌是否合法，属于该用户（card id）
	 * 3. 出牌是否符合规则
	 *
	 */
	result := ""
	return []byte(result)
}

func newRoom(n int) {
	index := len(rooms)
	for i := index; i < n; i++ {
		room := &Room{
			Id:         i,
			Status:     0,
			register:   nil,
			unregister: nil,
			Battle:     &Battle{},
		}
		rooms[i] = room
		go room.roomHandle()
	}
}

func Rooms(w http.ResponseWriter, r *http.Request) {
	res := HttpRes{
		Status:  1,
		Message: "success",
	}
	for _, v := range rooms {
		res.Datas = append(res.Datas, v)
	}
	body, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err.Error())
	}
	w.Write([]byte(body))
}

func Home(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("../resource/html/index.html")
	if err != nil {
		log.Fatal(err.Error())
	}
	t.Execute(w, nil)
}

func main() {
	http.HandleFunc("/register", Register)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/ws", Ws)
	http.HandleFunc("rooms", Rooms)
	http.HandleFunc("/", Home)
	http.ListenAndServe(addr, nil)
}

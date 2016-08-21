package main

import (
	"github.com/gorilla/websocket"
)

type Room struct {
	Id         int
	Status     int
	register   chan *Connection
	unregister chan *Connection
	boardCast  chan *BoardCast
	Battle     *Battle
}

type BoardCast struct {
	userid  int
	message []byte
}

type Battle struct {
	Id       int
	Status   int
	Members  map[int]*Connection // key => user id
	Guests   map[int]*Connection
	Position map[int]int // value => user id
	Current  int         // current user   1-2-3-1-2-3-1-2-3
	Last     int         // last user
}

type Connection struct {
	User *User
	WS   *websocket.Conn
	Send chan []byte
}

type User struct {
	Id       int
	Name     string
	password string
	Role     int // is lord or farmer
	Ready    int
	RoomId   int
	CardInfo map[int]*Card
}

type Card struct {
	Id     int
	Flower int // 1,2  3,4,5,6
	Points int
}

type HttpRes struct {
	Status  int           `json:"status"`
	Message string        `json:"message"`
	Datas   []interface{} `json:"datas"`
}

package main

import (
	"encoding/gob"
	. "fmt"
	"net"
	"os"
	"sync"
	"time"
)

type TCP struct {
	ID, Cur int
}

type Proceso struct {
	id int
	i  chan int
}

type Clientes struct {
	m map[uint]bool
	sync.RWMutex
}

type Data struct {
	cur  []Proceso
	mine Clientes
}

func (p *Proceso) run() {
	for {
		i, mine := <-p.i
		if !mine {
			return
		}
		Printf("Proceso %d -> %d\n", p.id, i)
		time.Sleep(500 * time.Millisecond)
		p.i <- i + 1
	}
}

func getClientID(c net.Conn) uint {
	var name uint
	gob.NewDecoder(c).Decode(&name)
	return name
}

func (db *Data) hiClient(c net.Conn) {
	name := getClientID(c)
	cs := TCP{}

	db.mine.Lock()
	_, ex := db.mine.m[name]
	if ex {
		gob.NewDecoder(c).Decode(&cs)
		delete(db.mine.m, name)
		db.cur = append(db.cur, Proceso{cs.ID, make(chan int, 1)})
		db.cur[len(db.cur)-1].i <- cs.Cur + 1
		go db.cur[len(db.cur)-1].run()
	} else {
		db.mine.m[name] = true
		id := db.cur[0].id
		cur := <-db.cur[0].i
		close(db.cur[0].i)
		db.cur = db.cur[1:]
		cs = TCP{id, cur}
		aux := gob.NewEncoder(c).Encode(cs)
		if aux != nil {
			Println(aux)
		}
	}
	db.mine.Unlock()
}

func server() {
	s, err := net.Listen("tcp", ":9999")
	if err != nil {
		Println(err)
		os.Exit(0)
	}
	var u Data
	u.cur = make([]Proceso, 5)
	u.mine.m = make(map[uint]bool)
	for i := range u.cur {
		u.cur[i] = Proceso{i, make(chan int, 1)}
		u.cur[i].i <- 0
		go u.cur[i].run()
	}
	for {
		c, err := s.Accept()
		if err != nil {
			Println(err)
			continue
		}
		go u.hiClient(c)
	}
}

func main() {
	go server()
	var meh string
	Scan(&meh)
}

package main

import (
	"encoding/gob"
	. "fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type TCP struct {
	ID, Cur int
}

type Proceso struct {
	id int
	i  chan int
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

func hiServer(C uint) net.Conn {
	c, err := net.Dial("tcp", ":9999")
	if err != nil {
		Println(err)
		os.Exit(0)
	}
	aux := gob.NewEncoder(c).Encode(C)
	if aux != nil {
		Println(aux)
	}
	return c
}

func getP(c net.Conn) TCP {
	var p TCP
	gob.NewDecoder(c).Decode(&p)
	c.Close()
	return p
}

func giveP(c net.Conn, p *TCP) {
	aux := gob.NewEncoder(c).Encode(p)
	if aux != nil {
		Println(aux)
	}
	c.Close()
}

func getID() uint {
	var id uint
	id = uint(time.Now().Nanosecond())
	id ^= id << 13
	id ^= id >> 17
	id ^= id << 5
	return id
}

func main() {
	ClientID := getID()
	c := hiServer(ClientID)
	p := getP(c)
	mine := Proceso{p.ID, make(chan int, 1)}
	mine.i <- p.Cur
	go mine.run()
	ch := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		p.Cur = <-mine.i
		close(mine.i)
		c = hiServer(ClientID)
		giveP(c, &p)
		done <- true
	}()
	<-done
}

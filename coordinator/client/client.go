package client

import (
	"context"
	"fmt"
	"log"

	"github.com/khoakmp/brgame/api"
	"github.com/khoakmp/brgame/coordinator/network"
)

type MessageHandler interface {
	HandleClientMessage(msg *api.Message, c *Client)
}

type Client struct {
	id             string
	conn           network.Conn
	handler        MessageHandler
	outMessageChan chan *api.Message
	context        context.Context
	cancelFn       context.CancelFunc
	hub            *Hub
}

func (c *Client) ID() string {
	return c.id
}
func (c *Client) SendMessage(msg *api.Message) {
	c.outMessageChan <- msg
}

func (c *Client) readLoop() {
	for {
		msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("[Client %s] failed to read message %s\n", c.id, err)
			break
		}
		c.handler.HandleClientMessage(&msg, c)
	}
	c.close()
}

func (c *Client) Alive() bool {
	return c.context.Err() == nil
}
func (c *Client) close() {
	if err := c.conn.Close(); err != nil {
		log.Printf("[Client %s] close failed, %s\n", c.id, err)
	}
	c.cancelFn()
	c.hub.RemoveClient(c.id)
}

func (c *Client) WaitTimeout() {
	c.SendMessage(&api.Message{
		SessionID:   "",
		SenderID:    "coordinator_",
		ReceiverIDs: nil,
		Type:        api.MessageWaitTimeout,
		Payload:     "",
	})
}

func (c *Client) writeLoop() {
	for {
		select {
		case <-c.context.Done():
			fmt.Printf("[Client %s] stop write loop, caused by context Done\n", c.id)
			return
		case msg := <-c.outMessageChan:

			if err := c.conn.WriteMessage(msg); err != nil {
				log.Printf("[Client %s] failed to write message\n", c.id)
				return
			}
			// TODO:
			log.Printf("[Client %s] Write message type %s\n", c.id, msg.Type)

		}
	}
}

func New(id string, conn network.Conn, handler MessageHandler, hub *Hub) *Client {
	context, cancelFn := context.WithCancel(context.Background())
	c := &Client{
		id:             id,
		conn:           conn,
		handler:        handler,
		outMessageChan: make(chan *api.Message),
		context:        context,
		cancelFn:       cancelFn,
		hub:            hub,
	}
	go c.readLoop()
	go c.writeLoop()
	return c
}

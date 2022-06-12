package chaos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
)

type NodeConfig struct {
	Addr          string
	BootstrapAddr string
}

func RunNode(ctx context.Context, config NodeConfig) error {
	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return err
	}

	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		log.SetPrefix(fmt.Sprintf("NODE[%d]: ", tcpAddr.Port))
	}

	log.Print("Running node, sending introduction")
	service := NewNodeService()
	go func() {
		defer listener.Close()
		<-ctx.Done()
		log.Println("Shutting down")
	}()

	if err := service.JoinNetwork(config.BootstrapAddr, listener.Addr().String()); err != nil {
		log.Printf("failed to join network: %s", err)
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go service.Serve(ctx, conn)
	}
}

type NodeService struct {
	Neighbors []string
}

func NewNodeService() *NodeService {
	return &NodeService{}
}

func (s *NodeService) Serve(ctx context.Context, conn net.Conn) {
	if err := s.handle(ctx, conn); err != nil {
		log.Printf("bootstrap serve error: %s", err)
	}
	if err := conn.Close(); err != nil {
		log.Printf("Close client error: %s", err)
	}
}

// handle will process all incoming messages
func (s *NodeService) handle(ctx context.Context, conn net.Conn) error {
	var msg Message
	decoder := json.NewDecoder(conn)

	if err := decoder.Decode(&msg); err != nil {
		return err
	}
	message, err := DecodeMessage(msg)

	if err != nil {
		return err
	}
	switch params := message.(type) {
	case MessageNodeIntroduction:
		response, err := s.NodeIntroduction(ctx, params)
		if err != nil {
			return err
		}
		return Response(conn, response)
	default:
		return fmt.Errorf("unhandled message type: %T", msg)
	}
}

// NodeIntroduction will receive MessageNodeIntroduction with new node address and add that node to the neighbors.
func (s *NodeService) NodeIntroduction(_ context.Context, params MessageNodeIntroduction) (MessageOkResponse, error) {
	log.Printf("node %s introduced", params.Addr)
	s.Neighbors = append(s.Neighbors, params.Addr)
	return MessageOkResponse{
		OK: true,
	}, nil
}

// JoinNetwork adds node to network. It will first call bootstrap to get address of node
//it needs to connect to and after that it will introduce to the node and add it as neighbor.
func (s *NodeService) JoinNetwork(bAddr, addr string) error {
	msg := MessageNodeIntroduction{Addr: addr}
	bootstrapClient := NewClient(bAddr)
	resp, err := bootstrapClient.IntroduceToBootstrap(msg)
	if err != nil {
		return fmt.Errorf("failed bootstrap introduction: %s", err)
	}
	log.Printf("introduced to bootstrap: %q %t", resp.Addr, resp.First)

	// We will skip node introduction if it is first node added to network
	if resp.First {
		return nil
	}
	s.Neighbors = append(s.Neighbors, resp.Addr)

	nodeClient := NewClient(resp.Addr)
	if err := nodeClient.IntroduceToNode(msg); err != nil {
		return err
	}
	log.Printf("introduced to node: %q", resp.Addr)
	return nil
}

// Client is used for communicating with other nodes and bootstrap.
type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{addr: addr}
}

// IntroduceToBootstrap will introduce itself to bootstrap and receive address of node in network.
func (c *Client) IntroduceToBootstrap(message MessageNodeIntroduction) (MessageNodeIntroductionResponse, error) {
	var resp MessageNodeIntroductionResponse
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return resp, err
	}
	data, err := EncodeMessage(message)
	if err != nil {
		return resp, err
	}
	if _, err = io.Copy(conn, bytes.NewReader(data)); err != nil {
		return resp, err
	}
	return resp, json.NewDecoder(conn).Decode(&resp)
}

// IntroduceToNode will introduce itself to bootstrap and receive address of node in network.
func (c *Client) IntroduceToNode(message MessageNodeIntroduction) error {
	var resp MessageOkResponse
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	data, err := EncodeMessage(message)
	if err != nil {
		return err
	}
	if _, err = io.Copy(conn, bytes.NewReader(data)); err != nil {
		return err
	}
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}
	if resp.OK == false {
		return fmt.Errorf("introduce to node failed")
	}
	return nil
}

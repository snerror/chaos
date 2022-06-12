package chaos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type BootstrapConfig struct {
	Addr string
}

func RunBootstrap(ctx context.Context, config BootstrapConfig) error {
	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return err
	}
	log.Printf("Bootstrap server running on %s", listener.Addr())

	service := NewBootstrapService()

	go func() {
		defer listener.Close()
		<-ctx.Done()
		log.Println("Shutting down")
	}()

	defer func() { _ = listener.Close() }()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go service.Serve(ctx, conn)
	}
}

type BootstrapService struct {
	Nodes []string
}

func NewBootstrapService() *BootstrapService {
	return &BootstrapService{}
}

func (s *BootstrapService) Serve(ctx context.Context, conn net.Conn) {
	if err := s.handle(ctx, conn); err != nil {
		log.Printf("Bootstrap serve error: %s", err)
	}
	if err := conn.Close(); err != nil {
		log.Printf("Close client error: %s", err)
	}
}

func (s *BootstrapService) handle(ctx context.Context, conn net.Conn) error {
	decoder := json.NewDecoder(conn)
	var msg Message
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

func (s *BootstrapService) NodeIntroduction(_ context.Context, params MessageNodeIntroduction) (MessageNodeIntroductionResponse, error) {
	log.Printf("Added node %s to list", params.Addr)
	var parent string
	if len(s.Nodes) != 0 {
		parent = s.Nodes[len(s.Nodes)-1]
	}

	s.Nodes = append(s.Nodes, params.Addr)
	return MessageNodeIntroductionResponse{
		Addr:  parent,
		First: parent == "",
		Nodes: s.Nodes,
	}, nil
}

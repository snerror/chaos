package chaos

import (
	"bufio"
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const (
	CommandBootstrapStart string = "bootstrap"
	CommandNodeStart      string = "node"
)

type CliConfig struct {
	BootstrapPath string
	NodePath      string
	BootstrapAddr string
}

// RunCli executes user input commands from os.Stdin, and shuts
// down on "quit" command.
func RunCli(ctx context.Context, config CliConfig) error {
	buf := bufio.NewScanner(os.Stdin)
	service := NewCliService(config.BootstrapAddr, config.BootstrapPath, config.NodePath)
	commands := map[string]func(context.Context, []string){
		CommandNodeStart:      service.StartNode,
		CommandBootstrapStart: service.StartBootstrap,
	}
	log.Print("Waiting for commands:")
	for buf.Scan() {
		args := strings.Fields(buf.Text())
		if len(args) == 0 {
			continue
		}
		command := args[0]
		handler, ok := commands[command]
		if !ok {
			log.Printf("Unknown command: %q", command)
			continue
		}
		handler(ctx, args[1:])
	}
	return buf.Err()
}

type CliService struct {
	bootstrapAddr string
	bootstrapPath string
	nodePath      string
	bootstrapMu   sync.Mutex
	bootstrapOn   bool
}

func NewCliService(bootstrapAddr, bootstrapPath, nodePath string) *CliService {
	return &CliService{
		bootstrapAddr: bootstrapAddr,
		bootstrapPath: bootstrapPath,
		nodePath:      nodePath,
	}
}

func (s *CliService) StartNode(ctx context.Context, args []string) {
	cmd := exec.Command(s.nodePath)
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("Start node error: %s", err)
	}
}

func (s *CliService) StartBootstrap(ctx context.Context, args []string) {
	defer s.bootstrapMu.Unlock()
	s.bootstrapMu.Lock()

	if s.bootstrapOn {
		log.Print("Bootstrap already running")
		return
	}
	s.bootstrapOn = true

	log.Print("Starting bootstrap")
	cmd := exec.Command(s.bootstrapPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("Start bootstrap error: %s", err)
	}
}

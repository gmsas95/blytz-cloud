package provisioner

import (
	"context"
	"fmt"
)

type PortAllocator struct {
	startPort int
	endPort   int
	allocated map[int]bool
}

func NewPortAllocator(startPort, endPort int) *PortAllocator {
	return &PortAllocator{
		startPort: startPort,
		endPort:   endPort,
		allocated: make(map[int]bool),
	}
}

func (pa *PortAllocator) LoadAllocatedPorts(ctx context.Context, db interface {
	GetAllocatedPorts(context.Context) ([]int, error)
}) error {
	ports, err := db.GetAllocatedPorts(ctx)
	if err != nil {
		return fmt.Errorf("load allocated ports: %w", err)
	}

	for _, port := range ports {
		pa.allocated[port] = true
	}

	return nil
}

func (pa *PortAllocator) AllocatePort() (int, error) {
	for port := pa.startPort; port <= pa.endPort; port++ {
		if !pa.allocated[port] {
			pa.allocated[port] = true
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", pa.startPort, pa.endPort)
}

func (pa *PortAllocator) ReleasePort(port int) {
	delete(pa.allocated, port)
}

package tcpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
)

func (s *Server) Start(ctx context.Context) error {
	var err error

	s.listener, err = net.Listen("tcp", ":"+s.cfg.TCPPort)
	if err != nil {
		return fmt.Errorf("TCP Server Start net.Listen(...): %w", err)
	}

	defer func() {
		if err = s.listener.Close(); err != nil {
			s.log.Errorf("TCP Server Start s.listener.Close(): %v", err)
		}
	}()

	s.log.Info("TCP Server listening on port ", s.cfg.TCPPort)

	connChan := make(chan net.Conn)
	errChan := make(chan error)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				errChan <- err

				return
			}
			connChan <- conn
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.log.Info("TCP server shutdown initiated.")

				return
			case err := <-errChan:
				if !errors.Is(err, net.ErrClosed) {
					s.log.WithError(err).Error("Error accepting connection")
				}

				continue
			case conn := <-connChan:
				go func() {
					if err = s.handleConnection(ctx, conn); err != nil {
						errChan <- fmt.Errorf("TCP Server Start s.handleConnection(...): %w", err)
					}
				}()
			}
		}
	}()

	<-ctx.Done()

	return nil
}

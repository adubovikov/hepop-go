package server

import (
	"net"
	"sync"

	"github.com/sipcapture/hepop-go/internal/writer"
	"github.com/sipcapture/hepop-go/pkg/protocol"
	"github.com/sirupsen/logrus"
)

type HEPServer struct {
	config      *Config
	writer      writer.Writer
	udpConn     *net.UDPConn
	tcpListener net.Listener
	wg          sync.WaitGroup
	done        chan struct{}
}

type Config struct {
	Host string
	Port int
}

func NewHEPServer(config *Config, writer writer.Writer) *HEPServer {
	return &HEPServer{
		config: config,
		writer: writer,
		done:   make(chan struct{}),
	}
}

func (s *HEPServer) Start() error {
	// Start UDP server
	udpAddr := net.UDPAddr{
		Port: s.config.Port,
		IP:   net.ParseIP(s.config.Host),
	}

	conn, err := net.ListenUDP("udp", &udpAddr)
	if err != nil {
		return err
	}
	s.udpConn = conn

	// Start TCP server
	tcpAddr := net.TCPAddr{
		Port: s.config.Port,
		IP:   net.ParseIP(s.config.Host),
	}

	listener, err := net.ListenTCP("tcp", &tcpAddr)
	if err != nil {
		return err
	}
	s.tcpListener = listener

	// Start handlers
	s.wg.Add(2)
	go s.handleUDP()
	go s.handleTCP()

	logrus.Infof("HEP server started on %s:%d", s.config.Host, s.config.Port)
	return nil
}

func (s *HEPServer) Stop() error {
	close(s.done)
	if s.udpConn != nil {
		s.udpConn.Close()
	}
	if s.tcpListener != nil {
		s.tcpListener.Close()
	}
	s.wg.Wait()
	return nil
}

func (s *HEPServer) handleUDP() {
	defer s.wg.Done()
	buffer := make([]byte, 65535)

	for {
		select {
		case <-s.done:
			return
		default:
			n, addr, err := s.udpConn.ReadFromUDP(buffer)
			if err != nil {
				logrus.Error("UDP read error:", err)
				continue
			}

			packet := make([]byte, n)
			copy(packet, buffer[:n])
			go s.processHEP(packet, addr.String())
		}
	}
}

func (s *HEPServer) handleTCP() {
	defer s.wg.Done()

	for {
		select {
		case <-s.done:
			return
		default:
			conn, err := s.tcpListener.Accept()
			if err != nil {
				logrus.Error("TCP accept error:", err)
				continue
			}
			go s.handleTCPConnection(conn)
		}
	}
}

func (s *HEPServer) handleTCPConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 65535)

	for {
		select {
		case <-s.done:
			return
		default:
			n, err := conn.Read(buffer)
			if err != nil {
				logrus.Error("TCP read error:", err)
				return
			}

			packet := make([]byte, n)
			copy(packet, buffer[:n])
			go s.processHEP(packet, conn.RemoteAddr().String())
		}
	}
}

func (s *HEPServer) processHEP(packet []byte, addr string) {
	hep, err := protocol.DecodeHEP(packet)
	if err != nil {
		logrus.Error("HEP decode error:", err)
		return
	}

	if err := s.writer.Write(hep); err != nil {
		logrus.Error("Writer error:", err)
	}
}

package device

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/MJKhaani/ESP32-Marauder/internal/capabilities"
	"github.com/MJKhaani/ESP32-Marauder/internal/marauder"
	"github.com/MJKhaani/ESP32-Marauder/internal/serialtransport"
)

var (
	ErrSessionClosed   = errors.New("device session is closed")
	ErrCommandTimeout = errors.New("command timeout")
)

// SessionListener handles events produced during an active device session.
type SessionListener func(ev *marauder.ParsedEvent)

// Session owns a single serial port connection and command dispatcher.
type Session struct {
	Device   *Device
	Port     serialtransport.SerialPort
	Parser   *marauder.WireParser
	Commands *marauder.CommandBuilder

	mu          sync.Mutex
	closed      bool
	listenersMu sync.RWMutex
	listeners   []SessionListener

	ctx    context.Context
	cancel context.CancelFunc
}

// NewSession creates and starts a new serial session.
func NewSession(d *Device, port serialtransport.SerialPort) *Session {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Session{
		Device:    d,
		Port:      port,
		Parser:    marauder.NewWireParser(),
		Commands:  marauder.NewCommandBuilder(),
		listeners: make([]SessionListener, 0),
		ctx:       ctx,
		cancel:    cancel,
	}

	go s.readLoop()
	return s
}

// AddListener registers an event callback for parsed wire messages.
func (s *Session) AddListener(l SessionListener) {
	s.listenersMu.Lock()
	defer s.listenersMu.Unlock()
	s.listeners = append(s.listeners, l)
}

func (s *Session) emit(ev *marauder.ParsedEvent) {
	if ev == nil {
		return
	}
	s.listenersMu.RLock()
	listeners := append([]SessionListener(nil), s.listeners...)
	s.listenersMu.RUnlock()

	for _, l := range listeners {
		l(ev)
	}
}

// WriteCommand sends a raw command string to the device.
func (s *Session) WriteCommand(cmd string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrSessionClosed
	}

	_, err := s.Port.Write([]byte(cmd))
	return err
}

// InitializeInterrogation sends `info` and `help` to fingerprint the connected board.
func (s *Session) InitializeInterrogation(evaluator *capabilities.Evaluator) error {
	profile := &marauder.FirmwareProfile{
		SupportedCommands: make(map[string]bool),
	}

	infoChan := make(chan marauder.InfoEventData, 1)
	helpChan := make(chan marauder.HelpEventData, 1)

	// Temporary listener to catch initial info and help
	s.AddListener(func(ev *marauder.ParsedEvent) {
		if ev.Type == marauder.EventInfo {
			if data, ok := ev.Data.(marauder.InfoEventData); ok {
				select {
				case infoChan <- data:
				default:
				}
			}
		} else if ev.Type == marauder.EventHelp {
			if data, ok := ev.Data.(marauder.HelpEventData); ok {
				select {
				case helpChan <- data:
				default:
				}
			}
		}
	})

	// Send info
	_ = s.WriteCommand(s.Commands.Info())

	// Wait up to 1.5 seconds for info response
	select {
	case info := <-infoChan:
		profile.RawVersion = info.Version
		profile.Major = info.Major
		profile.Minor = info.Minor
		profile.Patch = info.Patch
		profile.HardwareName = info.Hardware
		profile.HasSD = info.HasSD
		profile.HasGPS = info.HasGPS
		profile.BatteryStatus = info.BatteryStatus
		s.Device.Identity.FirmwareVer = info.Version
		s.Device.Identity.BoardModel = info.Hardware
		s.Device.Identity.Confidence = "verified"
	case <-time.After(1500 * time.Millisecond):
		s.Device.Identity.Confidence = "presumed"
	}

	// Send help
	_ = s.WriteCommand(s.Commands.Help())

	// Wait up to 1.5 seconds for help response
	select {
	case help := <-helpChan:
		for _, cmd := range help.Commands {
			profile.SupportedCommands[cmd] = true
		}
	case <-time.After(1500 * time.Millisecond):
		// Fallback defaults if help timed out
		if profile.IsModern() {
			profile.SupportedCommands["scanall"] = true
			profile.SupportedCommands["sniffbeacon"] = true
			profile.SupportedCommands["sniffdeauth"] = true
		} else {
			profile.SupportedCommands["scanap"] = true
			profile.SupportedCommands["scansta"] = true
		}
	}

	s.Device.Profile = profile
	s.Device.Capabilities = evaluator.Evaluate(s.Device.ID, profile)
	return nil
}

func (s *Session) readLoop() {
	scanner := bufio.NewScanner(s.Port)
	// Bounded scanner buffer up to 64KB per line
	buf := make([]byte, 4096)
	scanner.Buffer(buf, 65536)

	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			return
		default:
			line := scanner.Text()
			ev := s.Parser.ParseLine(line)
			if ev != nil {
				s.emit(ev)
			}
		}
	}

	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		s.emit(&marauder.ParsedEvent{
			Type:      marauder.EventStatus,
			Timestamp: time.Now(),
			Data:      fmt.Sprintf("Serial read error: %v", err),
			RawLine:   "",
		})
	}
}

// Close terminates the session and closes the serial port.
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true
	s.cancel()

	return s.Port.Close()
}

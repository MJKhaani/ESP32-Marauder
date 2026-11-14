package marauder

import (
	"fmt"
	"strings"
)

// CommandBuilder constructs validated CLI command payloads.
type CommandBuilder struct{}

// NewCommandBuilder returns a new CommandBuilder instance.
func NewCommandBuilder() *CommandBuilder {
	return &CommandBuilder{}
}

func (c *CommandBuilder) Info() string {
	return "info\r\n"
}

func (c *CommandBuilder) Help() string {
	return "help\r\n"
}

func (c *CommandBuilder) ScanAll() string {
	return "scanall\r\n"
}

func (c *CommandBuilder) ScanAP() string {
	return "scanap\r\n"
}

func (c *CommandBuilder) ScanSTA() string {
	return "scansta\r\n"
}

func (c *CommandBuilder) StopScan() string {
	return "stopscan\r\n"
}

func (c *CommandBuilder) Channel(ch int) (string, error) {
	if ch < 1 || ch > 165 {
		return "", fmt.Errorf("invalid channel %d; must be between 1 and 165", ch)
	}
	return fmt.Sprintf("channel -s %d\r\n", ch), nil
}

func (c *CommandBuilder) SniffPMKID() string {
	return "sniffpmkid\r\n"
}

func (c *CommandBuilder) SniffBeacon() string {
	return "sniffbeacon\r\n"
}

func (c *CommandBuilder) SniffDeauth() string {
	return "sniffdeauth\r\n"
}

func (c *CommandBuilder) SniffRaw() string {
	return "sniffraw\r\n"
}

func (c *CommandBuilder) SniffBT() string {
	return "sniffbt\r\n"
}

func (c *CommandBuilder) Reboot() string {
	return "reboot\r\n"
}

func (c *CommandBuilder) ClearList() string {
	return "clearlist\r\n"
}

// FormatCustom validates and terminates arbitrary expert console commands.
func (c *CommandBuilder) FormatCustom(cmd string) (string, error) {
	clean := strings.TrimSpace(cmd)
	if clean == "" {
		return "", fmt.Errorf("command cannot be empty")
	}
	if strings.ContainsAny(clean, "\r\n") {
		return "", fmt.Errorf("command contains illegal newline characters")
	}
	return clean + "\r\n", nil
}

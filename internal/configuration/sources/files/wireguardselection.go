package files

import (
	"errors"
	"fmt"
	"net"

	"github.com/qdm12/gluetun/internal/configuration/settings"
	"github.com/qdm12/govalid/port"
	"gopkg.in/ini.v1"
)

var (
	ErrEndpointHostNotIP = errors.New("endpoint host is not an IP")
)

func (s *Source) readWireguardSelection() (selection settings.WireguardSelection, err error) {
	fileStringPtr, err := ReadFromFile(s.wireguardConfigPath)
	if err != nil {
		return selection, fmt.Errorf("reading file: %w", err)
	}

	if fileStringPtr == nil {
		return selection, nil
	}

	rawData := []byte(*fileStringPtr)
	iniFile, err := ini.Load(rawData)
	if err != nil {
		return selection, fmt.Errorf("loading ini from reader: %w", err)
	}

	peerSection, err := iniFile.GetSection("Peer")
	if err == nil {
		err = parseWireguardPeerSection(peerSection, &selection)
		if err != nil {
			return selection, fmt.Errorf("parsing peer section: %w", err)
		}
	} else if !regexINISectionNotExist.MatchString(err.Error()) {
		return selection, fmt.Errorf("getting peer section: %w", err)
	}

	return selection, nil
}

func parseWireguardPeerSection(peerSection *ini.Section,
	selection *settings.WireguardSelection) (err error) {
	publicKeyPtr, err := parseINIWireguardKey(peerSection, "PublicKey")
	if err != nil {
		return err // error is already wrapped correctly
	} else if publicKeyPtr != nil {
		selection.PublicKey = *publicKeyPtr
	}

	endpointKey, err := peerSection.GetKey("Endpoint")
	if err == nil {
		endpoint := endpointKey.String()
		host, portString, err := net.SplitHostPort(endpoint)
		if err != nil {
			return fmt.Errorf("splitting endpoint: %w", err)
		}

		ip := net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("%w: %s", ErrEndpointHostNotIP, host)
		}
		selection.EndpointIP = ip

		selection.EndpointPort = new(uint16)
		*selection.EndpointPort, err = port.Validate(portString)
		if err != nil {
			return fmt.Errorf("port from Endpoint key: %w", err)
		}
	} else if !regexINIKeyNotExist.MatchString(err.Error()) {
		return fmt.Errorf("getting endpoint key: %w", err)
	}

	return nil
}

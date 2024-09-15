package ipfetcher

import (
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

func shuffleServices(services []string) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(services), func(i, j int) { services[i], services[j] = services[j], services[i] })
	return services
}

func isValidIP(ip string, isIPv6 bool) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	if isIPv6 {
		return parsedIP.To4() == nil
	}
	return parsedIP.To4() != nil
}

func GetExternalIP() (string, error) {
	services := []string{
		"https://api.ipify.org?format=text",
		"https://ifconfig.co",
		"https://checkip.amazonaws.com",
		"https://myexternalip.com/raw",
	}

	services = shuffleServices(services)

	for _, service := range services {
		resp, err := http.Get(service)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			ip, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			ipStr := strings.TrimSpace(string(ip))

			if isValidIP(ipStr, false) {
				return ipStr, nil
			}
		}
	}
	return "", errors.New("could not fetch a valid external IPv4 address from any service")
}

func GetExternalIPv6() (string, error) {
	services := []string{
		"https://v6.ident.me",
		"ipv6.icanhazip.com",
		"https://api6.ipify.org",
	}

	services = shuffleServices(services)

	for _, service := range services {
		resp, err := http.Get(service)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			ip, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			ipStr := strings.TrimSpace(string(ip))

			if isValidIP(ipStr, true) {
				return ipStr, nil
			}
		}
	}
	return "", errors.New("could not fetch a valid external IPv6 address from any service")
}

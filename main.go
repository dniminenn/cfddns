package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cfddns/config"
	"cfddns/ipfetcher"
	"cfddns/providers"
	"cfddns/providers/clouddns"
	"cfddns/providers/cloudflare"
	"cfddns/providers/digitalocean"
	"cfddns/providers/duckdns"
	"cfddns/providers/freedns"
	"cfddns/providers/noip"
	"cfddns/providers/route53"

	"github.com/sirupsen/logrus"
)

func main() {
	runAsDaemon := flag.Bool("daemon", false, "Run as a daemon service")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	setupLogging(*verbose, *runAsDaemon)

	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Error loading configuration: %v", err)
	}

	if *runAsDaemon {
		runDaemon(cfg)
	} else {
		runOnce(cfg)
	}
}

func setupLogging(verbose, runAsDaemon bool) {
	// Set logging level based on verbosity
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Set up log formatter
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
		DisableColors:   runAsDaemon,
	}

	if !verbose && !runAsDaemon {
		formatter.TimestampFormat = "2006-01-02 15:04:05"
	}

	logrus.SetFormatter(formatter)
}

func runOnce(cfg *config.Config) {
	ipv4Address, err := ipfetcher.GetExternalIP()
	if err != nil {
		logrus.Warnf("Error fetching external IPv4 address: %v", err)
		ipv4Address = ""
	}

	ipv6Address, err := ipfetcher.GetExternalIPv6()
	if err != nil {
		logrus.Warnf("Error fetching external IPv6 address: %v", err)
		ipv6Address = ""
	}

	for _, providerCfg := range cfg.Providers {
		var provider providers.Provider

		switch providerCfg.Type {
		case "cloudflare":
			settings := providerCfg.Settings

			email, _ := settings["email"].(string)
			apiToken, _ := settings["apiToken"].(string)
			globalAPIKey, _ := settings["globalApiKey"].(string)
			zoneName, _ := settings["zone"].(string)

			provider = &cloudflare.CloudflareProvider{
				Email:        email,
				APIToken:     apiToken,
				GlobalAPIKey: globalAPIKey,
				ZoneName:     zoneName,
			}

		case "route53":
			settings := providerCfg.Settings

			zoneName, _ := settings["zone"].(string)
			region, _ := settings["region"].(string)
			accessKeyID, _ := settings["accessKeyId"].(string)
			secretAccessKey, _ := settings["secretAccessKey"].(string)
			if region == "" {
				region = "us-east-1"
			}

			provider = &route53.Route53Provider{
				ZoneName:        zoneName,
				Region:          region,
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
			}

		case "digitalocean":
			settings := providerCfg.Settings
			apiToken, _ := settings["apiToken"].(string)
			domain, _ := settings["domain"].(string)

			provider = &digitalocean.DigitalOceanProvider{
				APIToken: apiToken,
				Domain:   domain,
			}

		case "clouddns":
			settings := providerCfg.Settings
			projectID, _ := settings["projectId"].(string)
			credentialsJSONPath, _ := settings["credentialsJsonPath"].(string)
			zoneName, _ := settings["zone"].(string)

			// Read the credentials JSON file
			credentialsJSON, err := os.ReadFile(credentialsJSONPath)
			if err != nil {
				logrus.Errorf("Failed to read credentials JSON file: %v", err)
				continue
			}

			provider = &clouddns.CloudDNSProvider{
				ProjectID:       projectID,
				CredentialsJSON: credentialsJSON,
				ZoneName:        zoneName,
			}

		case "duckdns":
			settings := providerCfg.Settings
			token, _ := settings["token"].(string)

			provider = &duckdns.DuckDNSProvider{
				Token: token,
			}

		case "noip":
			settings := providerCfg.Settings
			username, _ := settings["username"].(string)
			password, _ := settings["password"].(string)

			provider = &noip.NoIPProvider{
				Username: username,
				Password: password,
			}

		case "freedns":
			provider = &freedns.FreeDNSProvider{}

		default:
			logrus.Errorf("Unsupported provider type: %s", providerCfg.Type)
			continue
		}

		for _, record := range providerCfg.Records {
			var ipAddress string

			if record.Type == "A" && ipv4Address != "" {
				ipAddress = ipv4Address
			} else if record.Type == "AAAA" && ipv6Address != "" {
				ipAddress = ipv6Address
			} else {
				logrus.Warnf("Skipping record %s of type %s due to missing IP", record.Name, record.Type)
				continue
			}

			dnsRecord := providers.DNSRecord{
				Name:        record.Name,
				Type:        record.Type,
				Content:     ipAddress,
				TTL:         record.TTL,
				Proxied:     record.Proxied,
				UpdateToken: record.UpdateToken,
			}

			err := provider.CommitRecord(dnsRecord)
			if err != nil {
				logrus.Errorf("Error updating DNS record for %s: %v", record.Name, err)
			}
		}
	}
}

func runDaemon(cfg *config.Config) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	updateInterval := time.Duration(cfg.GeneralSettings.UpdateInterval) * time.Second
	connectivityCheckInterval := time.Duration(cfg.GeneralSettings.ConnectivityCheckInterval) * time.Second

	var lastIPV4, lastIPV6 string
	var isConnected bool

	updateTimer := time.NewTimer(updateInterval)
	connectivityTicker := time.NewTicker(connectivityCheckInterval)

	defer updateTimer.Stop()
	defer connectivityTicker.Stop()

	// Immediate connectivity check and update
	connected := isInternetAvailable(cfg)
	if connected {
		isConnected = true
		logrus.Info("Daemon started and internet connection is available. Updating DNS records.")
		runOnce(cfg)
		lastIPV4, _ = ipfetcher.GetExternalIP()
		lastIPV6, _ = ipfetcher.GetExternalIPv6()
	} else {
		isConnected = false
		logrus.Warn("Daemon started but no internet connection is available.")
	}

	go func() {
		sig := <-sigs
		logrus.Infof("Received signal: %v, shutting down gracefully...", sig)
		done <- true
	}()

	for {
		select {
		case <-connectivityTicker.C:
			connected := isInternetAvailable(cfg)
			if connected && !isConnected {
				isConnected = true
				logrus.Info("Internet connection restored. Updating DNS records.")
				runOnce(cfg)
				lastIPV4, _ = ipfetcher.GetExternalIP()
				lastIPV6, _ = ipfetcher.GetExternalIPv6()
				// Reset the update timer
				if !updateTimer.Stop() {
					select {
					case <-updateTimer.C:
					default:
					}
				}
				updateTimer.Reset(updateInterval)
			} else if !connected && isConnected {
				isConnected = false
				logrus.Warn("Internet connection lost.")
			}
		case <-updateTimer.C:
			if isConnected {
				ipv4Address, err := ipfetcher.GetExternalIP()
				if err != nil {
					logrus.Warnf("Error fetching external IPv4 address: %v", err)
					ipv4Address = ""
				}

				ipv6Address, err := ipfetcher.GetExternalIPv6()
				if err != nil {
					logrus.Warnf("Error fetching external IPv6 address: %v", err)
					ipv6Address = ""
				}

				if ipv4Address != lastIPV4 || ipv6Address != lastIPV6 {
					logrus.Info("IP address changed. Updating DNS records.")
					runOnce(cfg)
					lastIPV4 = ipv4Address
					lastIPV6 = ipv6Address
				} else {
					logrus.Debug("IP address has not changed. No update necessary.")
				}
				// Reset the update timer
				updateTimer.Reset(updateInterval)
			} else {
				// If not connected, reset the update timer to wait for the next interval
				updateTimer.Reset(updateInterval)
			}
		case <-done:
			logrus.Info("Service stopped.")
			return
		}
	}
}

// isInternetAvailable checks if the internet connection is available
func isInternetAvailable(cfg *config.Config) bool {
	timeout := time.Second
	address := net.JoinHostPort(cfg.GeneralSettings.ConnectivityCheckIP, cfg.GeneralSettings.ConnectivityCheckPort)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

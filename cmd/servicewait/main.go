package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Service struct {
	Name       string
	Host       string
	Port       string
	Protocol   string
	Endpoint   string
	Timeout    time.Duration
	MaxRetries int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(color.CyanString("servicewait - Smart Service Dependency Waiter"))
		fmt.Println()
		fmt.Println("Usage: servicewait <service1> <service2> ...")
		fmt.Println("Format: name:host[:port][protocol][endpoint]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  servicewait db:localhost:5432:tcp")
		fmt.Println("  servicewait api:localhost:8080:http:/health")
		os.Exit(1)
	}

	services := parseServices(os.Args[1:])
	waitForServices(services)
}

func parseServices(args []string) []Service {
	var services []Service

	for _, arg := range args {
		service := parseServiceConfig(arg)
		services = append(services, service)
	}

	return services
}

func parseServiceConfig(config string) Service {
	service := Service{
		Timeout:  5 * time.Second,
		MaxRetries: 30,
	}

	parts := strings.Split(config, ":")
	if len(parts) >= 2 {
		service.Name = parts[0]
		service.Host = parts[1]
		
		if len(parts) > 2 {
			service.Port = parts[2]
		}
		
		if len(parts) > 3 {
			service.Protocol = parts[3]
		}
		
		if len(parts) > 4 {
			service.Endpoint = parts[4]
		}
	}

	return service
}

func waitForServices(services []Service) {
	fmt.Println(color.CyanString("\n=== SERVICE DEPENDENCY WAITER ===\n"))

	var unavailable []Service

	for _, service := range services {
		fmt.Printf("Waiting for %s (%s:%s)...\n", 
			service.Name, service.Host, service.Port)

		startTime := time.Now()
		available, _ := waitForService(service)
		elapsed := time.Since(startTime)

		if available {
			fmt.Printf("  ✓ %s is ready (%s)\n", 
				color.GreenString(service.Name), formatDuration(elapsed))
		} else {
			fmt.Printf("  ✗ %s failed to start (%s)\n", 
				color.RedString(service.Name), formatDuration(elapsed))
			unavailable = append(unavailable, service)
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d services ready, %d services unavailable\n", 
		len(services)-len(unavailable), len(unavailable))

	if len(unavailable) > 0 {
		fmt.Println(color.YellowString("\nFailed services:"))
		for _, s := range unavailable {
			fmt.Printf("  - %s (%s:%s)\n", s.Name, s.Host, s.Port)
		}
		os.Exit(1)
	}
}

func waitForService(service Service) (bool, bool) {
	for i := 0; i < service.MaxRetries; i++ {
		if checkService(service) {
			return true, true
		}

		time.Sleep(2 * time.Second)
	}

	return false, false
}

func checkService(service Service) bool {
	switch strings.ToLower(service.Protocol) {
	case "tcp", "":
		return checkTCP(service.Host, service.Port)
	case "http", "https":
		return checkHTTP(service.Host, service.Port, service.Endpoint)
	case "unix":
		return checkUnix(service.Host, service.Port)
	default:
		return checkTCP(service.Host, service.Port)
	}
}

func checkTCP(host, port string) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func checkHTTP(host, port, endpoint string) bool {
	url := fmt.Sprintf("%s://%s%s", 
		getProtocol(host), 
		fmt.Sprintf("%s:%s", host, port),
		getEndpoint(endpoint),
	)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func checkUnix(host, port string) bool {
	conn, err := net.DialTimeout("unix", fmt.Sprintf("%s/%s", host, port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func getProtocol(host string) string {
	if strings.Contains(host, "https") {
		return "https"
	}
	return "http"
}

func getEndpoint(endpoint string) string {
	if endpoint != "" {
		return endpoint
	}
	return "/"
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
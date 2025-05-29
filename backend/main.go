package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var knockPorts = []int{5222}

func knock(portArr []int, host string) {
	for _, port := range portArr {
		address := fmt.Sprintf("%s:%d", host, port)
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			conn.Close()
		}
		time.Sleep(300 * time.Millisecond)
	}
}
func removeLeaseFromHosts(ip string, hosts []string) (string, error) {
	hostname := ""
	var lastErr error
	for _, host := range hosts {
		if strings.Contains("94.228.84.26", host) {
			hostname = "hip.net.e-net.sk"
		} else {
			hostname = "hop.net.e-net.sk"
		}
		res, err := removeLease(ip, host)
		if err == nil {
			return fmt.Sprintf("==> Maze sa lease na %s:\n%s", hostname, res), nil
		}
		log.Printf("nepodarilo sa odstranit lease %s: %v\nOutput:\n%s", host, err, res)
		lastErr = err
	}
	return "", fmt.Errorf("nepodarilo sa odstranit lease: %w", lastErr)
}

func removeLease(ip string, host string) (string, error) {
	knock(knockPorts, host)
	keyPath := os.ExpandEnv("$HOME/.ssh/id_rsa")
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("nenacital spravne key : %w", err)
	}
	key, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return "", fmt.Errorf("tento kluc vyzera byt zasifrovany: %w", err)
	}
	if !strings.Contains(host, ":") {
		host += ":22"
	}
	config := &ssh.ClientConfig{
		User:            "brunovsky",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return "", fmt.Errorf("nespojilo sa ssh: %w", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("nedokazal som vytvorit novu session: %w", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("sudo /usr/local/bin/remove-lease.sh %s", ip)
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("ssh command sa neexecutol: %w", err)
	}
	return string(output), nil
}

func handleRemove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	removeIP := r.URL.Query().Get("ip")
	if removeIP == "" {
		http.Error(w, "musis pridat parameter ip=", http.StatusBadRequest)
		return
	}

	hosts := []string{
		"94.228.84.26",
		"94.228.84.27",
	}

	res, err := removeLeaseFromHosts(removeIP, hosts)
	if err != nil {
		http.Error(w, fmt.Sprintf("nepodarilo sa odstranit lease:\n%s", res), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, res)
}

func main() {
	http.HandleFunc("/", handleRemove)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Panic(err)
	}
}

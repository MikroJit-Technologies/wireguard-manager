package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Peer struct {
	Name          string `json:"name"`
	PublicKey     string `json:"public_key"`
	PresharedKey  string `json:"preshared_key,omitempty"`
	AllowedIPs    string `json:"allowed_ips"`
	Endpoint      string `json:"endpoint"`
	LastHandshake int64  `json:"last_handshake"`
	RxBytes       int64  `json:"rx_bytes"`
	TxBytes       int64  `json:"tx_bytes"`
	Online        bool   `json:"online"`
}

type Interface struct {
	PrivateKey string
	PublicKey  string
	Address    string
	ListenPort string
	DNS        string
	PostUp     []string
	PostDown   []string
}

// readConfig parses /etc/wireguard/wg0.conf
func readConfig(path string) (*Interface, []Peer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	var iface Interface
	var peers []Peer
	var cur *Peer

	scanner := bufio.NewScanner(f)
	section := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			if cur != nil && strings.HasPrefix(line, "# Name = ") {
				cur.Name = strings.TrimPrefix(line, "# Name = ")
			}
			continue
		}
		if line == "[Interface]" {
			section = "interface"
			continue
		}
		if line == "[Peer]" {
			if cur != nil {
				peers = append(peers, *cur)
			}
			cur = &Peer{}
			section = "peer"
			continue
		}
		k, v, _ := strings.Cut(line, " = ")
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		switch section {
		case "interface":
			switch k {
			case "PrivateKey":
				iface.PrivateKey = v
			case "Address":
				iface.Address = v
			case "ListenPort":
				iface.ListenPort = v
			case "DNS":
				iface.DNS = v
			case "PostUp":
				iface.PostUp = append(iface.PostUp, v)
			case "PostDown":
				iface.PostDown = append(iface.PostDown, v)
			}
		case "peer":
			switch k {
			case "PublicKey":
				cur.PublicKey = v
			case "PresharedKey":
				cur.PresharedKey = v
			case "AllowedIPs":
				cur.AllowedIPs = v
			case "Endpoint":
				cur.Endpoint = v
			}
		}
	}
	if cur != nil {
		peers = append(peers, *cur)
	}

	// derive server public key
	if iface.PrivateKey != "" {
		pub, err := wgPubKey(iface.PrivateKey)
		if err == nil {
			iface.PublicKey = pub
		}
	}

	return &iface, peers, scanner.Err()
}

// liveStats merges `wg show wg0 dump` into peer list
func liveStats(iface string, peers []Peer) []Peer {
	out, err := exec.Command("wg", "show", iface, "dump").Output()
	if err != nil {
		return peers
	}
	stats := map[string][]string{}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, l := range lines[1:] { // skip interface line
		f := strings.Fields(l)
		if len(f) >= 7 {
			// pubkey psk endpoint allowed-ips last-handshake rx tx
			stats[f[0]] = f
		}
	}
	for i := range peers {
		f, ok := stats[peers[i].PublicKey]
		if !ok {
			continue
		}
		peers[i].Endpoint = f[2]
		peers[i].LastHandshake, _ = strconv.ParseInt(f[4], 10, 64)
		peers[i].RxBytes, _ = strconv.ParseInt(f[5], 10, 64)
		peers[i].TxBytes, _ = strconv.ParseInt(f[6], 10, 64)
		peers[i].Online = time.Now().Unix()-peers[i].LastHandshake < 180
	}
	return peers
}

func addPeer(wgConf, name, pubkey, psk, allowedIPs string) error {
	entry := fmt.Sprintf("\n[Peer]\n# Name = %s\nPublicKey = %s\n", name, pubkey)
	if psk != "" {
		entry += fmt.Sprintf("PresharedKey = %s\n", psk)
	}
	entry += fmt.Sprintf("AllowedIPs = %s\n", allowedIPs)

	f, err := os.OpenFile(wgConf, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(entry)
	if err != nil {
		return err
	}

	// apply to running interface
	exec.Command("wg", "addconf", interfaceName(wgConf),
		fmt.Sprintf("<(echo '[Peer]\nPublicKey = %s\nAllowedIPs = %s')", pubkey, allowedIPs)).Run()
	exec.Command("wg", "set", interfaceName(wgConf), "peer", pubkey, "allowed-ips", allowedIPs).Run()
	return nil
}

func removePeer(wgConf, pubkey string) error {
	// remove from running interface first
	exec.Command("wg", "set", interfaceName(wgConf), "peer", pubkey, "remove").Run()

	// rewrite config file without this peer
	data, err := os.ReadFile(wgConf)
	if err != nil {
		return err
	}

	// block-based removal
	blocks := strings.Split(string(data), "\n[Peer]")
	newContent := blocks[0]
	for _, block := range blocks[1:] {
		if !strings.Contains(block, pubkey) {
			newContent += "\n[Peer]" + block
		}
	}
	return os.WriteFile(wgConf, []byte(newContent), 0600)
}

func genKey() (priv, pub, psk string, err error) {
	privBytes, err := exec.Command("wg", "genkey").Output()
	if err != nil {
		return "", "", "", fmt.Errorf("genkey: %w", err)
	}
	priv = strings.TrimSpace(string(privBytes))

	pub, err = wgPubKey(priv)
	if err != nil {
		return "", "", "", err
	}

	pskBytes, err := exec.Command("wg", "genpsk").Output()
	if err != nil {
		return "", "", "", fmt.Errorf("genpsk: %w", err)
	}
	psk = strings.TrimSpace(string(pskBytes))
	return priv, pub, psk, nil
}

func wgPubKey(privKey string) (string, error) {
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privKey)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func interfaceName(wgConf string) string {
	base := wgConf[strings.LastIndex(wgConf, "/")+1:]
	return strings.TrimSuffix(base, ".conf")
}

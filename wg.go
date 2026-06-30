package main

import (
	"bufio"
	"fmt"
	"net"
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
	HasConfig     bool   `json:"has_config"`
}

type WGInterface struct {
	PrivateKey string
	PublicKey  string
	Address    string
	ListenPort string
	DNS        string
}

func readConfig(path string) (*WGInterface, []Peer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	var iface WGInterface
	var peers []Peer
	var cur *Peer

	scanner := bufio.NewScanner(f)
	section := ""
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			if cur != nil && strings.HasPrefix(trimmed, "# Name = ") {
				cur.Name = strings.TrimPrefix(trimmed, "# Name = ")
			}
			continue
		}
		if trimmed == "[Interface]" {
			section = "interface"
			continue
		}
		if trimmed == "[Peer]" {
			if cur != nil {
				peers = append(peers, *cur)
			}
			cur = &Peer{}
			section = "peer"
			continue
		}
		k, v, _ := strings.Cut(trimmed, " = ")
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
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
	if iface.PrivateKey != "" {
		if pub, err := wgPubKey(iface.PrivateKey); err == nil {
			iface.PublicKey = pub
		}
	}
	return &iface, peers, scanner.Err()
}

func liveStats(ifaceName string, peers []Peer) []Peer {
	out, err := exec.Command("wg", "show", ifaceName, "dump").Output()
	if err != nil {
		return peers
	}
	stats := map[string][]string{}
	for i, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if i == 0 {
			continue
		}
		if f := strings.Fields(l); len(f) >= 7 {
			stats[f[0]] = f
		}
	}
	now := time.Now().Unix()
	for i := range peers {
		f, ok := stats[peers[i].PublicKey]
		if !ok {
			continue
		}
		if f[2] != "(none)" {
			peers[i].Endpoint = f[2]
		}
		peers[i].LastHandshake, _ = strconv.ParseInt(f[4], 10, 64)
		peers[i].RxBytes, _ = strconv.ParseInt(f[5], 10, 64)
		peers[i].TxBytes, _ = strconv.ParseInt(f[6], 10, 64)
		peers[i].Online = peers[i].LastHandshake > 0 && now-peers[i].LastHandshake < 180
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
	if _, err = f.WriteString(entry); err != nil {
		return err
	}

	iface := interfaceName(wgConf)
	exec.Command("wg", "set", iface, "peer", pubkey, "allowed-ips", allowedIPs).Run()

	if psk != "" {
		tmp, err := os.CreateTemp("", "wg-psk-*")
		if err == nil {
			tmp.WriteString(psk)
			tmp.Close()
			exec.Command("wg", "set", iface, "peer", pubkey, "preshared-key", tmp.Name()).Run()
			os.Remove(tmp.Name())
		}
	}
	return nil
}

func removePeer(wgConf, pubkey string) error {
	exec.Command("wg", "set", interfaceName(wgConf), "peer", pubkey, "remove").Run()

	data, err := os.ReadFile(wgConf)
	if err != nil {
		return err
	}
	blocks := strings.Split(string(data), "\n[Peer]")
	out := blocks[0]
	for _, block := range blocks[1:] {
		if !strings.Contains(block, pubkey) {
			out += "\n[Peer]" + block
		}
	}
	return os.WriteFile(wgConf, []byte(out), 0600)
}

func nextAvailableIP(wgConf string) (string, error) {
	iface, peers, err := readConfig(wgConf)
	if err != nil {
		return "", err
	}
	if iface.Address == "" {
		return "", fmt.Errorf("interface has no Address configured")
	}
	serverIP, ipNet, err := net.ParseCIDR(iface.Address)
	if err != nil {
		return "", fmt.Errorf("invalid interface address: %w", err)
	}
	used := map[string]bool{serverIP.String(): true}
	for _, p := range peers {
		for _, cidr := range strings.Split(p.AllowedIPs, ",") {
			if ip, _, e := net.ParseCIDR(strings.TrimSpace(cidr)); e == nil {
				used[ip.String()] = true
			}
		}
	}
	ip := serverIP.Mask(ipNet.Mask).To4()
	if ip == nil {
		return "", fmt.Errorf("only IPv4 supported for auto-IP")
	}
	incIP(ip)
	for ipNet.Contains(ip) {
		s := ip.String()
		if !used[s] && !strings.HasSuffix(s, ".0") && !strings.HasSuffix(s, ".255") {
			return s + "/32", nil
		}
		incIP(ip)
	}
	return "", fmt.Errorf("no available IPs in %s", ipNet)
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func genKey() (priv, pub, psk string, err error) {
	privBytes, err := exec.Command("wg", "genkey").Output()
	if err != nil {
		return "", "", "", fmt.Errorf("wg genkey: %w", err)
	}
	priv = strings.TrimSpace(string(privBytes))
	if pub, err = wgPubKey(priv); err != nil {
		return "", "", "", err
	}
	pskBytes, err := exec.Command("wg", "genpsk").Output()
	if err != nil {
		return "", "", "", fmt.Errorf("wg genpsk: %w", err)
	}
	psk = strings.TrimSpace(string(pskBytes))
	return priv, pub, psk, nil
}

func wgPubKey(privKey string) (string, error) {
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privKey)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("wg pubkey: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func interfaceName(wgConf string) string {
	base := wgConf[strings.LastIndex(wgConf, "/")+1:]
	return strings.TrimSuffix(base, ".conf")
}

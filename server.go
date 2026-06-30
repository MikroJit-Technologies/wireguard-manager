package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type server struct {
	cfg *Config
	mux *http.ServeMux
}

func newServer(cfg *Config) http.Handler {
	s := &server{cfg: cfg, mux: http.NewServeMux()}
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/peers", s.handlePeers)
	s.mux.HandleFunc("/api/peers/", s.handlePeer)
	s.mux.HandleFunc("/api/generate", s.handleGenerate)
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	if cfg.Auth.Username != "" {
		return basicAuth(cfg.Auth.Username, cfg.Auth.Password, s.mux)
	}
	return s.mux
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	iface, peers, _ := readConfig(s.cfg.WgConfig)
	peers = liveStats(s.cfg.Interface, peers)

	online := 0
	for _, p := range peers {
		if p.Online {
			online++
		}
	}

	data := map[string]any{
		"Interface": s.cfg.Interface,
		"Peers":     peers,
		"Total":     len(peers),
		"Online":    online,
		"ServerPub": "",
	}
	if iface != nil {
		data["ServerPub"] = iface.PublicKey
		data["ServerAddr"] = iface.Address
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	renderHTML(w, data)
}

func (s *server) handlePeers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		_, peers, err := readConfig(s.cfg.WgConfig)
		if err != nil {
			jsonErr(w, err, 500)
			return
		}
		peers = liveStats(s.cfg.Interface, peers)
		json.NewEncoder(w).Encode(peers)

	case http.MethodPost:
		var req struct {
			Name       string `json:"name"`
			PublicKey  string `json:"public_key"`
			PresharedKey string `json:"preshared_key"`
			AllowedIPs string `json:"allowed_ips"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonErr(w, err, 400)
			return
		}
		if req.AllowedIPs == "" {
			req.AllowedIPs = s.cfg.Defaults.AllowedIPs
		}
		if err := addPeer(s.cfg.WgConfig, req.Name, req.PublicKey, req.PresharedKey, req.AllowedIPs); err != nil {
			jsonErr(w, err, 500)
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		w.WriteHeader(405)
	}
}

func (s *server) handlePeer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/peers/"), "/")
	pubkey := parts[0]

	if len(parts) == 2 && parts[1] == "qr" {
		s.handleQR(w, r, pubkey)
		return
	}
	if len(parts) == 2 && parts[1] == "config" {
		s.handlePeerConfig(w, r, pubkey)
		return
	}

	if r.Method == http.MethodDelete {
		if err := removePeer(s.cfg.WgConfig, pubkey); err != nil {
			jsonErr(w, err, 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}
	w.WriteHeader(405)
}

func (s *server) handleQR(w http.ResponseWriter, r *http.Request, pubkey string) {
	cfg, err := s.buildPeerConfig(pubkey)
	if err != nil {
		jsonErr(w, err, 404)
		return
	}
	png, err := generateQR(cfg)
	if err != nil {
		jsonErr(w, err, 500)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

func (s *server) handlePeerConfig(w http.ResponseWriter, r *http.Request, pubkey string) {
	cfg, err := s.buildPeerConfig(pubkey)
	if err != nil {
		jsonErr(w, err, 404)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=wg-client.conf")
	w.Write([]byte(cfg))
}

func (s *server) buildPeerConfig(pubkey string) (string, error) {
	iface, peers, err := readConfig(s.cfg.WgConfig)
	if err != nil {
		return "", err
	}
	for _, p := range peers {
		if p.PublicKey == pubkey && p.PresharedKey != "" {
			// we stored the private key at add time — we don't re-derive it
			// config download only works if private key was returned at creation
			_ = iface
			return "", fmt.Errorf("private key not stored — use the config returned at peer creation")
		}
	}
	return "", fmt.Errorf("peer not found or private key unavailable")
}

func (s *server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	iface, _, err := readConfig(s.cfg.WgConfig)
	if err != nil {
		jsonErr(w, err, 500)
		return
	}

	priv, pub, psk, err := genKey()
	if err != nil {
		jsonErr(w, err, 500)
		return
	}

	serverPub := ""
	if iface != nil {
		serverPub = iface.PublicKey
	}

	json.NewEncoder(w).Encode(map[string]string{
		"private_key":   priv,
		"public_key":    pub,
		"preshared_key": psk,
		"server_pub":    serverPub,
		"server_endpoint": s.cfg.Server.Endpoint,
		"dns":           s.cfg.Server.DNS,
		"allowed_ips":   s.cfg.Defaults.AllowedIPs,
	})
}

func basicAuth(user, pass string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != user || p != pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="wireguard-manager"`)
			w.WriteHeader(401)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonErr(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

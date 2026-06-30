package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type appServer struct {
	cfg *Config
	mux *http.ServeMux
}

func newServer(cfg *Config) http.Handler {
	s := &appServer{cfg: cfg, mux: http.NewServeMux()}
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/peers", s.handlePeers)
	s.mux.HandleFunc("/api/peers/", s.handlePeer)
	s.mux.HandleFunc("/api/generate", s.handleGenerate)
	s.mux.HandleFunc("/api/suggest-ip", s.handleSuggestIP)
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	if cfg.Auth.Username != "" {
		return basicAuth(cfg.Auth.Username, cfg.Auth.Password, s.mux)
	}
	return s.mux
}

func (s *appServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	iface, _, _ := readConfig(s.cfg.WgConfig)

	serverPub := ""
	if iface != nil {
		serverPub = iface.PublicKey
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	renderHTML(w, map[string]any{
		"Interface":      s.cfg.Interface,
		"ServerPub":      serverPub,
		"DefaultIPs":     s.cfg.Defaults.AllowedIPs,
		"ServerEndpoint": s.cfg.Server.Endpoint,
		"ServerDNS":      s.cfg.Server.DNS,
	})
}

func (s *appServer) handlePeers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		_, peers, err := readConfig(s.cfg.WgConfig)
		if err != nil {
			jsonErr(w, err, 500)
			return
		}
		peers = liveStats(s.cfg.Interface, peers)
		for i := range peers {
			_, peers[i].HasConfig = cfgStore.Get(peers[i].PublicKey)
		}
		json.NewEncoder(w).Encode(peers)

	case http.MethodPost:
		var req struct {
			Name         string `json:"name"`
			PublicKey    string `json:"public_key"`
			PresharedKey string `json:"preshared_key"`
			AllowedIPs   string `json:"allowed_ips"`
			ClientConfig string `json:"client_config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonErr(w, err, 400)
			return
		}
		if req.PublicKey == "" {
			jsonErr(w, fmt.Errorf("public_key is required"), 400)
			return
		}
		if req.AllowedIPs == "" {
			req.AllowedIPs = s.cfg.Defaults.AllowedIPs
		}
		if err := addPeer(s.cfg.WgConfig, req.Name, req.PublicKey, req.PresharedKey, req.AllowedIPs); err != nil {
			jsonErr(w, err, 500)
			return
		}
		if req.ClientConfig != "" {
			cfgStore.Set(req.PublicKey, req.ClientConfig)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		w.WriteHeader(405)
	}
}

func (s *appServer) handlePeer(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/peers/"), "/")
	pubkey := parts[0]
	if pubkey == "" {
		w.WriteHeader(400)
		return
	}

	if len(parts) == 2 {
		switch parts[1] {
		case "qr":
			s.handleQR(w, pubkey)
			return
		case "config":
			s.handlePeerConfig(w, pubkey)
			return
		}
	}

	if r.Method == http.MethodDelete {
		w.Header().Set("Content-Type", "application/json")
		if err := removePeer(s.cfg.WgConfig, pubkey); err != nil {
			jsonErr(w, err, 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}
	w.WriteHeader(405)
}

func (s *appServer) handleQR(w http.ResponseWriter, pubkey string) {
	clientCfg, ok := cfgStore.Get(pubkey)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		jsonErr(w, fmt.Errorf("config not available — peer was added without key generation"), 404)
		return
	}
	png, err := generateQR(clientCfg)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		jsonErr(w, err, 500)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

func (s *appServer) handlePeerConfig(w http.ResponseWriter, pubkey string) {
	clientCfg, ok := cfgStore.Get(pubkey)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		jsonErr(w, fmt.Errorf("config not available — peer was added without key generation"), 404)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=wg-client.conf")
	w.Write([]byte(clientCfg))
}

func (s *appServer) handleGenerate(w http.ResponseWriter, r *http.Request) {
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
		"private_key":     priv,
		"public_key":      pub,
		"preshared_key":   psk,
		"server_pub":      serverPub,
		"server_endpoint": s.cfg.Server.Endpoint,
		"dns":             s.cfg.Server.DNS,
		"allowed_ips":     s.cfg.Defaults.AllowedIPs,
	})
}

func (s *appServer) handleSuggestIP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ip, err := nextAvailableIP(s.cfg.WgConfig)
	if err != nil {
		jsonErr(w, err, 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"ip": ip})
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

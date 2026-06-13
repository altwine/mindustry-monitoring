package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Address struct {
	Host string
	Port int
}

type Server struct {
	Name    string
	Address []Address
}

type rawServer struct {
	Name        string   `json:"name"`
	Address     []string `json:"address"`
	Prioritized bool     `json:"prioritized,omitempty"`
}

const SERVERS_URL = "https://raw.githubusercontent.com/Anuken/MindustryServerList/refs/heads/main/servers_v8.json"

var servers = []Server{}

func initServers() {
	resp, err := http.Get(SERVERS_URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var rawServers []rawServer
	if err := json.Unmarshal(bodyBytes, &rawServers); err != nil {
		panic(err)
	}

	for _, rs := range rawServers {
		addresses := make([]Address, 0, len(rs.Address))
		for _, addrStr := range rs.Address {
			host, port := parseAddress(addrStr)
			addresses = append(addresses, Address{Host: host, Port: port})
		}
		servers = append(servers, Server{
			Name:    rs.Name,
			Address: addresses,
		})
	}
}

func parseAddress(addr string) (string, int) {
	parts := strings.SplitN(addr, ":", 2)
	host := parts[0]
	port := 6567
	if len(parts) == 2 {
		if p, err := strconv.Atoi(parts[1]); err == nil {
			port = p
		}
	}
	return host, port
}

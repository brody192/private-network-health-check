package main

import (
	"net"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).DialContext,
	},
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"main/internal/logger"
	"net"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func checkReplicasHandler(targetURL url.URL, client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		replicaResponse := &ReplicaResponse{}

		// Perform an AAAA DNS lookup for the target host (IPv6 only)
		dnsCtx, dnsCancel := context.WithTimeout(context.Background(), (10 * time.Second))
		defer dnsCancel()

		ips, err := net.DefaultResolver.LookupIP(dnsCtx, "ip6", targetURL.Hostname())
		if err != nil {
			logger.Stderr.Error("failed to lookup IPv6 address", logger.ErrAttr(err))

			replicaResponse.ServerError = ptr(err.Error())

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(replicaResponse)

			return
		}

		replicaResponse.TotalReplicas = len(ips)

		if replicaResponse.TotalReplicas == 0 {
			err := errors.New("no IPv6 addresses found for target host")

			logger.Stderr.Error(err.Error(), slog.String("target_host", targetURL.Hostname()))

			replicaResponse.ServerError = ptr(err.Error())

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(replicaResponse)

			return
		}

		replicaResponse.ReplicaResponses = make([]ReplicaResponseItem, replicaResponse.TotalReplicas)

		wg := sync.WaitGroup{}
		wg.Add(replicaResponse.TotalReplicas)

		for i, ip := range ips {
			go func(i int, ip net.IP, url url.URL) {
				defer wg.Done()

				replicaResponse.ReplicaResponses[i].IpAddress = ip

				startTime := time.Now()

				resp, err := client.Get(fmt.Sprintf("http://[%s]:%s%s", ip.String(), url.Port(), url.Path))

				replicaResponse.ReplicaResponses[i].ResponseTime = time.Since(startTime).Milliseconds()

				if err != nil {
					logger.Stderr.Error(err.Error(), slog.String("target_host", targetURL.Hostname()))

					replicaResponse.ReplicaResponses[i].Error = ptr(err.Error())

					return
				}

				defer resp.Body.Close()

				replicaResponse.ReplicaResponses[i].StatusCode = resp.StatusCode

				body, err := io.ReadAll(io.LimitReader(resp.Body, 128))
				if err != nil {
					logger.Stderr.Error(err.Error(), slog.String("target_host", targetURL.Hostname()))

					replicaResponse.ReplicaResponses[i].Error = ptr(err.Error())

					return
				}

				replicaResponse.ReplicaResponses[i].ResponseBody = string(body)
			}(i, ip, targetURL)
		}

		wg.Wait()

		slices.SortStableFunc(replicaResponse.ReplicaResponses, func(a, b ReplicaResponseItem) int {
			return bytes.Compare(a.IpAddress, b.IpAddress)
		})

		for _, replicaResponseItem := range replicaResponse.ReplicaResponses {
			replicaResponse.TotalResponseTime += replicaResponseItem.ResponseTime

			if replicaResponseItem.Error != nil || (replicaResponseItem.StatusCode < 200 || replicaResponseItem.StatusCode > 299) {
				replicaResponse.OfflineReplicas++
			}
		}

		replicaResponse.OnlineReplicas = replicaResponse.TotalReplicas - replicaResponse.OfflineReplicas

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(replicaResponse)
	}
}

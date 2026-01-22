package main

import "net"

type ReplicaResponse struct {
	ServerError *string `json:"server_error"`

	TotalReplicas   int `json:"total_replicas"`
	OnlineReplicas  int `json:"online_replicas"`
	OfflineReplicas int `json:"offline_replicas"`

	TotalResponseTime int64 `json:"total_response_time"`

	ReplicaResponses []ReplicaResponseItem `json:"replica_responses"`
}

type ReplicaResponseItem struct {
	IpAddress    net.IP  `json:"ip_address"`
	StatusCode   int     `json:"status_code"`
	ResponseTime int64   `json:"response_time"`
	ResponseBody string  `json:"response_body"`
	Error        *string `json:"error"`
}

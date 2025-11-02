package model

import "dlbackend/pkg/client"

type FileinfoResponse struct {
	Fileinfo    client.OneFichierInfoResponse `json:"fileinfo"`
	Directories map[DownloadType][]string     `json:"directories"`
}

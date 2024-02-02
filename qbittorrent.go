package main

import (
	"strings"
	"strconv"
	"encoding/json"
	"net/url"
)

type MainDataStruct struct {
	FullUpdate bool                     `json:"full_update"`
	Torrents   map[string]TorrentStruct `json:"torrents"`
}
type TorrentStruct struct {
	NumLeechs int64 `json:"num_leechs"`
	TotalSize int64 `json:"total_size"`
}
type PeerStruct struct {
	IP       string
	Port     int
	Client   string
	Progress float64
	Uploaded int64
}
type TorrentPeersStruct struct {
	FullUpdate bool                  `json:"full_update"`
	Peers      map[string]PeerStruct `json:"peers"`
}

var useNewBanPeersMethod = false

func Login() bool {
	if config.QBUsername == "" {
		return true
	}
	loginParams := url.Values {}
	loginParams.Set("username", config.QBUsername)
	loginParams.Set("password", config.QBPassword)
	loginResponseBody := Submit(config.QBURL + "/api/v2/auth/login", loginParams.Encode())
	if loginResponseBody == nil {
		Log("Login", "登录时发生了错误", true)
		return false
	}

	loginResponseBodyStr := StrTrim(string(loginResponseBody))
	if loginResponseBodyStr == "Ok." {
		Log("Login", "登录成功", true)
		return true
	} else if loginResponseBodyStr == "Fails." {
		Log("Login", "登录失败: 账号或密码错误", true)
	} else {
		Log("Login", "登录失败: %s", true, loginResponseBodyStr)
	}
	return false
}
func FetchMaindata() *MainDataStruct {
	maindataResponseBody := Fetch(config.QBURL + "/api/v2/sync/maindata?rid=0")
	if maindataResponseBody == nil {
		Log("FetchMaindata", "发生错误", true)
		return nil
	}

	var mainDataResult MainDataStruct
	if err := json.Unmarshal(maindataResponseBody, &mainDataResult); err != nil {
		Log("FetchMaindata", "解析时发生了错误: %s", true, err.Error())
		return nil
	}

	Log("Debug-FetchMaindata", "完整更新: %s", false, strconv.FormatBool(mainDataResult.FullUpdate))

	return &mainDataResult
}
func FetchTorrentPeers(infoHash string) *TorrentPeersStruct {
	torrentPeersResponseBody := Fetch(config.QBURL + "/api/v2/sync/torrentPeers?rid=0&hash=" + infoHash)
	if torrentPeersResponseBody == nil {
		Log("FetchTorrentPeers", "发生错误", true)
		return nil
	}

	var torrentPeersResult TorrentPeersStruct
	if err := json.Unmarshal(torrentPeersResponseBody, &torrentPeersResult); err != nil {
		Log("FetchTorrentPeers", "解析时发生了错误: %s", true, err.Error())
		return nil
	}

	if config.LogDebug_CheckTorrent {
		Log("Debug-FetchTorrentPeers", "完整更新: %s", false, strconv.FormatBool(torrentPeersResult.FullUpdate))
	}

	return &torrentPeersResult
}
func GenBlockPeersStr() string {
	ip_ports := ""
	if useNewBanPeersMethod {
		for peerIP, peerInfo := range blockPeerMap {
			if peerInfo.Port == -1 {
				for port := 0; port <= 65535; port++ {
					ip_ports += peerIP + ":" + strconv.Itoa(port) + "|"
				}
			} else {
				ip_ports += peerIP + ":" + strconv.Itoa(peerInfo.Port) + "|"
			}
		}
		ip_ports = strings.TrimRight(ip_ports, "|")
	} else {
		for peerIP := range blockPeerMap {
			ip_ports += peerIP + "\n"
		}
	}
	return ip_ports
}
func SubmitBlockPeer(banIPPortsStr string) {
	var banResponseBody []byte
	if useNewBanPeersMethod && banIPPortsStr != "" {
		banIPPortsStr = url.QueryEscape(banIPPortsStr)
		banResponseBody = Submit(config.QBURL + "/api/v2/transfer/banPeers", banIPPortsStr)
	} else {
		banIPPortsStr = url.QueryEscape("{\"banned_IPs\": \"" + banIPPortsStr + "\"}")
		banResponseBody = Submit(config.QBURL + "/api/v2/app/setPreferences", "json=" + banIPPortsStr)
	}
	if banResponseBody == nil {
		Log("SubmitBlockPeer", "发生错误", true)
	}
}
package httpClient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const hideMyNameUrl = "http://hidemy.name/ru/api/proxylist.php?out=js"

type HideMyNameProxy struct {
	Host        string  `json:"host"`
	Ip          string  `json:"ip"`
	Port        string  `json:"port"`
	Lastseen    int     `json:"lastseen"`
	Delay       int     `json:"delay"`
	Cid         string  `json:"cid"`
	CountryCode *string `json:"country_code"`
	CountryName *string `json:"country_name"`
	City        *string `json:"city"`
	ChecksUp    string  `json:"checks_up"`
	ChecksDown  string  `json:"checks_down"`
	Anon        string  `json:"anon"`
	Http        string  `json:"http"`
	Ssl         string  `json:"ssl"`
	Socks4      string  `json:"socks4"`
	Socks5      string  `json:"socks5"`
}

type proxyManager struct {
	log          *zap.Logger
	mu           sync.Mutex
	proxies      []string
	totalProxies int
	lastProxyIdx int
}

type ProxyManager interface {
	GetProxy() string
}

type ProxyFun func() ([]string, error)

func ProxyFromList(proxyList []string) ProxyFun {
	return func() ([]string, error) {
		return proxyList, nil
	}
}

func ProxyFromHideMyNameUrl(apiToken string) ([]string, error) {
	var (
		proxyList []string
		tmpProxy  []HideMyNameProxy
	)
	hideMyNameApiUrl := fmt.Sprintf("%s&key=%s", hideMyNameUrl, apiToken)
	_, body, err := fasthttp.Get(nil, hideMyNameApiUrl)
	if err != nil {
		return proxyList, err
	}
	err = json.Unmarshal(body, &tmpProxy)
	if err != nil {
		return proxyList, err
	}
	for _, pr := range tmpProxy {
		if pr.Socks5 == "1" {
			proxyList = append(proxyList, fmt.Sprintf("socks5://%s:%s", pr.Ip, pr.Port))
		} else if pr.Socks4 == "1" {
			proxyList = append(proxyList, fmt.Sprintf("socks4://%s:%s", pr.Ip, pr.Port))
		} else {
			proxyList = append(proxyList, fmt.Sprintf("%s:%s", pr.Ip, pr.Port))
		}
	}
	return proxyList, nil
}

func ProxyFromHideMyNameFile(filePath string) ([]string, error) {
	var (
		proxyList []string
		tmpProxy  []HideMyNameProxy
	)

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return proxyList, err
	}
	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(jsonFile)
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return proxyList, err
	}
	err = json.Unmarshal(byteValue, &tmpProxy)
	if err != nil {
		return proxyList, err
	}
	for _, pr := range tmpProxy {
		if pr.Socks5 == "1" {
			proxyList = append(proxyList, fmt.Sprintf("socks5://%s:%s", pr.Ip, pr.Port))
		} else if pr.Socks4 == "1" {
			proxyList = append(proxyList, fmt.Sprintf("socks4://%s:%s", pr.Ip, pr.Port))
		} else {
			proxyList = append(proxyList, fmt.Sprintf("%s:%s", pr.Ip, pr.Port))
		}
	}
	return proxyList, nil
}

func NewProxyManager(log *zap.Logger, updateInterval time.Duration, proxyFuns ...ProxyFun) ProxyManager {
	logger := log.Named("proxy_manager")
	var proxyList []string
	for _, fun := range proxyFuns {
		tmpProxy, err := fun()
		if err != nil {
			logger.Error("can not call proxy func", zap.Error(err))
			continue
		}
		proxyList = append(proxyList, tmpProxy...)
	}
	pm := &proxyManager{
		log:          log,
		proxies:      proxyList,
		mu:           sync.Mutex{},
		totalProxies: len(proxyList),
		lastProxyIdx: 0,
	}
	if updateInterval != 0 {
		if updateInterval < time.Minute {
			updateInterval = time.Minute
		}
		go pm.updater(updateInterval, proxyFuns...)
	}
	return pm
}

func (p *proxyManager) GetProxy() string {
	if p.totalProxies == 1 {
		return p.proxies[0]
	}
	p.mu.Lock()
	switch p.lastProxyIdx {
	case p.totalProxies - 1:
		p.lastProxyIdx = 0
	default:
		p.lastProxyIdx += 1
	}
	p.mu.Unlock()
	return p.proxies[p.lastProxyIdx]
}

func (p *proxyManager) updater(updateInterval time.Duration, proxyFuncs ...ProxyFun) {
	p.log.Info("proxy updater started", zap.Duration("interval", updateInterval))
	ticker := time.NewTicker(updateInterval)
	for {
		select {
		case <-ticker.C:
			p.log.Info("updating proxy list")
			var proxyList []string
			for _, fun := range proxyFuncs {
				tmpProxy, err := fun()
				if err != nil {
					p.log.Error("can not call proxy func in updater", zap.Error(err))
					continue
				}
				proxyList = append(proxyList, tmpProxy...)
			}
			p.mu.Lock()
			p.proxies = proxyList
			p.totalProxies = len(proxyList)
			p.lastProxyIdx = 0
			p.mu.Unlock()
			p.log.Info("proxy list updated")
		}
	}
}

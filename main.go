package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sagernet/sing-box/common/srs"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
	"golang.org/x/sync/singleflight"
)

const URL = "https://static-file-global.353355.xyz/rules/cn-additional-list.txt"

const HEADER = "{\"version\":3,\"rules\":[{\"domain_suffix\":["
const FOOTER = "]}]}"

var cache sync.Map
var sfg singleflight.Group

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/rule-set", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "cn-additional" {
			// 根据url获取资源
			ruleSet, err := getResource(URL)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			plainRuleSet, err := json.UnmarshalExtended[option.PlainRuleSetCompat]([]byte(ruleSet))
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.srs"`, name))
			srs.Write(w, plainRuleSet.Options, 3)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			// 定时加载
			loadResource(URL)
		}
	}()
	log.Println("service started")
	http.ListenAndServe(":9080", mux)
}

func getResource(url string) (string, error) {
	v, ok := cache.Load(url)
	if ok {
		log.Println("从缓存加载", url)
		return v.(string), nil
	} else {
		return loadResource(url)
	}
}

func loadResource(url string) (string, error) {
	// 避免并发
	v, err, _ := sfg.Do(url, func() (any, error) {
		rsp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer rsp.Body.Close()

		ruleSet := HEADER
		scanner := bufio.NewScanner(rsp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) == 0 || line[0] == '#' {
				continue
			} else {
				rule := "\"" + line + "\","
				ruleSet += rule
			}
		}
		ruleSet = strings.TrimSuffix(ruleSet, ",")
		ruleSet += FOOTER
		log.Println("缓存已更新", url)
		cache.Store(url, ruleSet)
		return ruleSet, nil
	})

	if err != nil {
		return "", err
	}
	return v.(string), nil
}

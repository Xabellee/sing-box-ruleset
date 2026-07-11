package main

import (
	"bufio"
	"net/http"
	"os"
	"strings"

	"github.com/sagernet/sing-box/common/convertor/adguard"
	"github.com/sagernet/sing-box/common/srs"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

func main() {
	data, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	var configList ConfigList
	json.Unmarshal(data, &configList)

	ruleSetList := make(map[string]option.PlainRuleSetCompat)

	for _, config := range configList.Source {
		name := config.Name
		switch config.Type {
		case "raw":
			ruleSetList[name] = handleRaw(config)
		case "static":
			ruleSetList[name] = handleStatic(config)
		case "adguard":
			ruleSetList[name] = handleAdguard(config)
		}
	}
	output(ruleSetList)
}

func handleAdguard(c Config) option.PlainRuleSetCompat {
	url := c.Url
	rsp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	rules, err := adguard.ToOptions(rsp.Body, log.StdLogger())
	ruleSet := option.PlainRuleSet{Rules: rules}
	plainRuleSet := option.PlainRuleSetCompat{Options: ruleSet, Version: 3}
	return plainRuleSet
}

func handleStatic(c Config) option.PlainRuleSetCompat {
	return c.Content
}

func handleRaw(c Config) option.PlainRuleSetCompat {
	const HEADER = "{\"version\":3,\"rules\":[{\"domain_suffix\":["
	const FOOTER = "]}]}"
	url := c.Url
	rsp, err := http.Get(url)
	if err != nil {
		panic(err)
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
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	ruleSet = strings.TrimSuffix(ruleSet, ",")
	ruleSet += FOOTER
	plainRuleSet := option.PlainRuleSetCompat{}
	json.Unmarshal([]byte(ruleSet), &plainRuleSet)
	return plainRuleSet
}

func output(ruleSetList map[string]option.PlainRuleSetCompat) {
	for k, v := range ruleSetList {
		writer, err := os.Create(k + ".srs")
		if err != nil {
			panic(err)
		}
		srs.Write(writer, v.Options, 3)
	}
}

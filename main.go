package main

import (
	"bufio"
	"net/http"
	"os"
	"strings"

	"github.com/sagernet/sing-box/common/srs"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

const URL = "https://static-file-global.353355.xyz/rules/cn-additional-list.txt"

var resourceMap = map[string]string{
	"cn-additional": "https://static-file-global.353355.xyz/rules/cn-additional-list.txt",
}

const HEADER = "{\"version\":3,\"rules\":[{\"domain_suffix\":["
const FOOTER = "]}]}"

func main() {
	for name, url := range resourceMap {
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
		ruleSet = strings.TrimSuffix(ruleSet, ",")
		ruleSet += FOOTER
		plainRuleSet, err := json.UnmarshalExtended[option.PlainRuleSetCompat]([]byte(ruleSet))
		if err != nil {
			panic(err)
		}
		writer, err := os.Create(name + ".srs")
		if err != nil {
			panic(err)
		}
		srs.Write(writer, plainRuleSet.Options, 3)
	}
}

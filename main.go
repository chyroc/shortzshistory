package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

func action(left []string) error {
	fmt.Println("开始 ~/.zsh_history 历史记录去重")
	fmt.Println("left:", left)

	content, err := readHistory()
	if err != nil {
		return err
	}
	historyList, err := splitHistory(content)
	if err != nil {
		return err
	}
	fmt.Println("原始记录数：", len(historyList))

	sort.Slice(historyList, func(i, j int) bool {
		return historyList[i].Ts < historyList[j].Ts
	})
	historyList = filterHistory(historyList, left)
	fmt.Println("去重后记录数：", len(historyList))

	return writeHistory(historyList)
}

func readHistory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	bs, err := os.ReadFile(home + "/.zsh_history")
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func writeHistory(historyList []*History) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	builder := strings.Builder{}
	for _, v := range historyList {
		builder.WriteString(fmt.Sprintf(": %s:%s;%s\n", v.Ts, v.Typ, v.Cmd))
	}
	return os.WriteFile(home+"/.zsh_history", []byte(builder.String()), 0644)
}

type History struct {
	Ts  string
	Typ string
	Cmd string
}

func splitHistory(content string) ([]*History, error) {
	res := make([]*History, 0)
	idx := 0
	list := strings.Split(content, "\n")
	for idx < len(list) {
		line := list[idx]
		for idx+1 < len(list) && list[idx+1] != "" && list[idx+1][0] != ':' {
			line += list[idx+1]
			idx++
		}
		idx++
		match := re.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}
		if len(match) != 4 {
			return nil, fmt.Errorf("invalid history line: %d, %s", len(match), line)
		} else {
			res = append(res, &History{
				Ts:  match[1],
				Typ: match[2],
				Cmd: match[3],
			})
		}
	}
	return res, nil
}

// 留下第一个，最后一个
func filterHistory(historyList []*History, left []string) []*History {
	done := map[string]bool{}
	res := make([]*History, 0)
	for _, v := range historyList {
		if done[v.Cmd] {
			continue
		}
		skip := false
		for _, vv := range left {
			if strings.HasPrefix(v.Cmd, vv+" ") || v.Cmd == vv {
				skip = true
			}
		}
		if skip {
			continue
		}
		done[v.Cmd] = true
		res = append(res, v)
	}

	return res

}

var re = regexp.MustCompile(`: (\d+):(\d+);(.*)`)

func main() {
	app := &cli.App{
		Name: "shortzshistory",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "left",
				Usage: "remove word which start with <left>",
			},
		},
		Action: func(c *cli.Context) error {
			left := c.StringSlice("left")
			return action(left)
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
		log.Fatalln(err)
	}
}

package main

import (
	"time"
	"strings"
	"bytes"
	"github.com/logrusorgru/aurora"
	"fmt"
	"log"
	"os"
	"bufio"
	"flag"
	"errors"
	"encoding/json"
	"net/http"
	"io/ioutil"

)

var au aurora.Aurora
func init() {
	au = aurora.NewAurora(true)
}

const (
	sendurl   = `https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=`
	get_token = `https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=`
)

type access_token struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
}

type send_msg_error struct {
	Errcode int    `json:"errcode`
	Errmsg  string `json:"errmsg"`
}

type send_msg struct {
	Touser  string            `json:"touser"`
	Toparty string            `json:"toparty"`
	Totag   string            `json:"totag"`
	Msgtype string            `json:"msgtype"`
	Agentid int               `json:"agentid"`
	Text    map[string]string `json:"text"`
	Safe    int               `json:"safe"`
}

var requestError = errors.New("request error,check url or network")


func Get_token(corpid, corpsecret string) (at access_token, err error) {
	resp, err := http.Get(get_token + corpid + "&corpsecret=" + corpsecret)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = requestError
		return
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(buf, &at)
	if at.Access_token == "" {
		err = errors.New("corpid or corpsecret error.")
	}
	return
}

func Send_msg(Access_token string, msgbody []byte) error {
	body := bytes.NewBuffer(msgbody)
	resp, err := http.Post(sendurl+Access_token, "application/json", body)
	if resp.StatusCode != 200 {
		return requestError
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var e send_msg_error
	err = json.Unmarshal(buf, &e)
	if err != nil {
		return err
	}
	if e.Errcode != 0 && e.Errmsg != "ok" {
		return errors.New(string(buf))
	}
	return nil
}



func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()

}


func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	w := bufio.NewWriter(file)

	for _, line := range lines {
		fmt.Fprintln(w, line)
	}

	return w.Flush()

}

func fetchUrlscan(page string, ok bool) ([]string, error) {
	resp, err := http.Get(
		fmt.Sprintf("https://dncapi.bqrank.net/api/v2/exchange/web-exchange?token=&page=%s&pagesize=100&sort_type=exrank&asc=1&isinnovation=1&type=all&area=&webp=1", page),
	)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	output := make([]string, 0)

	dec := json.NewDecoder(resp.Body)

	wrapper := struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"data"`
	}{}

	err = dec.Decode(&wrapper)
	if err != nil {
		return []string{}, err
	}

	for _, r := range wrapper.Results {
		if ok && len(output) >= 20 {
			break
		}
		output = append(output, r.Name)
	}


	return output, nil
}



func main() {
	agentid := flag.Int("i", 1000003, "？？？？")
	corpid := flag.String("p", "ww762f1f2b3b28bc7a", "？？？？")
	corpsecret := flag.String("s", "v0vSdFwTNOOQnJaLeXkWaAI8mdeS1tGwmG_IvpeLEKo", "？？？？？")
	flag.Parse()

	for true  {

		ok := make(map[string]bool)

		out1, _:= fetchUrlscan("1", false)
		out2, _:= fetchUrlscan("2", true)

		out := append(out1, out2...)

		info := ""
		var cstSh, _ = time.LoadLocation("Asia/Shanghai") //上海

		fmt.Println(au.Yellow(time.Now().In(cstSh).Format("2006-01-02 15:04:05")))


		if len(out) == 120 {
			fmt.Println(au.Green("API接口正常运行"))
			info = "API接口正常运行\n"
		} else {
			fmt.Println(au.Magenta("API接口异常， 请检查"))
			info = "API接口异常， 请检查\n"
			continue
		}

		lines, err := readLines("dvp.monitor.txt")
		if err != nil {
			log.Fatalf("readLines: %s", err)
		}

		for _, line := range lines {
			ok[line] = true
		}

		newdatas := []string {}
		for _, o := range out {
			if ok[o] {
				continue
			}
			newdatas = append(newdatas, o)
		}

		var content string



		if len(newdatas) == 0 {
			fmt.Println(au.Green("运行完毕， 没有找到新厂商"))
		} else {
			fmt.Println(au.Magenta("运行完毕， 找到新的厂商， 注意微信消息"))
			content = "运行完毕， 找到新的厂商，赶紧捡垃圾， 奥利给\n"
			for _, n := range newdatas{
				fmt.Println(au.Yellow("[!] " + n))
				content = content + "[!] " + n + "\n"
			}

			content = info + strings.TrimRight(content, "\n")


			//通知
			var meg send_msg = send_msg{Touser: "@all", Msgtype: "text", Agentid: *agentid, Text: map[string]string{"content": content}}

			token, err := Get_token(*corpid, *corpsecret)
			if err != nil {
				println(err.Error())
				return
			}
			buf, err := json.Marshal(meg)
			if err != nil {
				return
			}
			err = json.Unmarshal(buf, &meg)
			if err != nil {
				println(err)
				return
			}
			err = Send_msg(token.Access_token, buf)
			if err != nil {
				println(err.Error())
			}

		}


		if err := writeLines(out, "dvp.monitor.txt"); err != nil {
			log.Fatalf("writeLines: %s", err)
		}

	}


}

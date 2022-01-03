package main

import (
        "encoding/json"
        "io/ioutil"
        "bufio"
        "log"
        "sort"
        "strings"
        "os"
        "net/http"
)

type server struct {
        Util int `json:"util"`
        Ip string `json:"ip"`
}

func getLowestLoad() []server {
        resp, err := http.Get("https://subdomain.domain.com/")

        if err != nil {
                log.Fatal(err)
        }

        defer resp.Body.Close()

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                log.Fatal(err)
        }

        var serverList []server
        err = json.Unmarshal(body, &serverList)
        if err != nil {
                log.Fatal(err)
        }

        sort.Slice(serverList, func(i, j int) bool {
                return serverList[i].Util < serverList[j].Util
        })
        return serverList
}

func main() {
        scanner := bufio.NewScanner(os.Stdin)
        writer := bufio.NewWriter(os.Stdout)

        scanner.Scan()
        writer.Write([]byte("OK\tMy Backend\n"))
        writer.Flush()

        for scanner.Scan() {
                strs := strings.TrimRight(scanner.Text(), "\n")
                str := strings.Split(strs, "\t")
                if len(str) < 6 {
                        writer.Write([]byte("LOG\tPowerDNS sent unparseable line\n"))
                        writer.Flush()
                        continue
                }

                qname, qclass, qtype := str[1], str[2], str[3]

                if (qtype == "SOA" || qtype == "ANY") && qname == "vpntask3.domain.com" {
                        writer.Write([]byte("DATA\t" + qname + "\t" + qclass +
			"\tSOA\t3600\t-1\tns1.vpntask3.domain.com ahu.vpntask3.domain.com 2021120100 1800 3600 604800 3600\n"))
                }

                if (qtype == "NS" || qtype == "ANY") && qname == "vpntask3.domain.com" {
                        writer.Write([]byte("DATA\t" + qname + "\t" + qclass + "\tNS\t3600\t-1\tns1.vpntask3.domain.com\n"))
                        writer.Write([]byte("DATA\t" + qname + "\t" + qclass + "\tNS\t3600\t-1\tns2.vpntask3.domain.com\n"))
                } else if (qtype == "A" || qtype == "ANY") && qname == "server.vpntask3.domain.com" {
                	servers := getLowestLoad()
                        for i := 0; i < 3; i++ {
                                writer.Write([]byte("DATA\t" + qname + "\t" + qclass + "\tA\t1\t-1\t" + servers[i].Ip +"\n"))
                        }
                }

                writer.Write([]byte("END\n"))
                writer.Flush()
        }
}

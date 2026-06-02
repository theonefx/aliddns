package utils

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/OpenIoTHub/aliddns/config"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

//u4="http://ipv4.ident.me http://ipv4.icanhazip.com http://nsupdate.info/myip http://whatismyip.akamai.com http://ipv4.myip.dk/api/info/IPv4Address http://checkip4.spdyn.de http://v4.ipv6-test.com/api/myip.php http://checkip.amazonaws.com http://ipinfo.io/ip http://bot.whatismyipaddress.com http://ipv4.ident.me http://ipv4.icanhazip.com http://nsupdate.info/myip http://whatismyip.akamai.com http://ipv4.myip.dk/api/info/IPv4Address http://checkip4.spdyn.de http://v4.ipv6-test.com/api/myip.php http://checkip.amazonaws.com http://ipinfo.io/ip http://bot.whatismyipaddress.com"
//u6="http://ipv6.ident.me http://ipv6.icanhazip.com http://ipv6.ident.me http://ipv6.icanhazip.com http://ipv6.yunohost.org http://v6.ipv6-test.com/api/myip.php http://ipv6.ident.me http://ipv6.icanhazip.com http://ipv6.yunohost.org http://v6.ipv6-test.com/api/myip.php http://ipv6.ident.me http://ipv6.icanhazip.com http://ipv6.ident.me http://ipv6.icanhazip.com http://ipv6.yunohost.org http://v6.ipv6-test.com/api/myip.php http://ipv6.ident.me http://ipv6.icanhazip.com http://ipv6.yunohost.org http://v6.ipv6-test.com/api/myip.php"

var Ipv4APIUrls = []string{
	"http://whatismyip.akamai.com",
	"http://v4.ipv6-test.com/api/myip.php",
	"http://checkip.amazonaws.com",
	"api.ipify.org",
	"canhazip.com",
	"ident.me",
	"whatismyip.akamai.com",
	"myip.dnsomatic.com",
	"http://members.3322.org/dyndns/getip",
	"http://ifconfig.me/ip",
	"http://ip.3322.net",
	"https://myexternalip.com/raw",
	"http://ipv4.ident.me",
	"http://ipv4.icanhazip.com",
	"http://nsupdate.info/myip",
	"http://ipv4.myip.dk/api/info/IPv4Address",
	"http://checkip4.spdyn.de",
	"http://ipinfo.io/ip",
}
var Ipv6APIUrls = []string{
	"http://v6.ipv6-test.com/api/myip.php",
	"http://bbs6.ustc.edu.cn/cgi-bin/myip",
	"http://ipv6.ident.me",
	"http://ipv6.icanhazip.com",
	"http://ipv6.yunohost.org",
}

func GetMyPublicIpv4() string {
	interfaceName := strings.TrimSpace(config.ConfigModel.Ipv4InterfaceName)
	if interfaceName != "" {
		ipv4, err := GetIPByInterface(interfaceName, "ipv4")
		if err != nil {
			log.Printf("get ipv4 from interface err：%s", err)
			return ""
		}
		log.Printf("got ipv4 addr from interface %s: %s", interfaceName, ipv4)
		return ipv4
	}
	urls := Ipv4APIUrls
	if apiURL := strings.TrimSpace(config.ConfigModel.Ipv4ApiUrl); apiURL != "" {
		urls = append([]string{apiURL}, Ipv4APIUrls...)
	}
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("get public ipv4 err：%s", err)
			continue
		}
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("get public ipv4 err：%s", err)
			_ = resp.Body.Close()
			continue
		}
		ipv4 := strings.TrimSpace(string(bytes))
		ip := net.ParseIP(ipv4)
		if ip != nil {
			log.Println("got ipv4 addr:", ip.String())
			_ = resp.Body.Close()
			return ip.String()
		}
		_ = resp.Body.Close()
	}
	return ""
}

func GetMyPublicIpv6() string {
	interfaceName := strings.TrimSpace(config.ConfigModel.Ipv6InterfaceName)
	if interfaceName != "" {
		ipv6, err := GetIPByInterface(interfaceName, "ipv6")
		if err != nil {
			log.Printf("get ipv6 from interface err：%s", err)
			return ""
		}
		log.Printf("got ipv6 addr from interface %s: %s", interfaceName, ipv6)
		return ipv6
	}
	urls := Ipv6APIUrls
	if apiURL := strings.TrimSpace(config.ConfigModel.Ipv6ApiUrl); apiURL != "" {
		urls = append([]string{apiURL}, Ipv6APIUrls...)
	}
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("get public ipv6 err：%s", err)
			continue
		}
		// 读取 IPv6
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("get public ipv6 err：%s", err)
			_ = resp.Body.Close()
			continue
		}
		// 删除 document.write(xxx) (如有)
		tmp := strings.Replace(string(bytes), "document.write('", "", -1)
		tmp = strings.Replace(tmp, "');", "", -1)
		ipv6 := strings.TrimSpace(tmp)
		ip := net.ParseIP(ipv6)
		if ip != nil {
			log.Println("got ipv6 addr:", ip.String())
			_ = resp.Body.Close()
			return ip.String()
		}
		_ = resp.Body.Close()
	}
	return ""
}

func GetIPByInterface(interfaceName, protocol string) (string, error) {
	networkInterface, err := net.InterfaceByName(strings.TrimSpace(interfaceName))
	if err != nil {
		return "", err
	}
	if networkInterface.Flags&net.FlagUp == 0 {
		return "", fmt.Errorf("network interface %q is down", networkInterface.Name)
	}
	addrs, err := networkInterface.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ip := getIPFromInterfaceAddr(addr)
		if ip == nil || !isUsableInterfaceIP(ip) {
			continue
		}
		switch protocol {
		case "ipv4":
			ipv4 := ip.To4()
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		case "ipv6":
			if ip.To4() == nil && ip.To16() != nil {
				return ip.String(), nil
			}
		default:
			return "", fmt.Errorf("unsupported protocol %q", protocol)
		}
	}
	return "", fmt.Errorf("no usable %s address found on interface %q", protocol, networkInterface.Name)
}

func getIPFromInterfaceAddr(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPAddr:
		return v.IP
	case *net.IPNet:
		return v.IP
	default:
		return nil
	}
}

func isUsableInterfaceIP(ip net.IP) bool {
	return ip.IsGlobalUnicast() &&
		!ip.IsLoopback() &&
		!ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() &&
		!ip.IsMulticast() &&
		!ip.IsUnspecified()
}

func GetSubDomains(mainDomian string) (*alidns.DescribeDomainRecordsResponse, error) {
	client, err := GetAliYunClient()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = mainDomian
	return client.DescribeDomainRecords(request)
}

func UpdateSubDomain(subDomain *alidns.Record) error {
	client, err := GetAliYunClient()
	if err != nil {
		log.Println(err)
		return err
	}
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = subDomain.RecordId
	request.RR = subDomain.RR
	request.Type = subDomain.Type
	request.Value = subDomain.Value
	request.TTL = requests.NewInteger64(subDomain.TTL)

	_, err = client.UpdateDomainRecord(request)
	if err != nil {
		log.Print("UpdateDomainRecord:", err)
		return err
	}
	return nil
}

func AddSubDomainRecord(subDomain *alidns.Record) error {
	client, err := GetAliYunClient()
	if err != nil {
		log.Println(err)
		return err
	}

	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.DomainName = subDomain.DomainName
	request.RR = subDomain.RR
	request.Type = subDomain.Type
	request.Value = subDomain.Value
	request.TTL = requests.NewInteger64(subDomain.TTL)

	_, err = client.AddDomainRecord(request)
	if err != nil {
		log.Print("AddSubDomainRecord:", err)
		return err
	}
	return nil
}

func GetAliYunClient() (*alidns.Client, error) {
	return alidns.NewClientWithAccessKey("cn-hangzhou", config.ConfigModel.AccessId, config.ConfigModel.AccessKey)
}

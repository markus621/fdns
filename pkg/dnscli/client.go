package dnscli

import (
	"math/rand"
	"strings"
	"time"

	"github.com/dewep-online/fdns/pkg/utils"
	"github.com/deweppro/go-logger"
	"github.com/miekg/dns"
)

type Client struct {
	ips []string
	cli *dns.Client
}

func New(c *Config) *Client {
	return &Client{
		cli: &dns.Client{
			Net:          "",
			ReadTimeout:  time.Second * 5,
			WriteTimeout: time.Second * 5,
		},
		ips: append([]string{}, c.DNS...),
	}
}

func (o *Client) ExchangeRandomDNS(msg *dns.Msg) (*dns.Msg, error) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(o.ips), func(i, j int) { o.ips[i], o.ips[j] = o.ips[j], o.ips[i] })

	return o.Exchange(msg, o.ips)
}

func (o *Client) Exchange(msg *dns.Msg, addrs []string) (resp *dns.Msg, err error) {
	var mq, mr []string

	for _, ns := range addrs {
		resp, _, err = o.cli.Exchange(msg, ns)
		if err != nil {
			continue
		}

		for _, q := range msg.Question {
			mq = append(mq, q.String())
		}

		for _, a := range resp.Answer {
			mr = append(mr, a.String())
		}

		logger.Infof("reverse: NS: %s QUERY: %s RESPONSE: %s",
			ns, strings.Join(mq, ","), strings.Join(mr, ","))

		break
	}

	if err != nil {
		logger.Infof("reverse: NS: %s QUERY: %s ERROR: %s",
			addrs, strings.Join(mq, ","), err.Error())
	}

	return
}

func (o *Client) Up() error {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn("example.com."), dns.TypeNS)

	list := utils.ValidateDNSs(o.ips)
	o.ips = make([]string, 0)

	for _, ip := range list {
		if _, _, err := o.cli.Exchange(msg, ip); err == nil {
			logger.Infof("add dns: %s", ip)
			o.ips = append(o.ips, ip)
		}
	}

	if len(o.ips) == 0 {
		return utils.ErrEmptyDNSList
	}
	return nil
}

func (o *Client) Down() error {
	return nil
}

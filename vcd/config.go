package vcd

import (
	"fmt"
	"net/url"

	govcd "github.com/vmware/go-vcloud-director/govcd"
)

type Config struct {
	User            string
	Password        string
	Org             string
	Href            string
	MaxRetryTimeout int
	InsecureFlag    bool
}

type VCDClient struct {
	*govcd.VCDClient
	MaxRetryTimeout int
	InsecureFlag    bool
}

func (c *Config) Client() (*VCDClient, error) {
	u, err := url.ParseRequestURI(c.Href)
	if err != nil {
		return nil, fmt.Errorf("Something went wrong: %s", err)
	}

	vcdclient := &VCDClient{
		govcd.NewVCDClient(*u, c.InsecureFlag),
		c.MaxRetryTimeout, c.InsecureFlag}
	err = vcdclient.Authenticate(c.User, c.Password, c.Org)
	if err != nil {
		return nil, fmt.Errorf("Something went wrong: %s", err)
	}
	//vcdclient.Org = org
	//vcdclient.OrgVdc = vdc
	return vcdclient, nil
}

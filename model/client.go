package model

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

type ClientType int

const (
	CT_NONE ClientType = iota
	CT_ADMIN
	CT_INNER
	CT_OUTTER
)

type Client struct {
	BaseModel
	ClientUser  string
	Secret      string
	RedirectUri string `gorm:"size(2000)"`
	Type        ClientType
	UserData    string "type:text"
}

const (
	C_PREFIX_APPLICATION = "mna_"
	C_PREFIX_WEB         = "mnw_"
)

func init() {
	AddModel("Client", Client{})
}

func NewClient() *Client {
	cli := &Client{}
	return cli
}

func (c *Client) FirstByID(id string) {
	FirstById(c, id)
}

func (c *Client) FindAll() *[]Client {
	var cls []Client
	Gorm().Find(&cls)
	return &cls
}

func (c *Client) GetSecret() string {
	return c.Secret
}

//func (c *Client) GetUri() string {
//	return c.RedirectUri
//}

func (c *Client) GetClient() string {
	return c.ClientUser
}

func (c *Client) String() string {
	b, e := json.Marshal(*c)
	if e != nil {
		return ""
	}
	return string(b)
}

func (c *Client) IsNull() bool {
	return uuid.Equal(c.ID, uuid.Nil)
}

func (c *Client) ParseJson(s string) {
	json.Unmarshal([]byte(s), c)
}
func (c *Client) CheckRedirectUri(redirect_uri string) (flag bool) {
	var uri []string
	flag = false

	err := json.Unmarshal([]byte(c.RedirectUri), &uri)
	if err == nil {
		for _, v := range uri {
			if v == redirect_uri {
				flag = true
				break
			}
		}
	}
	return flag
}

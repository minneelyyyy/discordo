package keyring

import (
	"encoding/json"

	"github.com/ayn2op/discordo/internal/consts"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/zalando/go-keyring"
)

const (
	keyringService = consts.Name
	keyringUser    = "token"
)

type AccountInfo struct {
	Id    discord.UserID `json:"id"`
	Token string         `json:"token"`
}

func GetAccounts() ([]AccountInfo, error) {
	data, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		return nil, err
	}

	accounts := make([]AccountInfo, 0)
	if err := json.Unmarshal([]byte(data), &accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}

func GetToken(accounts []AccountInfo, id discord.UserID) string {
	for _, account := range accounts {
		if account.Id == id {
			return account.Token
		}
	}

	return ""
}

func setAccounts(accounts []AccountInfo) error {
	data, err := json.Marshal(accounts)
	if err != nil {
		return err
	}

	return keyring.Set(keyringService, keyringUser, string(data))
}

func SetTokenWithAccounts(accounts []AccountInfo, id discord.UserID, s string) error {
	for _, account := range accounts {
		if account.Id == id {
			account.Token = s
			break
		}
	}

	return setAccounts(accounts)
}

func SetToken(id discord.UserID, s string) error {
	accounts, err := GetAccounts()
	if err != nil {
		return setAccounts([]AccountInfo{{Id: id, Token: s}})
	}

	return SetTokenWithAccounts(accounts, id, s)
}

func DeleteToken() error {
	return keyring.Delete(keyringService, keyringUser)
}

package vhackapi

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type VHackOSClient struct {
	Client         *http.Client
	ClientSettings *Settings
	SettingsPath   string
}

type AccountSettings struct {
	Username  string      `json:"username"`
	Alias     string      `json:"alias"`
	Password  string      `json:"password"`
	Email     string      `json:"email"`
	Token     string      `json:"token"`
	UserAgent string      `json:"useragent"`
	Uid       string      `json:"uid"`
	Firebase  string      `json:"firebase"`
	LastLogin string      `json:"last_login"`
	Language  string      `json:"language"`
	Channel   string      `json:"channel"`
	MinSleep  int         `json:"min_sleep"`
	MaxSleep  int         `json:"max_sleep"`
	Odds      TalkingOdds `json:"talking_odds"`
	Timezone  string      `json:"timezone"`
}

type TalkingOdds struct {
	Mentioned int `json:"mentioned"`
	Random    int `json:"random"`
}

type Settings struct {
	Host        string          `json:"host"`
	ChatAddress string          `json:"chat_address"`
	Version     string          `json:"version"`
	Account     AccountSettings `json:"account"`
}

type Account struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	Uid      string `json:"uid"`
	Email    string `json:"email"`
	Exp      int    `json:"exp"`
	ExpReq   int    `json:"expreq"`
	IP       string `json:"ip"`
	Money    int    `json:"money"`
	Netcoins int    `json:"netcoins"`
	Inet     string `json:"inet"`
	Level    int    `json:"level"`
}

func stringToInt(str string) int {
	res, _ := strconv.Atoi(str)
	return res
}

func (client VHackOSClient) SaveSettings() error {
	settingsJson, _ := json.Marshal(client.ClientSettings)
	err := ioutil.WriteFile(client.SettingsPath, settingsJson, 0644)
	return err
}

func GetClient(path string) VHackOSClient {
	var result VHackOSClient
	f, errOpen := os.Open(path)
	if errOpen == nil {
		bytes, _ := ioutil.ReadAll(f)
		settings := new(Settings)
		_ = json.Unmarshal(bytes, &settings)
		result = VHackOSClient{
			Client: &http.Client{
				Timeout: time.Second * 10,
				Transport: &http.Transport{
					TLSHandshakeTimeout: 5 * time.Second,
					DisableCompression:  false,
					DisableKeepAlives:   false,
				},
			},
			ClientSettings: settings,
			SettingsPath:   path,
		}
	} else {
		log.Fatal(errOpen)
		result = VHackOSClient{}
	}
	return result
}

func (client *VHackOSClient) makeRequest(destination string, data map[string]interface{}) *http.Request {
	switch destination {
	case "update":
		if len(client.ClientSettings.Account.Firebase) > 0 {
			data["firebase"] = client.ClientSettings.Account.Firebase
		}
		data["uid"] = client.ClientSettings.Account.Uid
		data["accesstoken"] = client.ClientSettings.Account.Token
		data["lang"] = client.ClientSettings.Account.Language
		break
	case "login":
		break
	case "register":
		break
	case "requestpass":
		break
	default:
		data["uid"] = client.ClientSettings.Account.Uid
		data["accesstoken"] = client.ClientSettings.Account.Token
		data["lang"] = client.ClientSettings.Account.Language
		break
	}
	jsonData, _ := json.Marshal(data)
	result, _ := http.NewRequest("GET", client.ClientSettings.Host+client.ClientSettings.Version+"/"+destination+".php", nil)
	result.Header.Add("User-Agent", client.ClientSettings.Account.UserAgent)
	query := result.URL.Query()
	user, pass := client.makeParams(string(jsonData))
	query.Add("user", user)
	query.Add("pass", pass)
	result.URL.RawQuery = query.Encode()
	return result
}

func (client *VHackOSClient) makeParams(data string) (string, string) {
	var user string
	var pass string
	user = base64.URLEncoding.EncodeToString([]byte(data))
	hash := MD5hash(data)
	pass = MD5hash(data + data + hash)
	return user, pass
}

func MD5hash(text string) string {
	algorithm := md5.New()
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}

func (client *VHackOSClient) Login() (Account, error) {
	fmt.Println("Logging in")
	var result Account
	var errResult error
	if len(client.ClientSettings.Account.Username) != 0 {
		if len(client.ClientSettings.Account.Password) != 0 {
			params := make(map[string]interface{})
			// interface{} allows you to use all types of variables there, may not be needed but i need to check for later
			params["username"] = client.ClientSettings.Account.Username
			params["password"] = MD5hash(client.ClientSettings.Account.Password)
			resp, err := client.Client.Do(client.makeRequest("login", params))
			errResult = err
			bodyBytes, err2 := ioutil.ReadAll(resp.Body)
			errResult = err2
			response := make(map[string]string)
			_ = json.Unmarshal(bodyBytes, &response)
			switch response["result"] {
			case "0":
				client.ClientSettings.Account.Token = response["accesstoken"]
				client.ClientSettings.Account.Uid = response["uid"]
				client.ClientSettings.Account.LastLogin = time.Now().Format(time.RFC850)
				_ = client.SaveSettings()
				result = Account{
					Username: response["username"],
					Token:    response["accesstoken"],
					Uid:      response["uid"],
					Email:    response["email"],
					IP:       response["ipaddress"],
					Inet:     response["inet"],
					Level:    stringToInt(response["level"]),
					Money:    stringToInt(response["money"]),
					Netcoins: stringToInt(response["netcoins"]),
					Exp:      stringToInt(response["exp"]),
					ExpReq:   stringToInt(response["expreq"]),
				}
				errResult = nil
			case "1":
				errResult = errors.New("account banned - untested")
			case "2":
				errResult = errors.New("username or password not correct")
			}
		} else {
			errResult = errors.New("password not set in the settings file")
		}
	} else {
		errResult = errors.New("username not set in the settings file")
	}
	return result, errResult
}

func (client *VHackOSClient) Init() bool {
	result := false
	if len(client.ClientSettings.Account.Token) > 0 && len(client.ClientSettings.Account.Uid) > 0 {
		if client.Update() {
			result = true
		} else {
			_, err := client.Login()
			if err == nil {
				result = true
			} else {
				fmt.Println(err)
			}
		}
	} else {
		if len(client.ClientSettings.Account.Username) > 0 && len(client.ClientSettings.Account.Password) > 0 {
			_, err := client.Login()
			if err == nil {
				result = true
			} else {
				fmt.Println(err)
			}
		}
	}
	return result
}

func (client *VHackOSClient) Update() bool {
	result := false
	if len(client.ClientSettings.Account.Username) != 0 {
		if len(client.ClientSettings.Account.Password) != 0 {
			params := make(map[string]interface{})
			// interface{} allows you to use all types of variables there, may not be needed but i need to check for later
			params["lastread"] = "0"
			params["notify"] = "1"
			resp, err := client.Client.Do(client.makeRequest("update", params))
			if err != nil {
				log.Fatalln(err)
			}
			bodyBytes, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				log.Fatalln(err2)
			}
			response := make(map[string]string)
			_ = json.Unmarshal(bodyBytes, &response)
			switch response["result"] {
			case "0":
				result = true
			}
		}
	}
	return result
}

func (client *VHackOSClient) PrivateMessageUser(userId, message string) bool {
	result := false
	if len(client.ClientSettings.Account.Username) != 0 {
		if len(client.ClientSettings.Account.Password) != 0 {
			params := make(map[string]interface{})
			// interface{} allows you to use all types of variables there, may not be needed but i need to check for later
			params["action"] = "300"
			params["partner_id"] = userId
			params["message"] = message
			resp, err := client.Client.Do(client.makeRequest("messenger", params))
			if err != nil {
				log.Fatalln(err)
			}
			bodyBytes, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				log.Fatalln(err2)
			}
			response := make(map[string]string)
			_ = json.Unmarshal(bodyBytes, &response)
			switch response["result"] {
			case "0":
				result = true
			}
		}
	}
	return result
}

func (client *VHackOSClient) Register() bool {
	result := false
	if len(client.ClientSettings.Account.Username) != 0 {
		if len(client.ClientSettings.Account.Password) != 0 {
			params := make(map[string]interface{})
			// interface{} allows you to use all types of variables there, may not be needed but i need to check for later
			params["username"] = client.ClientSettings.Account.Username
			params["email"] = client.ClientSettings.Account.Email
			params["password"] = MD5hash(client.ClientSettings.Account.Password)
			resp, err := client.Client.Do(client.makeRequest("register", params))
			if err != nil {
				log.Fatalln(err)
			}
			bodyBytes, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				log.Fatalln(err2)
			}
			fmt.Println(resp.Status)
			fmt.Println(string(bodyBytes))
			result = true
		}
	}
	return result
}

func (client *VHackOSClient) BaseFunc() bool {
	result := false
	if len(client.ClientSettings.Account.Username) != 0 {
		if len(client.ClientSettings.Account.Password) != 0 {
			params := make(map[string]interface{})
			// interface{} allows you to use all types of variables there, may not be needed but i need to check for later
			params["username"] = client.ClientSettings.Account.Username
			params["email"] = client.ClientSettings.Account.Email
			params["password"] = MD5hash(client.ClientSettings.Account.Password)
			resp, err := client.Client.Do(client.makeRequest("register", params))
			if err != nil {
				log.Fatalln(err)
			}
			bodyBytes, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				log.Fatalln(err2)
			}
			fmt.Println(resp.Status)
			fmt.Println(string(bodyBytes))
			result = true
		}
	}
	return result
}

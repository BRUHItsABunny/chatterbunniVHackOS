package clevergo

import (
	"chatterBunniVhackOS/gotranslate"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CleverClient struct {
	Client      *http.Client
	AIS         string
	SessionId   string
	PrevRef     string
	LineRef     string
	Messages    []string
	Language    string
	GoogleTrans gotranslate.GoTranslateClient
}

type XMLResponse struct {
	XMLName xml.Name      `xml:"webservicexml"`
	Session CleverSession `xml:"session"`
}

type CleverSession struct {
	XMLName        xml.Name `xml:"session"`
	Id             string   `xml:"sesionid"`
	Response       string   `xml:"response"`
	VText2         string   `xml:"vtext2"`
	VText3         string   `xml:"vtext3"`
	VText4         string   `xml:"vtext4"`
	VText5         string   `xml:"vtext5"`
	VText6         string   `xml:"vtext6"`
	VText7         string   `xml:"vtext7"`
	VText8         string   `xml:"vtext8"`
	Reaction       string   `xml:"reaction"`
	ReactionDegree string   `xml:"reactionDegree"`
	Emotion        string   `xml:"emotion"`
	EmotionDegree  string   `xml:"emotionDegree"`
	AIS            string   `xml:"rpsais"`
	PrevRef        string   `xml:"prevref"`
	LineRef        string   `xml:"lineRef"`
}

func GetClient() CleverClient {
	return CleverClient{
		Client: &http.Client{
			Timeout: time.Second * time.Duration(20),
			Transport: &http.Transport{
				TLSHandshakeTimeout: time.Duration(5) * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
			},
		},
		Messages:    []string{"", "", "", "", "", "", ""},
		GoogleTrans: gotranslate.GetClient(),
	}
}

func (client CleverClient) makeRequest(form url.Values) *http.Request {
	var httpUrl string
	if len(client.AIS) == 0 {
		httpUrl = "https://app.cleverbot.com/webservicexml" + "_ais_" + client.AIS
	} else {
		httpUrl = "https://app.cleverbot.com/webservicexml"
	}
	req, _ := http.NewRequest("POST", httpUrl, strings.NewReader(form.Encode()))
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func (client *CleverClient) FlushHistory() {
	client.Messages = []string{"", "", "", "", "", "", ""}
}

func (client *CleverClient) Ask(sentence string) (CleverSession, error) {
	var result XMLResponse
	var errResult error
	form := url.Values{}
	form.Add("stimulus", sentence)
	form.Add("sessionid", client.SessionId)
	form.Add("vtext2", client.Messages[len(client.Messages)-1])
	form.Add("vtext3", client.Messages[len(client.Messages)-2])
	form.Add("vtext4", client.Messages[len(client.Messages)-3])
	form.Add("vtext5", client.Messages[len(client.Messages)-4])
	form.Add("vtext6", client.Messages[len(client.Messages)-5])
	form.Add("vtext7", client.Messages[len(client.Messages)-6])
	form.Add("vtext8", client.Messages[len(client.Messages)-7])
	form.Add("prevref", client.PrevRef)
	form.Add("lineRef", client.LineRef)
	form.Add("icognoCheck", appSecret)
	form.Add("icognoId", appId)
	resp, err := client.Client.Do(client.makeRequest(form))
	if err != nil {
		errResult = err
	}
	if resp != nil {
		bodyBytes, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			errResult = err2
		}
		_ = xml.Unmarshal(bodyBytes, &result)
		client.AIS = result.Session.AIS
		client.SessionId = result.Session.Id
		client.PrevRef = result.Session.PrevRef
		client.LineRef = result.Session.LineRef
		client.Messages = append(client.Messages, sentence)
		client.Messages = append(client.Messages, result.Session.Response)
		result.Session.Response = strings.Replace(result.Session.Response, "Clever Bot", "BUNNY", 1)
		result.Session.Response = strings.Replace(result.Session.Response, "clever bot", "bunny", 1)
		result.Session.Response = strings.Replace(result.Session.Response, "Robot", "Bunny", 1)
		result.Session.Response = strings.Replace(result.Session.Response, "robot", "bunny", 1)
		result.Session.Response = strings.Replace(result.Session.Response, "clever", "smart", 1)
		result.Session.Response = strings.Replace(result.Session.Response, "bot", "bunny", 1)
		result.Session.Response = strings.Replace(result.Session.Response, "Clever", "Smart", 1)
		result.Session.Response = strings.Replace(result.Session.Response, "Bot", "Bunny", 1)
		if len(client.Language) > 0 {
			res, _ := client.GoogleTrans.Translate(result.Session.Response, client.Language)
			result.Session.Response = res.GetTranslation()
		}
	}
	return result.Session, errResult
}

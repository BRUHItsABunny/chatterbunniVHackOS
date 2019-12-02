package gotranslate

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	userAgent       = "GoogleTranslate/5.25.1.RC06.220517201 (Linux; U; Android 8.0.0; ONEPLUS A3003)"
	translateUrl    = "https://translate.google.com/translate_a/single"
	translateCookie = "NID=154=GPfWrDUa-fXPqmXD7L9BAX25EsnBoiF9DVjVQczKf2SZOshvld5LYBrgvJRB5ma_SHkv5Xea7OuCmIaqEEoGHpRk29AG3pycDnlBj6ezSlexz7FXstPXofbS_uyq3J1VyVAgjWsy9U2iwSv5K-8dHUDH0MXaWYJCUQqTAT8temk"
)

type GoTranslateClient struct {
	Client *http.Client
}

type Response struct {
	Sentences []struct {
		Translated string `json:"trans"`
		Original   string `json:"orig"`
	} `json:"sentences"`
	Source     string `json:"src"`
	Confidence string `json:"confidence"`
}

func GetClient() GoTranslateClient {
	return GoTranslateClient{
		Client: &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				TLSHandshakeTimeout: 5 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
			},
		},
	}
}

func (client GoTranslateClient) Translate(text, language string) (Response, error) {
	var (
		errTrans error
		result   Response
	)
	req, err := http.NewRequest("GET", translateUrl, nil)
	if err == nil {
		errTrans = err
	}
	req.Header.Add("User-Agent", userAgent)
	// req.Header.Add("Cookie", translateCookie)
	query := req.URL.Query()
	query.Add("dj", "1")
	query.Add("q", text)
	query.Add("sl", "auto")
	query.Add("tl", language)
	query.Add("hl", "en")
	query.Add("ie", "UTF-8")
	query.Add("oe", "UTF-8")
	query.Add("client", "at")
	query.Add("dt", "t")
	query.Add("dt", "ld")
	query.Add("dt", "qca")
	query.Add("dt", "rm")
	query.Add("dt", "bd")
	query.Add("dt", "md")
	query.Add("dt", "ss")
	query.Add("dt", "ex")
	query.Add("otf", "2")
	req.URL.RawQuery = query.Encode()

	resp, err2 := client.Client.Do(req)
	errTrans = err2
	bodyBytes, err3 := ioutil.ReadAll(resp.Body)
	errTrans = err3
	_ = json.Unmarshal(bodyBytes, &result)
	return result, errTrans
}

func (res Response) GetTranslation() string {
	result := ""
	for _, it := range res.Sentences {
		result = result + it.Translated
	}
	return result
}

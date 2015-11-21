package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	sourceRoot := os.Args[1]
	if sourceRoot == "" {
		fmt.Fprintf(os.Stderr, "expected path to build sources as first argument")
		os.Exit(1)
	}

	var indata struct {
		Source struct {
			SMS struct {
				AccountSID  string
				AccessToken string
			}
			From string
			To   string
		}
		Params struct {
			Body string
		}
	}

	inbytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(inbytes, &indata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing input as JSON: %s", err)
		os.Exit(1)
	}

	if indata.Source.SMS.AccountSID == "" {
		fmt.Fprintf(os.Stderr, `missing required field "source.sms.account_sid"`)
		os.Exit(1)
	}

	if indata.Source.SMS.AccessToken == "" {
		fmt.Fprintf(os.Stderr, `missing required field "source.sms.access_token"`)
		os.Exit(1)
	}

	if indata.Source.From == "" {
		fmt.Fprintf(os.Stderr, `missing required field "source.from"`)
		os.Exit(1)
	}

	if len(indata.Source.To) == 0 {
		fmt.Fprintf(os.Stderr, `missing required field "source.to"`)
		os.Exit(1)
	}

	if indata.Params.Body == "" {
		fmt.Fprintf(os.Stderr, `missing required field "params.body"`)
		os.Exit(1)
	}

	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + indata.Source.SMS.AccountSID + "/Messages.json"

	v := url.Values{}
	v.Set("To", indata.Source.SMS.To)
	v.Set("From", indata.Source.SMS.From)
	v.Set("Body", "hello from concourse!")

	rb := *strings.NewReader(v.Encode())

	client := &http.Client{}

	req, err := http.NewRequest("POST", urlStr, &rb)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	req.SetBasicAuth(indata.Source.SMS.AccountSID, indata.Source.SMS.AccessToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	type MetadataItem struct {
		Name  string
		Value string
	}
	var outdata struct {
		Version struct {
			Time time.Time
		} `json:"version"`
		Metadata []MetadataItem
	}
	outdata.Version.Time = time.Now().UTC()
	outdata.Metadata = []MetadataItem{
		{Name: "sms", Value: resp},
	}

	outbytes, err := json.Marshal(outdata)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", []byte(outbytes))
}

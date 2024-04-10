package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
)

func submitUrl(url url.URL, server *url.URL, code *string, plain bool) {
	var jsonData []byte
	if code != nil {
		jsonData = []byte(`{
			"link": "` + url.String() + `",
			"requestedCode": "` + *code + `"
		}`)
	} else {
		jsonData = []byte(`{
			"link": "` + url.String() + `"
		}`)
	}
	resp, err := http.Post(server.String(), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic("Error sending post request to server!")
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		if plain {
			fmt.Println(strings.TrimRight(server.String(), "/") + "/" + string(body))
		} else {
			fmt.Println("Successfully created shortlink pointing to " + url.String() + " from " + strings.TrimRight(server.String(), "/") + "/" + string(body))
		}
	} else {
		panic("Error " + resp.Status + " while shortening " + url.String() + ": \n" + string(body))
	}
	resp.Body.Close()
}

func main() {
	homeDir, _ := os.UserHomeDir()
	serverURLdata, err := os.ReadFile(homeDir + string(os.PathSeparator) + ".shortlinks_server")
	var serverURL *url.URL
	if err == nil {
		serverURL, err = url.Parse(string(serverURLdata))
		if err != nil {
			panic("Error reading saved server url. Delete `.shortlinks_service` from your home directory and try again.")
		}
	}

	var options struct {
		URLs   []url.URL `arg:"positional"`
		Plain  bool      `arg:"-p,--plain" help:"only outputs created shortlinks seperated by newlines"`
		Code   *string   `arg:"-c,--request-code" placeholder:"CODE" help:"path to be requested from server"`
		Server *url.URL  `arg:"-s,--set-server" placeholder:"URL" help:"server url to use for this conversion. When ran with this option alone, it sets the default server."`
	}
	p := arg.MustParse(&options)

	if options.Code == nil && options.Server == nil && len(options.URLs) == 0 {
		p.WriteHelp(os.Stdout)
	}

	if options.Code != nil && len(options.URLs) != 1 {
		p.Fail("Request-Code option only valid when shortening a single url.")
	} else if options.Server != nil && len(options.URLs) == 0 {
		os.Remove(homeDir + string(os.PathSeparator) + ".shortlinks_server")
		os.WriteFile(homeDir+string(os.PathSeparator)+".shortlinks_server", []byte(options.Server.String()), 0666)
	} else {
		if options.Server != nil {
			serverURL = options.Server
		}
		if serverURL == nil {
			p.Fail("Shortlinks Server URL not provided. Run \"shorten -s [Server URL]\" to set your default server url.")
		}
		if options.Code != nil {
			submitUrl(options.URLs[0], serverURL, options.Code, options.Plain)
		} else {
			for _, link := range options.URLs {
				submitUrl(link, serverURL, nil, options.Plain)
			}
		}
	}
}

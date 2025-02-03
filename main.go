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

var options struct {
	URLs   []url.URL `arg:"positional"`
	Plain  bool      `arg:"-p,--plain" help:"only outputs created shortlinks seperated by newlines"`
	Code   *string   `arg:"-c,--request-code" placeholder:"CODE" help:"path to be requested from server"`
	Server *url.URL  `arg:"-s,--set-server" placeholder:"URL" help:"server url to use for this conversion. When ran with this option alone, it sets the default server."`
	Help   bool      `arg:"-h,--help" help:"show this help message"`
}

var configPath = loadConfig()

func loadConfig() string {
	configPath := os.Getenv("XDG_CONFIG_HOME")
	if len(configPath) == 0 {
		homeDir, _ := os.UserHomeDir()
		configPath = homeDir + string(os.PathSeparator) + ".config"
	}
	return configPath + string(os.PathSeparator) + "shortlinks" + string(os.PathSeparator)
}

func submitUrl(url url.URL, server *url.URL, authToken string, code *string, plain bool) {
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
	// Create request
	req, err := http.NewRequest("POST", server.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		panic("Error creating post request to server!")
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic("Error sending post request to server!")
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 201 {
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

func authenticate(server *url.URL) []byte {
	// Collect Username, Password, and OTP from user
	var username, password, otp string
	fmt.Print("Username: ")
	fmt.Scanln(&username)
	fmt.Print("Password: ")
	fmt.Scanln(&password)
	fmt.Print("OTP: ")
	fmt.Scanln(&otp)
	reqBody := []byte(`{
		"username": "` + username + `",
		"password": "` + password + `",
		"otp": "` + otp + `"
	}`)
	// Create request
	req, err := http.NewRequest("POST", server.String()+"/login", bytes.NewBuffer(reqBody))
	if err != nil {
		panic("Error creating post request to server!")
	}
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic("Error sending post request to server!")
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("Error " + resp.Status + " while authenticating.\n" + string(body))
	}
	os.Remove(configPath)
	folders := strings.Split(configPath, string(os.PathSeparator))
	os.MkdirAll(strings.Join(folders[:len(folders)-1], string(os.PathSeparator)), 0666)
	os.WriteFile(configPath+"authToken.txt", []byte(body), 0666)
	if !options.Plain {
		fmt.Println("Authentication successful. Session Token saved.\n", string(body))
	}
	return body
}

func main() {
	// Load settings - Default Server URL
	var serverURL *url.URL
	serverURLdata, err := os.ReadFile(configPath + "server.txt")
	if err == nil {
		serverURL, err = url.Parse(string(serverURLdata))
		if err != nil {
			panic("Error reading saved server url. Delete `server.txt` from your config directory (" + configPath + ") and try again.")
		}
	}

	// Parse arguments
	p := arg.MustParse(&options)

	if (options.Code == nil && options.Server == nil && len(options.URLs) == 0) || options.Help {
		p.WriteHelp(os.Stdout)
	} else if options.Code != nil && len(options.URLs) != 1 {
		p.Fail("Request-Code option only valid when shortening a single url.")
	} else if options.Server != nil && len(options.URLs) == 0 {
		// Set default server
		os.Remove(configPath)
		folders := strings.Split(configPath, string(os.PathSeparator))
		os.MkdirAll(strings.Join(folders[:len(folders)-1], string(os.PathSeparator)), 0666)
		os.WriteFile(configPath+"server.txt", []byte(options.Server.String()), 0666)
		if !options.Plain {
			fmt.Println("Default Shortlinks Server URL set. Authenticate below to save a new session token.")
			authenticate(options.Server)
		}
	} else {
		if options.Server != nil {
			serverURL = options.Server
		}
		if serverURL == nil {
			p.Fail("Shortlinks Server URL not provided. Run \"shorten -s [Server URL]\" to set your default server url.")
		}
		// Check Auth
		authToken, err := os.ReadFile(configPath + "authToken.txt")
		if err != nil {
			fmt.Println("No authentication token found. Login Required.")
			authToken = authenticate(serverURL)
		}

		if options.Code != nil {
			submitUrl(options.URLs[0], serverURL, string(authToken), options.Code, options.Plain)
		} else {
			for _, link := range options.URLs {
				submitUrl(link, serverURL, string(authToken), nil, options.Plain)
			}
		}
	}
}

package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type LocationResponse struct {
	AaData []struct {
		Lat       string `json:"lat"`
		Lng       string `json:"lng"`
		Gpstime   string `json:"gpstime"`
		LatGoogle string `json:"lat_google"`
		LngGoogle string `json:"lng_google"`
	} `json:"aaData"`
}

var InvalidCookieError = errors.New("Invalid cookie")

func getHTTPClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return client
}

func createSessionCookie() string {
	u := "https://www.365gps.net/login.php"

	client := getHTTPClient()
	resp, err := client.Get(u)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	setCookie := resp.Header.Get("Set-Cookie")
	cookies := strings.Split(setCookie, ";")

	// Iterate through the cookies to find the one that starts with "PHPSESSID"
	for _, cookie := range cookies {
		cookie = strings.TrimSpace(cookie)
		if strings.HasPrefix(cookie, "PHPSESSID=") {
			return cookie
		}
	}

	panic("No PHPSESSID cookie found")
}

func login(cookie string, username string, password string) {
	hc := getHTTPClient()
	u := "https://www.365gps.net/npost_login.php?lang=en"

	form := url.Values{}
	form.Add("demo", "F")
	form.Add("username", username)
	form.Add("password", password)
	form.Add("form_type", "0")

	req, err := http.NewRequest("POST", u, strings.NewReader(form.Encode()))

	if err != nil {
		panic(err)
	}

	req.Header.Add("Cookie", cookie)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")

	resp, err := hc.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
}

func refreshLocation(cookie string, imei string) {
	hc := getHTTPClient()
	u := "https://www.365gps.net/post_submit_sendloc.php"

	form := url.Values{}
	form.Add("imei", imei)

	req, err := http.NewRequest("POST", u, strings.NewReader(form.Encode()))

	if err != nil {
		panic(err)
	}

	req.Header.Add("Cookie", cookie)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")

	resp, err := hc.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	// Check for BOM and strip it if present
	if len(bodyBytes) >= 3 && bodyBytes[0] == 0xEF && bodyBytes[1] == 0xBB && bodyBytes[2] == 0xBF {
		bodyBytes = bodyBytes[3:]
	}
	log.Default().Println("Cookie is: ", cookie)
	log.Default().Println("Refresh location response: ", string(bodyBytes))

	if respString := strings.TrimSpace(string(bodyBytes)); respString != "Y" {
		log.Fatalln("Failed to refresh location", respString)
	}

}
func getLocation(cookie string) (float64, float64, error) {
	hc := getHTTPClient()
	u := "https://www.365gps.net/post_map_marker_list.php?timezonemins=-180"

	req, err := http.NewRequest("GET", u, nil)

	if err != nil {
		panic(err)
	}

	req.Header.Add("Cookie", cookie)
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")

	resp, err := hc.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	var responseJson LocationResponse
	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	// Check for BOM and strip it if present
	if len(bodyBytes) >= 3 && bodyBytes[0] == 0xEF && bodyBytes[1] == 0xBB && bodyBytes[2] == 0xBF {
		bodyBytes = bodyBytes[3:]
	}

	if string(bodyBytes) == "{\"result\":\"NULL\"}" {
		return 0, 0, InvalidCookieError
	}

	if !json.Valid(bodyBytes) {
		panic("Response is not valid JSON")
	}

	err = json.Unmarshal(bodyBytes, &responseJson)
	if err != nil {
		panic(err)
	}

	latitude, err := strconv.ParseFloat(responseJson.AaData[0].LatGoogle, 64)
	if err != nil {
		panic(err)
	}

	longitude, err := strconv.ParseFloat(responseJson.AaData[0].LngGoogle, 64)
	if err != nil {
		panic(err)
	}

	return latitude, longitude, nil
}

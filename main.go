package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// CmwURL is the base url for the capital mind wealth site
	CmwURL = "https://progress.capitalmindwealth.com"

	// loginURL
	loginURL = "https://progress.capitalmindwealth.com/accounts/login/"
)

type cmwClient struct {
	client    http.Client    //http client
	jar       *cookiejar.Jar //jar
	csrftoken string         //csrftoken
	username  string         //username
	password  string         //password
	clientID  string         //client ID
}

// init - initialises the http client
func (s *cmwClient) init() {
	s.jar, _ = cookiejar.New(nil)
	s.client = http.Client{Jar: s.jar}
	s.username = os.Getenv("MY_USERNAME")
	s.password = os.Getenv("MY_PASSWORD")
	s.clientID = os.Getenv("MY_CLIENTID")
}

// getCookies - Gets cookies
func (s *cmwClient) getCookies() map[string][]*http.Cookie {
	cookies := make(map[string][]*http.Cookie)
	r, err := s.client.Get(CmwURL + "/accounts/login/?next=/")
	if err != nil {
		fmt.Println(err)
		return cookies
	}
	siteCookies := s.jar.Cookies(r.Request.URL)
	cookies["cookies"] = siteCookies
	defer r.Body.Close()
	return cookies
}

// getCsrfToken - fetches the csrf token given a list of cookies
func (s *cmwClient) getCsrfToken(cookies []*http.Cookie) {
	for _, c := range cookies {
		if c.Name == "csrftoken" {
			s.csrftoken = c.Value
		}
	}
	if s.csrftoken == "" {
		panic("there is no cookie with csrftoken")
	}
}

// getSessionCookies - gets session cookies
func (s *cmwClient) getSessionCookies() {
	data := url.Values{}
	data.Add("csrfmiddlewaretoken", s.csrftoken)
	data.Add("login", s.username)
	data.Add("password", s.password)
	data.Add("remember", "1")
	data.Add("next", "/")

	req, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	req.Header.Add("Referer", CmwURL+"/accounts/login/?next=/")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r, err := s.client.Do(req)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Body.Close()
}

// getDailyData - Gets the current allocation numbers across various asset classes
func (s *cmwClient) getDailyData() {
	r, err := s.client.Get(CmwURL + "/api/client/portfolio/daily_master_values/" + s.clientID)
	if err != nil {
		fmt.Println(err)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	fmt.Printf("%s", data)
}

// getDailyPortfolio - Gets your current portfolio
func (s *cmwClient) getDailyPortfolio() {
	today := time.Now().Local()
	r, err := s.client.Get(CmwURL + "/api/client/portfolio/holdings/client/daily/report/" + s.clientID + "?given_date=" + today.Format("2006-01-02"))
	if err != nil {
		fmt.Println(err)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	fmt.Printf("%s", data)
}

func main() {
	cc := &cmwClient{}
	cc.init()
	cookies := cc.getCookies()
	cc.getCsrfToken(cookies["cookies"])
	cc.getSessionCookies()
	cc.getDailyData()
	cc.getDailyPortfolio()
}

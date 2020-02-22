package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

const htmlIndex = `<html><body>
<a href="/login">Log in with Microsoft</a>
</body></html>
`

var endpoint = oauth2.Endpoint{
	AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
	TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
}

var apis []string = []string{
	"https://graph.microsoft.com/v1.0/me/",
	"https://graph.microsoft.com/v1.0/me/messages",
}

var msOauthConfig = &oauth2.Config{
	Endpoint: endpoint,
}

var period time.Duration = 30 * time.Second
var done = make(chan bool, 1)
var listen = "127.0.0.1:3000"

const oauthStateString = "random"

func init() {
	viper.SetConfigName("e5go")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath("$HOME")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Fatal error config file: %s \n", err)
	}

	msOauthConfig.ClientID = viper.GetString("client_id")
	msOauthConfig.ClientSecret = viper.GetString("client_secret")
	msOauthConfig.Scopes = viper.GetStringSlice("scope")
	msOauthConfig.RedirectURL = viper.GetString("redirect_uri")
	period = viper.GetDuration("period")
	apis = viper.GetStringSlice("apis")
	listen = viper.GetString("listen")
}

func readToken(token *oauth2.Token) error {
	token.AccessToken = viper.GetString("token.access_token")
	token.RefreshToken = viper.GetString("token.refresh_token")
	token.TokenType = viper.GetString("token.token_type")
	token.Expiry = viper.GetTime("token.expiry")

	if token.AccessToken == "" {
		return fmt.Errorf("no access_token loaded")
	}
	return nil
}

func saveToken(token *oauth2.Token) error {
	viper.Set("token.access_token", token.AccessToken)
	viper.Set("token.refresh_token", token.RefreshToken)
	viper.Set("token.token_type", token.TokenType)
	viper.Set("token.expiry", token.Expiry)

	return viper.WriteConfig()
}

func main() {
	var token = &oauth2.Token{}
	if err := readToken(token); err != nil {
		log.Println("no token loaded")
	} else {
		log.Println("token loaded")
		trigger(done)
	}

	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)
	log.Printf("listen at http://%s\n", listen)
	log.Println(http.ListenAndServe(listen, nil))
}

func trigger(done chan bool) {
	ticker := time.NewTicker(period)
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	go func() {
		for {
			select {
			case <-done:
				log.Println("trigger stopped")
				return
			case <-ticker.C:
				accessAPI(apis[r.Intn(len(apis))])
			}
		}
	}()

	log.Println("trigger started")
}

func accessAPI(url string) {
	var token = &oauth2.Token{}
	if err := readToken(token); err != nil {
		log.Println("no token loaded, skip access API")
	}

	tokenSource := msOauthConfig.TokenSource(oauth2.NoContext, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalln(err)
	}

	if newToken.AccessToken != token.AccessToken {
		saveToken(newToken)
		log.Println("Saved new token, expire at ", newToken.Expiry)
	}

	client := oauth2.NewClient(oauth2.NoContext, tokenSource)
	res, err := client.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	if res.StatusCode != 200 {
		log.Printf("access failed[%d]: %s", res.StatusCode, url)
		return
	}
	log.Println("access succeed ", url)
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlIndex)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	url := msOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := msOauthConfig.Exchange(oauth2.NoContext, code)

	if err != nil {
		log.Printf("Code exchange failed with %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	log.Println("got token")
	done <- true
	done = make(chan bool, 1)
	saveToken(token)
	trigger(done)

	data, _ := json.Marshal(token)
	w.Write(data)
}

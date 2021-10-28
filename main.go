package main

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"gopkg.in/irc.v3"
	"gopkg.in/yaml.v2"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	Consumer struct {
		ConsumerKey    string `yaml:"consumer_key"`
		ConsumerSecret string `yaml:"consumer_secret"`
	} `yaml:"consumer"`
	Token struct {
		TokenKey    string `yaml:"token_key"`
		TokenSecret string `yaml:"token_secret"`
	} `yaml:"token"`
	IRC struct {
		Connect string `yaml:"connection"`
		Chan    string `yaml:"chan"`
		Nick    string `yaml:"nick"`
		User    string `yaml:"user"`
		Name    string `yaml:"name"`
	} `yaml:"irc"`
}

func getTweetID(line string) int64 {
	re := regexp.MustCompile(`https://[^\s]+`)
	url := re.FindString(line)
	var splitURL = strings.Split(url, "/")
	tweetID := splitURL[len(splitURL)-1:]
	intID, err := strconv.ParseInt(tweetID[0], 10, 64)
	if err != nil {
		return 0
	}
	return intID
}
func handleMessage(message string, tclient *twitter.Client) string {
	if strings.Contains(message, "https://twitter.com") == true {
		log.Printf("its a tweet, fetch it")
		tweetID := getTweetID(message)
		if tweetID == 0 {
			log.Printf("tried to eat bad tweet url")
			return ""
		}

		params := new(twitter.StatusLookupParams)
		params.TweetMode = "extended"
		tweets, _, err := tclient.Statuses.Lookup([]int64{tweetID}, params)
		if err != nil {
			log.Fatalln(err)
			return ""
		}
		builtResponse := fmt.Sprintf("@%s: %s", tweets[0].User.ScreenName, tweets[0].FullText)

		return builtResponse
	}

	var firstChars = message[0:5]
	if firstChars == "twit " {
		log.Printf("its a twurt")
		log.Printf(message)
		_, _, err := tclient.Statuses.Update(message[5:], nil)
		if err != nil {
			log.Fatalln(err)
			return ""
		}
		return "tweeted"
	}
	return ""
}
func main() {
	f, err := os.Open("config.yml")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := net.Dial("tcp", cfg.IRC.Connect)
	if err != nil {
		log.Fatalln(err)
	}
	tconfig := oauth1.NewConfig(cfg.Consumer.ConsumerKey, cfg.Consumer.ConsumerSecret)
	token := oauth1.NewToken(cfg.Token.TokenKey, cfg.Token.TokenSecret)
	// http.Client will automatically authorize Requests
	httpClient := tconfig.Client(oauth1.NoContext, token)

	// Twitter client
	tclient := twitter.NewClient(httpClient)
	config := irc.ClientConfig{
		Nick: cfg.IRC.Nick,
		User: cfg.IRC.User,
		Name: cfg.IRC.Name,
		Handler: irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
			if m.Command == "001" {
				// 001 is a welcome event, so we join channels there
				joinString := fmt.Sprintf("JOIN %s", cfg.IRC.Chan)
				err := c.Write(joinString)
				if err != nil {
					log.Fatalln(err)
				}
			} else if m.Command == "PRIVMSG" && c.FromChannel(m) {
				// Create a handler on all messages.
				log.Printf(m.Trailing())
				response := handleMessage(m.Trailing(), tclient)
				if len(response) != 0 {
					err := c.WriteMessage(&irc.Message{
						Command: "PRIVMSG",
						Params: []string{
							m.Params[0],
							response,
						},
					})
					if err != nil {
						log.Fatalln(err)
					}

				}
			}
		}),
	}

	// Create the client
	client := irc.NewClient(conn, config)
	err = client.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

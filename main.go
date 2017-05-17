package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/mattn/go-mastodon"
)

const TimeLayout = "2006-01-02 15:04"

type config struct {
	UserName     string `json:"user_name"`
	Password     string `json:"password"`
	Server       string `json:"server"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

var (
	public       = flag.Bool("p", false, "show public timeline")
	local        = flag.Bool("l", false, "show local timeline")
	user         = flag.String("u", "", "show user statuses")
	notification = flag.Bool("n", false, "show notifications")
)

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage of mstdn:
  -p PUBLIC: show public timeline
  -l LOCAL: show local timeline
  -u USER_STATUSES: show user statuses
  -n NOTIFICATION: show notifications
  -i ID: specify in-reply ID, if not specify text, it will be RT.
`)
	}
	flag.Parse()

	cfg, err1 := loadConfig()
	if err1 != nil {
		log.Fatal("loadConfig error:", err1)
	}

	c := mastodon.NewClient(&mastodon.Config{
		Server:       cfg.Server,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
	})

	err := c.Authenticate(context.Background(), cfg.UserName, cfg.Password)
	if err != nil {
		log.Fatal(err)
	}

	if *public {
		showTimelinePublic(c, false)
	} else if *local {
		showTimelinePublic(c, true)
	} else if *user != "" {
		showUserStatuses(c, *user)
	} else if *notification {
		showNotification(c)
	} else if flag.NArg() == 0 {
		showTimelineHome(c)
	} else {
		toot(c, strings.Join(flag.Args(), " "))
	}
}

func showTimelineHome(c *mastodon.Client) {
	tl, err := c.GetTimelineHome(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	showStatuses(tl)
}

func showTimelinePublic(c *mastodon.Client, isLocal bool) {
	tl, err := c.GetTimelinePublic(context.Background(), isLocal, nil)
	if err != nil {
		log.Fatal(err)
	}
	showStatuses(tl)
}

func showStatuses(timeline []*mastodon.Status) {
	for i := len(timeline) - 1; i >= 0; i-- {
		r := strings.NewReader(timeline[i].Content)
		doc, err := goquery.NewDocumentFromReader(r)
		if err != nil {
			log.Fatal(err)
		}

		doc.Find("p").Each(func(_ int, s *goquery.Selection) {
			text := s.Text()
			fmt.Print(timeline[i].CreatedAt.Format(TimeLayout), " ")
			color.Set(color.FgBlue)
			fmt.Print(timeline[i].Account.Username, " ")
			color.Set(color.Reset)
			fmt.Println(text)
		})
	}
}

func toot(c *mastodon.Client, status string) {
	st, err := c.PostStatus(context.Background(), &mastodon.Toot{
		Status: status,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("toot!:", st.ID)
}

func showNotification(c *mastodon.Client) {
	notifications, err := c.GetNotifications(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range notifications {
		if n.Status != nil {
			r := strings.NewReader(n.Status.Content)
			doc, err := goquery.NewDocumentFromReader(r)
			if err != nil {
				log.Fatal(err)
			}
			doc.Find("p").Each(func(_ int, s *goquery.Selection) {
				fmt.Print(n.CreatedAt.Format(TimeLayout))
				color.Set(color.FgBlue)
				fmt.Print(" " + n.Account.Acct)
				color.Set(color.FgYellow)
				fmt.Print(" " + n.Type)
				color.Set(color.Reset)
				fmt.Println(" " + s.Text())
			})
		}
	}
}

func showUserStatuses(c *mastodon.Client, username string) {
	users, err := c.AccountsSearch(context.Background(), username, 1)
	if err != nil {
		log.Fatal(err)
	}
	statuses, err := c.GetAccountStatuses(context.Background(), users[0].ID, nil)
	if err != nil {
		log.Fatal(err)
	}
	showStatuses(statuses)
}

func loadConfig() (*config, error) {
	home := os.Getenv("HOME")
	if home == "" && runtime.GOOS == "windows" {
		home = os.Getenv("APPDATA")
	}
	fname := filepath.Join(home, ".config", "mstdn", "config.json")
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	var cfg config
	err = json.NewDecoder(f).Decode(&cfg)
	return &cfg, err
}

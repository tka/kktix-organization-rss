package main

import (
  "code.google.com/p/go.net/html"
  "fmt"
  . "github.com/PuerkitoBio/goquery"
  . "github.com/gorilla/feeds"
  "io"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "regexp"
  "strings"
  "time"
)

var port = func() string {
  tmpport := os.Getenv("PORT")
  if tmpport == "" {
    tmpport = "5000"
  }

  return tmpport
}

type organization struct {
  slug string
  html string
}

type event struct {
  name string
  url  string
}

func (org *organization) URL() string {
  return "http://" + org.slug + ".kktix.cc"
}

func (org *organization) GetHtmlPage() string {
  if len(org.html) == 0 {
    client := http.Client{}

    s1 := time.Now()
    resp, err := client.Get(org.URL())
    s2 := time.Now()
    fmt.Printf("get spend: %f\n", float64(s2.UnixNano()-s1.UnixNano())/1000000000)
    if err != nil {
      fmt.Println(err)
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    org.html = string(body)
  }
  return org.html
}

func (org *organization) Title() string {
  r, _ := regexp.Compile("<h1>(.*)</h1>")
  matchs := r.FindStringSubmatch(org.html)
  return matchs[1]
}

func (org *organization) Events() []event {
  var doc *Document
  var e error
  node, _ := html.Parse(strings.NewReader(org.html))
  doc = NewDocumentFromNode(node)

  if e != nil {
    fmt.Println(org.html)
    panic(e.Error())
  }

  events := make([]event, 0)

  doc.Find("ul.event-list h2 a").Each(func(i int, s *Selection) {
    href, _ := s.Attr("href")
    new_event := event{name: s.Text(), url: href}
    events = append(events, new_event)
  })

  doc.Find("ul.event-list-past li a").Each(func(i int, s *Selection) {
    href, _ := s.Attr("href")
    new_event := event{name: s.Text(), url: href}
    events = append(events, new_event)
  })

  fmt.Println(events)
  return events
}

func OutputRSS(w http.ResponseWriter, req *http.Request) {
  query_values := req.URL.Query()
  slugs := query_values["slug"]

  if len(slugs) > 0 {
    org := organization{slug: slugs[0]}
    org.GetHtmlPage()
    now := time.Now()

    feed := &Feed{
      Title:   org.Title(),
      Link:    &Link{Href: org.URL()},
      Created: now,
    }

    events := org.Events()
    events_len := len(events)
    feed.Items = make([]*Item, events_len)

    for i := 0; i < events_len; i++ {
      item_pointer := &Item{
        Title: events[i].name,
        Link:  &Link{Href: events[i].url},
      }
      feed.Items[i] = item_pointer
    }

    rss, _ := feed.ToRss()
    io.WriteString(w, rss)
  } else {
    io.WriteString(w, "need slug")
  }

}

func main() {
  http.HandleFunc("/", OutputRSS)
  log.Fatal(http.ListenAndServe(":"+port(), nil))
}

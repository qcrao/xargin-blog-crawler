package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://xargin.com"

type BlogPost struct {
	Title       string
	PublishDate string
	ReadTime    string
	URL         string
}

func main() {
	startURL, firstPostTitle := getFirstPostURLAndTitle()
	previousPosts, isFirstPostAlreadyCrawled := loadPreviousPosts(firstPostTitle)

	file, err := os.Create("xargin_blogs.md")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	now := time.Now().Format("2006-01-02 15:04")
	totalPosts := len(previousPosts)

	if isFirstPostAlreadyCrawled {
		fmt.Println("The first post on the homepage has already been crawled.")
	} else {
		blogPosts := crawlBlogPosts(startURL)
		previousPosts = append(previousPosts, blogPosts...)
		totalPosts = len(previousPosts)
	}

	fmt.Fprintln(file, fmt.Sprintf("页面更新时间：%s\n文章总数：%d\n", now, totalPosts))
	fmt.Fprintln(file, "| 序号 | 文章 | 发表时间 | 阅读时间 |")
	fmt.Fprintln(file, "| --- | --- | --- | --- |")

	for index, post := range previousPosts {
		fmt.Fprintf(file, "| %d | [%s](%s) | %s | %s |\n", totalPosts-index, post.Title, post.URL, post.PublishDate, post.ReadTime)
	}
}

func getFirstPostURLAndTitle() (string, string) {
	resp, err := http.Get(baseURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	postURL, _ := doc.Find(".post-title-link").First().Attr("href")
	postTitle := doc.Find(".post-title-link").First().Text()

	return baseURL + postURL, postTitle
}

func loadPreviousPosts(firstPostTitle string) ([]BlogPost, bool) {
	previousPosts := []BlogPost{}
	isFirstPostAlreadyCrawled := false

	content, err := ioutil.ReadFile("xargin_blogs.md")
	if err != nil {
		return previousPosts, isFirstPostAlreadyCrawled
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, firstPostTitle) {
			isFirstPostAlreadyCrawled = true
		}
		if strings.HasPrefix(line, "| [") {
			parts := strings.Split(line, "|")
			postTitleAndURL := strings.TrimSpace(parts[1])
			postTitleAndURL = strings.TrimPrefix(postTitleAndURL, "[")
			postTitleAndURL = strings.TrimSuffix(postTitleAndURL, ")")

			title := strings.Split(postTitleAndURL, "](")[0]
			url := strings.Split(postTitleAndURL, "](")[1]

			publishDate := strings.TrimSpace(parts[2])
			readTime := strings.TrimSpace(parts[3])

			post := BlogPost{
				Title:       title,
				PublishDate: publishDate,
				ReadTime:    readTime,
				URL:         url,
			}
			previousPosts = append(previousPosts, post)
		}
	}

	return previousPosts, isFirstPostAlreadyCrawled
}

func crawlBlogPosts(url string) []BlogPost {
	var posts []BlogPost

	for {
		fmt.Println("start get post", url)
		post, prevURL, err := getBlogPost(url)
		if err != nil {
			break
		}

		posts = append(posts, *post)

		if prevURL == "" {
			break
		}

		url = baseURL + prevURL
		fmt.Println("get post done", url, post.Title)
		time.Sleep(1 * time.Second) // 为了避免频繁请求，设置一个简单的延迟
	}

	return posts
}

func getBlogPost(url string) (*BlogPost, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", err
	}

	title := doc.Find(".post-title").Text()
	publishDate, _ := doc.Find(".post-meta-date time").Attr("datetime")
	readTime := strings.TrimSpace(doc.Find(".post-meta-item.post-meta-length").Text())
	prevURL, _ := doc.Find(".navigation-item.navigation-previous a").Attr("href")

	post := &BlogPost{
		Title:       title,
		PublishDate: publishDate,
		ReadTime:    readTime,
		URL:         url,
	}

	return post, prevURL, nil
}

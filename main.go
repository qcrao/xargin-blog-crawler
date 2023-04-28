package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://xargin.com"
const mdFilename = "xargin_blogs.md"

type BlogPost struct {
	Title       string
	PublishDate string
	ReadTime    string
	URL         string
}

func main() {
	startURL, firstPostTitle := getFirstPostURLAndTitle()
	previousPosts, isFirstPostAlreadyCrawled := loadPreviousPosts(firstPostTitle)

	// 创建或打开文件
	file, err := os.OpenFile(mdFilename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 使用 UTC+8 北京时间
	loc, _ := time.LoadLocation("Asia/Shanghai")
        now := time.Now().In(loc).Format("2006-01-02 15:04")

	if isFirstPostAlreadyCrawled {
		fmt.Println("The first post on the homepage has already been crawled.")
		updateTimestampInFile(mdFilename)
		return
	}

	blogPosts := crawlBlogPosts(startURL)
	previousPosts = append(previousPosts, blogPosts...)
	totalPosts := len(previousPosts)

	fmt.Fprintf(file, "页面更新时间（北京时间）：%s\n", now)
	fmt.Fprintln(file, "\n")
	fmt.Fprintf(file, "文章总数：%d\n", totalPosts)	
	fmt.Fprintln(file, "| 序号 | 文章 | 发表时间 | 阅读时间 |")
	fmt.Fprintln(file, "| --- | --- | --- | --- |")

	for index, post := range previousPosts {
		fmt.Fprintf(file, "| %d | [%s](%s) | %s | %s |\n", totalPosts-index, post.Title, post.URL, post.PublishDate, post.ReadTime)
	}
}

func updateTimestampInFile(filename string) {
	// 使用 UTC+8 北京时间
    	loc, _ := time.LoadLocation("Asia/Shanghai")
    	currentTime := time.Now().In(loc).Format("2006-01-02 15:04")

	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	lines := strings.Split(string(fileData), "\n")
	if len(lines) > 0 {
		lines[0] = fmt.Sprintf("页面更新时间（北京时间）：%s\n", currentTime)
	}

	updatedFileData := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filename, []byte(updatedFileData), 0644)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
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

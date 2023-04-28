# xargin-blog-crawler

xargin-blog-crawler 是一个用 Go 编写的博客爬虫，定期抓取并更新 [xargin.com](https://xargin.com) 上的文章信息。

## 功能

- 抓取 xargin.com 上的所有文章信息，包括标题、发表时间、阅读时间和 URL
- 将文章信息存储在 Markdown 文件中
- 使用 GitHub Actions 每小时自动更新文章信息

## 使用方法

1. 克隆项目到本地

```shell
git clone https://github.com/qcrao/xargin-blog-crawler.git
```

2. 运行程序

程序将抓取文章信息并将其存储在 `xargin_blogs.md` 文件中。

## GitHub Actions

本项目使用 GitHub Actions 定时执行 Go 程序并将结果推送到当前项目。具体配置请参考 `.github/workflows/crawler.yml` 文件。

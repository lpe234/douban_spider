package main

import (
	"encoding/csv"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"log"
	"os"
	"strings"
	"time"
)

type Movie struct {
	idx    string
	title  string
	year   string
	info   string
	rating string
	url    string
}

func main() {
	// 存储文件名
	fName := "douban_movie_top250.csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("创建文件失败 %q: %s\n", fName, err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	// 写CSV头部
	writer.Write([]string{"Idx", "Title", "Year", "Info", "Rating", "URL"})

	// 起始Url
	startUrl := "https://movie.douban.com/top250"

	// 创建Collector
	collector := colly.NewCollector(
		// 设置用户代理
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36"),
	)

	// 设置抓取频率限制
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 5 * time.Second, // 随机延迟
	})

	// 异常处理
	collector.OnError(func(response *colly.Response, err error) {
		log.Println(err.Error())
	})

	collector.OnRequest(func(request *colly.Request) {
		log.Println("start visit: ", request.URL.String())
	})

	// 解析列表
	collector.OnHTML("ol.grid_view", func(element *colly.HTMLElement) {
		// 依次遍历所有的li节点
		element.DOM.Find("li").Each(func(i int, selection *goquery.Selection) {
			href, found := selection.Find("div.hd > a").Attr("href")
			// 如果找到了详情页，则继续下一步的处理
			if found {
				parseDetail(collector, href, writer)
				log.Println(href)
			}
		})
	})

	// 查找下一页
	collector.OnHTML("div.paginator > span.next", func(element *colly.HTMLElement) {
		href, found := element.DOM.Find("a").Attr("href")
		// 如果有下一页，则继续访问
		if found {
			element.Request.Visit(element.Request.AbsoluteURL(href))
		}
	})

	// 起始入口
	collector.Visit(startUrl)
}

/**
 * 处理详情页
 */
func parseDetail(collector *colly.Collector, url string, writer *csv.Writer) {
	collector = collector.Clone()

	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 2 * time.Second,
	})

	collector.OnRequest(func(request *colly.Request) {
		log.Println("start visit: ", request.URL.String())
	})

	// 解析详情页数据
	collector.OnHTML("body", func(element *colly.HTMLElement) {
		selection := element.DOM.Find("div#content")
		idx := selection.Find("div.top250 > span.top250-no").Text()
		title := selection.Find("h1 > span").First().Text()
		year := selection.Find("h1 > span.year").Text()
		info := selection.Find("div#info").Text()
		info = strings.ReplaceAll(info, " ", "")
		info = strings.ReplaceAll(info, "\n", "; ")
		rating := selection.Find("strong.rating_num").Text()
		movie := Movie{
			idx:    idx,
			title:  title,
			year:   year,
			info:   info,
			rating: rating,
			url:    element.Request.URL.String(),
		}
		writer.Write([]string{
			idx,
			title,
			year,
			info,
			rating,
			element.Request.URL.String(),
		})
		log.Printf("%+v", movie)
	})

	collector.Visit(url)
}

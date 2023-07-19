package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

// 定义一个结构体，存放商品信息 title 价格，图片地址
type Item struct {
	Title string
	Price string
	Image string
}

func init() {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())
}
func main() {

	// 创建一个新的收集器  a-section a-spacing-base
	c := colly.NewCollector()
	count := 0
	//定义切片存放商品信息
	var items []Item
	var wg sync.WaitGroup // 创建一个 WaitGroup
	// 在找到每个a.storylink元素时
	c.OnHTML("div.s-main-slot.s-result-list.s-search-results.sg-row", func(h *colly.HTMLElement) {

		h.ForEach("div.a-section.a-spacing-base", func(i int, h *colly.HTMLElement) {
			count++
			//获取a-size-base-plus a-color-base a-text-normal text
			title := h.ChildText("span.a-size-base-plus.a-color-base.a-text-normal")
			//a-offscreen
			price := h.ChildText("span.a-offscreen")
			//打印价格和标题

			//还有图片s-image
			image := h.ChildAttr("img.s-image", "src")
			//如果其中一个为空，就不存入切片
			if title != "" && price != "" && image != "" {
				items = append(items, Item{
					Title: title,
					Price: price,
					Image: image,
				})
			}
		})
	})

	// 访问网站
	c.Visit("https://www.amazon.com/s?k=shoes&crid=3K11YUFUJOO07&sprefix=shoes%2Caps%2C309&ref=nb_sb_noss_2")
	//获取当前运行目录
	dir, _ := os.Getwd()
	//遍历切片，开启goroutine下载图片
	//声明随机数种子
	rand.Seed(time.Now().UnixNano())
	for _, item := range items {
		wg.Add(1) // 在启动一个 goroutine 之前，增加 WaitGroup 的计数
		go func(item Item) {
			//拼接上当前运行目录
			filename := fmt.Sprintf("%s/%d%s", dir, rand.Int(), ".jpg")
			DownloadFile(filename, item.Image)
			wg.Done() // 在 goroutine 完成后，减少 WaitGroup 的计数
		}(item)
	}
	wg.Wait() // 等待所有的 goroutine 完成
}

// 传入图片路径下载图片
func DownloadFile(filepath string, url string) error {
	//正在下载图片，路径为filepath，图片地址为url
	fmt.Println("正在下载图片，路径为", filepath, "图片地址为", url)
	// 创建自定义的 HTTP 客户端
	client := &http.Client{
		Timeout: time.Duration(5) * time.Second, // 设置超时时间为 5 秒
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConnsPerHost: 10, // 设置最大空闲连接数
		},
	}
	//根据url创建请求对象
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	//设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"+
		"(KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36")
	//根据请求对象获取响应对象
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	//关闭响应体
	defer resp.Body.Close()
	//根据响应体获取文件流
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	//关闭文件流
	defer file.Close()
	//将响应体中的文件流拷贝到文件中
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

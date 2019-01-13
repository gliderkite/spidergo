all:
	go build -o spiderbot src/spiderbot/spider.go src/spiderbot/spiderbot.go
test:
	go test src/spiderbot/spider_test.go src/spiderbot/spider.go
clean:
	rm spiderbot

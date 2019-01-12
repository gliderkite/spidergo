all:
	go build -o spiderbot src/spiderbot/spider.go src/spiderbot/spiderbot.go
clear:
	rm spiderbot

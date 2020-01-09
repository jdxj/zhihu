localName = zhihu.out
linuxName = zhihu_linux.out
androidName = zhihu_android.out
macName = zhihu_mac.out

local: clean
	go build -ldflags '-s -w' -o $(localName) *.go
linux: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o $(linuxName) *.go
	upx --best $(linuxName)
android: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '-s -w' -o $(androidName) *.go
	upx --best $(androidName)
mac: clean
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o $(macName) *.go
	upx --best $(macName)
clean:
	find . -name "*.log" | xargs rm -f
	find . -name "*.out" | xargs rm -f

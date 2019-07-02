package main

import (
	"fmt"
	"time"

	"github.com/karlseguin/ccache"
)

func main() {
	var cache = ccache.New(ccache.Configure().MaxSize(1000).ItemsToPrune(100))
	cache.Set("user:4", "OMG", time.Minute*10)
	item := cache.Get("user:4")
	fmt.Println(item.Value().(string))

	item = cache.Get("user:5")
	if nil == item {
		fmt.Println("Not Found")
	}
	fmt.Println(item.Value().(string))
}

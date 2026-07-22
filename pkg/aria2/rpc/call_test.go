package rpc

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestWebsocketCaller(t *testing.T) {
	endpoint := os.Getenv("ARIA2_TEST_WS_URL")
	if endpoint == "" {
		t.Skip("set ARIA2_TEST_WS_URL to run the aria2 websocket integration test")
	}
	time.Sleep(time.Second)
	c, err := newWebsocketCaller(context.Background(), endpoint, time.Second, &DummyNotifier{})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer c.Close()

	var info VersionInfo
	if err := c.Call(aria2GetVersion, []interface{}{}, &info); err != nil {
		t.Error(err.Error())
	} else {
		println(info.Version)
	}
}

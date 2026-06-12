package channels

import (
	"testing"
	"time"

	"github.com/fastclaw-ai/fastclaw/internal/bus"
)

func TestWeChatDispatchInboundImageOnly(t *testing.T) {
	mb := bus.New()
	w := &WeChat{
		bus:       mb,
		accountID: "acct-1",
		ctxTokens: map[string]string{},
	}

	w.dispatchInbound(wechatMessage{
		MessageID:    42,
		FromUserID:   "wx-user-1",
		MessageType:  wechatMsgTypeUser,
		MessageState: wechatMsgStateFinish,
		ContextToken: "ctx-1",
		ItemList: []wechatItem{
			{
				Type: wechatItemTypeImage,
				ImageItem: &wechatImageItem{
					URL: "https://img.example/a.jpg",
				},
			},
		},
	})

	select {
	case got := <-mb.Inbound:
		if got.Channel != "wechat" || got.AccountID != "acct-1" {
			t.Fatalf("routing = (%q, %q), want (wechat, acct-1)", got.Channel, got.AccountID)
		}
		if got.ChatID != "wx-user-1" || got.UserID != "wx-user-1" {
			t.Fatalf("user routing = (%q, %q), want (wx-user-1, wx-user-1)", got.ChatID, got.UserID)
		}
		if got.PhotoURL != "https://img.example/a.jpg" {
			t.Fatalf("PhotoURL = %q, want image URL", got.PhotoURL)
		}
		if got.Text != "" {
			t.Fatalf("Text = %q, want empty for image-only message", got.Text)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for inbound image message")
	}
}

func TestWeChatDispatchInboundTextAndMultipleImages(t *testing.T) {
	mb := bus.New()
	w := &WeChat{
		bus:       mb,
		accountID: "acct-1",
		ctxTokens: map[string]string{},
	}

	w.dispatchInbound(wechatMessage{
		MessageID:    43,
		FromUserID:   "wx-user-2",
		MessageType:  wechatMsgTypeUser,
		MessageState: wechatMsgStateFinish,
		ItemList: []wechatItem{
			{
				Type: wechatItemTypeText,
				TextItem: &wechatTextItem{
					Text: "看看这两张图",
				},
			},
			{
				Type: wechatItemTypeImage,
				ImageItem: &wechatImageItem{
					URL: "https://img.example/1.jpg",
				},
			},
			{
				Type: wechatItemTypeImage,
				ImageItem: &wechatImageItem{
					URL: "https://img.example/2.jpg",
				},
			},
		},
	})

	select {
	case got := <-mb.Inbound:
		if got.Text != "看看这两张图" {
			t.Fatalf("Text = %q, want caption text", got.Text)
		}
		if got.PhotoURL != "https://img.example/1.jpg" {
			t.Fatalf("PhotoURL = %q, want first image URL", got.PhotoURL)
		}
		if len(got.PhotoURLs) != 1 || got.PhotoURLs[0] != "https://img.example/2.jpg" {
			t.Fatalf("PhotoURLs = %#v, want second image URL only", got.PhotoURLs)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for inbound text+images message")
	}
}

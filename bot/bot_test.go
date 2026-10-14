package bot

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jomei/notionapi"
	"github.com/mymmrac/telego"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/state"
	"github.com/sirupsen/logrus"
)

// fakeTelegram is a Bot API server that records every call as
// "<method> <json body>" and serves the given updates on the first getUpdates.
type fakeTelegram struct {
	sync.Mutex
	calls   []string
	updates string
	served  bool
}

func (f *fakeTelegram) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]

	f.Lock()
	f.calls = append(f.calls, method+" "+string(body))
	first := method == "getUpdates" && !f.served
	if first {
		f.served = true
	}
	f.Unlock()

	switch method {
	case "getUpdates":
		if first && f.updates != "" {
			io.WriteString(w, `{"ok":true,"result":`+f.updates+`}`)
			return
		}
		select {
		case <-r.Context().Done():
		case <-time.After(50 * time.Millisecond):
		}
		io.WriteString(w, `{"ok":true,"result":[]}`)
	case "sendMessage":
		io.WriteString(w, `{"ok":true,"result":{"message_id":1,"date":0,"chat":{"id":42,"type":"private"}}}`)
	default:
		io.WriteString(w, `{"ok":true,"result":true}`)
	}
}

// methodCalls returns the recorded calls, minus the getUpdates polling.
func (f *fakeTelegram) methodCalls() []string {
	f.Lock()
	defer f.Unlock()
	var out []string
	for _, c := range f.calls {
		if !strings.HasPrefix(c, "getUpdates") {
			out = append(out, c)
		}
	}
	return out
}

func (f *fakeTelegram) allCalls() string {
	f.Lock()
	defer f.Unlock()
	return strings.Join(f.calls, "\n")
}

func newTestBot(t *testing.T, updates string) (*Bot, *fakeTelegram) {
	t.Helper()
	fake := &fakeTelegram{updates: updates}
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)

	tg, err := telego.NewBot("123456:"+strings.Repeat("A", 35),
		telego.WithAPIServer(srv.URL),
		telego.WithHTTPClient(srv.Client()),
		telego.WithDiscardLogger(),
	)
	if err != nil {
		t.Fatal(err)
	}
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return &Bot{
		TelegramClient: tg,
		UserRepo:       db.NewUserRepoMock(map[int]*ent.User{42: {ID: 42}}, nil),
		logger:         logger,
		state:          state.New(),
	}, fake
}

func assertContainsAll(t *testing.T, got string, want []string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in:\n%s", w, got)
		}
	}
}

func TestBotStart(t *testing.T) {
	tests := []struct {
		name    string
		updates string
		want    []string
	}{
		{
			"drops pending updates and only polls handled types",
			`[]`,
			[]string{
				`deleteWebhook {"drop_pending_updates":true}`,
				`"allowed_updates":["message","callback_query"]`,
			},
		},
		{
			"start sends reply keyboard",
			`[{"update_id":1,"message":{"message_id":1,"date":0,"chat":{"id":42,"type":"private"},"text":"/start"}}]`,
			[]string{
				`"text":"Welcome to notion-echo bot!"`,
				`"keyboard":[[{"text":"/note"},{"text":"/register"}],[{"text":"/defaultpage"},{"text":"/getdefaultpage"}],[{"text":"/help"}]]`,
			},
		},
		{
			"setpage callback sets default page and answers query",
			`[{"update_id":1,"callback_query":{"id":"cb1","from":{"id":42,"is_bot":false,"first_name":"x"},"chat_instance":"c","data":"setpage:Inbox","message":{"message_id":2,"date":1,"chat":{"id":42,"type":"private"}}}}]`,
			[]string{
				`sendMessage {"chat_id":42,"text":"Default page set to: Inbox"}`,
				`answerCallbackQuery {"callback_query_id":"cb1","text":"Default page updated"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, fake := newTestBot(t, tt.updates)
			// Cancel rather than time out: telego only stops polling immediately on
			// context.Canceled, a deadline makes it sleep its 8s retry timeout first.
			ctx, cancel := context.WithCancel(context.Background())
			time.AfterFunc(300*time.Millisecond, cancel)

			b.Start(ctx)

			assertContainsAll(t, fake.allCalls(), tt.want)
		})
	}
}

func TestBotStartSkipsCallbackWithoutMessage(t *testing.T) {
	b, fake := newTestBot(t, `[{"update_id":1,"callback_query":{"id":"cb1","from":{"id":42,"is_bot":false,"first_name":"x"},"chat_instance":"c","data":"setpage:Inbox"}}]`)
	// Cancel rather than time out: telego only stops polling immediately on
	// context.Canceled, a deadline makes it sleep its 8s retry timeout first.
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(300*time.Millisecond, cancel)

	b.Start(ctx)

	if calls := fake.methodCalls(); len(calls) != 1 {
		t.Errorf("want only deleteWebhook, got %v", calls)
	}
}

func TestBotSendMessage(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		wantCalls int
	}{
		{"short message", "hello", 1},
		{"long message is split", strings.Repeat("a ", 5000), 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, fake := newTestBot(t, "")

			if err := b.SendMessage(tt.msg, 42, false, true); err != nil {
				t.Fatal(err)
			}

			if calls := fake.methodCalls(); len(calls) != tt.wantCalls {
				t.Errorf("want %d sendMessage calls, got %d", tt.wantCalls, len(calls))
			}
		})
	}
}

func TestBotSendButtonWithURL(t *testing.T) {
	b, fake := newTestBot(t, "")

	if err := b.SendButtonWithURL(42, "Authorize", "https://example.com", "click"); err != nil {
		t.Fatal(err)
	}

	assertContainsAll(t, fake.allCalls(), []string{
		`"text":"click"`,
		`"inline_keyboard":[[{"text":"Authorize","url":"https://example.com"}]]`,
	})
}

func TestBotSendButtonWithData(t *testing.T) {
	page := func(title string) *notionapi.Page {
		return &notionapi.Page{Properties: notionapi.Properties{
			"title": &notionapi.TitleProperty{Title: []notionapi.RichText{{Text: &notionapi.Text{Content: title}}}},
		}}
	}
	tests := []struct {
		name    string
		pages   []*notionapi.Page
		want    []string
		wantErr bool
	}{
		{
			"one button per page, too-long callback data skipped",
			[]*notionapi.Page{page("Inbox"), page(strings.Repeat("x", 64)), page("")},
			[]string{`"inline_keyboard":[[{"text":"Inbox","callback_data":"setpage:Inbox"}]]`},
			false,
		},
		{
			"no valid pages",
			[]*notionapi.Page{page("")},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, fake := newTestBot(t, "")

			err := b.SendButtonWithData(42, "pick a page", tt.pages)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}

			assertContainsAll(t, fake.allCalls(), tt.want)
		})
	}
}

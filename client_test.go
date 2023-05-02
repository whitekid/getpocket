package getpocket

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/whitekid/goxp/cryptox"
)

func newClient(t *testing.T) Interface {
	secretKey := os.Getenv("POCKET_SECRET_KEY")
	consumerKey := os.Getenv("POCKET_TEST_CONSUMER_KEY")
	accessToken := os.Getenv("POCKET_TEST_ACCESS_TOKEN")
	if consumerKey == "" || accessToken == "" {
		t.Skip("consumer_key and access_token required")
	}

	consumerKey = cryptox.MustDecrypt(secretKey, consumerKey)
	accessToken = cryptox.MustDecrypt(secretKey, accessToken)

	return New(consumerKey, accessToken)
}

// Pocket scenerio test
func TestPocket(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := newClient(t)

	resp, err := client.Articles().Get().Detail(DetailComplete).Do(ctx)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(resp))

	now := time.Now()
	// at start, there is no articles
	{
		resp, err := client.Articles().Get().Since(now).Do(ctx)
		require.NoError(t, err)
		require.Equal(t, 0, len(resp))
	}

	// add new article
	var itemId string
	{
		url := "https://www.ciokorea.com/news/228924"
		resp, err := client.Articles().Add(url).Do(ctx)
		require.NoError(t, err)
		require.Equal(t, url, resp.Item.ResolvedURL)
		itemId = resp.Item.ItemID

		defer func() {
			resp, err := client.Modify().Delete(itemId).Do(ctx)
			require.NoError(t, err)
			require.Equal(t, 1, resp.Status)
			require.Equal(t, 1, len(resp.ActionResults))
		}()
	}

	// 1개 추가되었으니, 리스트에 하나 보임
	{
		resp, err := client.Articles().Get().Since(now).Do(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, len(resp))
	}

	// do favorate
	{
		resp, err := client.Modify().Favorite(itemId).Do(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, resp.Status)
		require.Equal(t, 1, len(resp.ActionResults))
	}
}

func TestAdd(t *testing.T) {
	type args struct {
		url string
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{`valid`, args{"https://www.ciokorea.com/news/228924"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			client := newClient(t)
			resp, err := client.Articles().Add(tt.args.url).Do(ctx)
			require.Truef(t, (err != nil) == tt.wantErr, `Add() failed: error = %+v, wantErr = %v`, err, tt.wantErr)
			if tt.wantErr {
				return
			}

			defer func() {
				resp, err := client.Modify().Delete(resp.Item.ItemID).Do(ctx)
				require.NoError(t, err)
				require.Equal(t, 1, resp.Status)
				require.Equal(t, 1, len(resp.ActionResults))
			}()

			require.Equal(t, tt.args.url, resp.Item.ResolvedURL)
		})
	}
}

func TestUnmarshalArticle(t *testing.T) {
	type args struct {
		fixture string
	}
	tests := [...]struct {
		name   string
		args   args
		wantID string
	}{
		{"default", args{"article.json"}, "3574271538"},
		{"without_video", args{"article_without_video.json"}, "3621189735"},
		{"without_image", args{"article_with_image.json"}, "3575623899"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open("fixtures/" + tt.args.fixture)
			require.NoError(t, err)

			var article Article
			require.NoError(t, json.NewDecoder(f).Decode(&article))

			require.Equal(t, tt.wantID, article.ItemID)
			require.Equal(t, article.HasImage, strconv.Itoa(len(article.Images)))
		})
	}
}

func TestUnmarshalGetResponse(t *testing.T) {
	type args struct {
		fixture string
	}
	tests := [...]struct {
		name string
		args args
	}{
		{"default", args{"get_response.json"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := os.Open("fixtures/" + tt.args.fixture)
			defer r.Close()

			var resp GetResponse
			err := json.NewDecoder(r).Decode(&resp)
			require.NoError(t, err)
			require.Equal(t, 1, resp.Status)

			require.NotEqual(t, 0, len(resp.List))
			article := resp.List["3574271538"]
			require.Equal(t, "3574271538", article.ItemID)
		})
	}
}

func TestUnmarshalAddResponse(t *testing.T) {
	type args struct {
		fixture string
	}
	tests := [...]struct {
		name       string
		args       args
		wantVideos int
		wantImages int
	}{
		{"without video", args{"add_response_without_video.json"}, 0, 1},
		{"without image", args{"add_response_without_image.json"}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open("fixtures/" + tt.args.fixture)
			require.NoError(t, err)
			defer f.Close()

			var resp AddResponse
			err = json.NewDecoder(f).Decode(&resp)
			require.NoError(t, err)

			require.Equal(t, tt.wantVideos, len(resp.Item.Videos))
			require.Equal(t, tt.wantImages, len(resp.Item.Images))
		})
	}
}

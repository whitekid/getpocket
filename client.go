package getpocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	gourl "net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/whitekid/goxp"
	"github.com/whitekid/goxp/log"
	"github.com/whitekid/goxp/requests"
)

//go:generate mockery --name Interface
type Interface interface {
	AuthorizedURL(ctx context.Context, redirectURI string) (string, string, error)
	NewAccessToken(ctx context.Context, requestToken string) (string, string, error)

	Articles() ArticleAPI
	Modify() ModifyAPI
}

// clientImpl get pocket api client
// please refer https://getpocket.com/developer/docs/overview
type clientImpl struct {
	consumerKey string
	accessToken string
	sess        requests.Interface // common sessions
}

var _ Interface = (*clientImpl)(nil)

// New create GetPocket API
func New(consumerKey, accessToken string) Interface {
	return &clientImpl{
		consumerKey: consumerKey,
		accessToken: accessToken,
		sess:        requests.NewSession(nil),
	}
}

// Article pocket article, see https://getpocket.com/developer/docs/v3/retrieve
type Article struct {
	ItemID        string           `json:"item_id"`
	ResolvedID    string           `json:"resolved_id"`
	GivenURL      string           `json:"given_url"`
	GivelTitle    string           `json:"given_title"`
	Favorite      string           `json:"favorite"`
	Status        string           `json:"status"`
	ResolvedTitle string           `json:"resolved_title"`
	ResolvedURL   string           `json:"resolved_url"`
	Excerpt       string           `json:"excerpt"`
	IsArticle     string           `json:"is_article"`
	HasVideo      string           `json:"has_video"`
	HasImage      string           `json:"has_image"`
	WordCount     string           `json:"word_count"`
	Images        map[string]Image `json:"-"`
	Videos        map[string]Video `json:"-"`
}

type Image struct {
	ItemID  string `json:"item_id"`
	ImageID string `json:"image_id"`
	Src     string `json:"src"`
	Width   string `json:"width"`
	Height  string `json:"height"`
	Credit  string `json:"credit"`
	Caption string `json:"caption"`
}

type Video struct {
	ItemID  string `json:"item_id"`
	VideoID string `json:"video_id"`
	Src     string `json:"src"`
	Width   string `json:"width"`
	Height  string `json:"height"`
	Type    string `json:"type"`
	Vid     string `json:"vid"`
}

// UnmarshalJSON add의 경우는 images/ videos가 없을 경우 []가 나온다. 이를 변경
func (a *Article) UnmarshalJSON(data []byte) error {
	type ArticleEx Article
	var articleEx ArticleEx

	if err := json.Unmarshal(data, &articleEx); err != nil {
		return err
	}
	*a = Article(articleEx)

	var extras struct {
		Images interface{} `json:"images"` // [] or map[string]Image
		Videos interface{} `json:"videos"` // [] or map[string]Image
	}

	if err := json.Unmarshal(data, &extras); err != nil {
		return err
	}

	if _, ok := extras.Images.([]interface{}); !ok {
		if err := goxp.JsonRecode(&a.Images, &extras.Images); err != nil {
			return err
		}
	}

	if _, ok := extras.Videos.([]interface{}); !ok {
		if err := goxp.JsonRecode(&a.Videos, &extras.Videos); err != nil {
			return err
		}
	}

	return nil
}

func (c *clientImpl) post(url string) *requests.Request {
	u, _ := gourl.JoinPath("https://getpocket.com/v3", url)
	return c.sess.Post(u).Header("X-Accept", "application/json")
}

func (c *clientImpl) sendRequest(ctx context.Context, req *requests.Request) (*requests.Response, error) {
	resp, err := req.Do(ctx)
	if err != nil {
		return resp, errors.Wrap(err, "request failed")
	}

	if err := resp.Success(); err != nil {
		message := resp.Header.Get("x-error")
		code := resp.Header.Get("x-error-code")
		return nil, errors.Wrapf(err, "error=%s, code=%s", message, code)
	}

	return resp, nil

}

// AuthorizedURL get authorizedURL
func (c *clientImpl) AuthorizedURL(ctx context.Context, redirectURI string) (string, string, error) {
	resp, err := c.sendRequest(ctx, c.post("/oauth/request").
		JSON(map[string]string{
			"consumer_key": c.consumerKey,
			"redirect_uri": redirectURI,
		}))
	if err != nil {
		return "", "", errors.Wrap(err, "authorized request failed")
	}

	type Response struct {
		Code string `json:"code"`
	}

	defer resp.Body.Close()
	response, err := goxp.ReadJSON[Response](resp.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "fail to parse json")
	}

	return response.Code, fmt.Sprintf("https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s", response.Code, redirectURI), nil
}

// NewAccessToken get accessToken, username from requestToken using oauth
func (c *clientImpl) NewAccessToken(ctx context.Context, requestToken string) (string, string, error) {
	log.Debugf("getAccessToken with %s", requestToken)

	resp, err := c.sendRequest(ctx, c.post("/oauth/authorize").
		JSON(map[string]string{
			"consumer_key": c.consumerKey,
			"code":         requestToken,
		}))
	if err != nil {
		return "", "", err
	}

	type Response struct {
		AccessToken string `json:"access_token"`
		Username    string `json:"username"`
	}

	defer resp.Body.Close()
	response, err := goxp.ReadJSON[Response](resp.Body)
	if err != nil {
		return "", "", err
	}

	return response.AccessToken, response.Username, nil
}

func (c *clientImpl) Articles() ArticleAPI {
	return &articlesAPIImpl{
		client: c,
	}
}

func (c *clientImpl) Modify() ModifyAPI {
	return &modifyAPIImpl{
		client: c,
	}
}

type ArticleAPI interface {
	Get() GetRequest
	Add(url string) AddRequest
}

// articlesAPIImpl ...
type articlesAPIImpl struct {
	client *clientImpl
}

type Favorate int

const (
	UnFavorited Favorate = iota + 1 // only return un-favorited items
	Favorited                       // only return favorited items
)

type State string

const (
	Unread  = "unread"
	Archive = "archive"
	All     = "all"
)

type DetailType string

const (
	DetailSimple   = "simple"
	DetailComplete = "complete"
)

// GetResponse ...
type GetResponse struct {
	Status int                 `json:"status"`
	List   map[string]*Article `json:"list"`
}

// Get Retrieving a User's Pocket Data
func (a *articlesAPIImpl) Get() GetRequest {
	return &getRequestImpl{
		api:        a,
		detailType: DetailSimple,
	}
}

type GetRequest interface {
	State(state State) GetRequest
	Favorite(favorite Favorate) GetRequest
	Detail(detailType DetailType) GetRequest
	Search(search string) GetRequest
	Domain(domain string) GetRequest
	Since(time.Time) GetRequest

	Do(ctx context.Context) (map[string]*Article, error)
}

type getRequestImpl struct {
	api *articlesAPIImpl

	state      State
	search     string
	domain     string
	favorite   Favorate
	detailType DetailType
	since      time.Time
}

func (r *getRequestImpl) Since(tm time.Time) GetRequest {
	r.since = tm
	return r
}

func (r *getRequestImpl) Search(search string) GetRequest {
	r.search = search
	return r
}

func (r *getRequestImpl) Domain(domain string) GetRequest {
	r.domain = domain
	return r
}

func (r *getRequestImpl) Favorite(favorite Favorate) GetRequest {
	r.favorite = favorite
	return r
}

func (r *getRequestImpl) State(state State) GetRequest {
	r.state = state
	return r
}

func (r *getRequestImpl) Detail(detailType DetailType) GetRequest {
	r.detailType = detailType
	return r
}

func (r *getRequestImpl) Do(ctx context.Context) (map[string]*Article, error) {
	params := map[string]interface{}{
		"consumer_key": r.api.client.consumerKey,
		"access_token": r.api.client.accessToken,
		"state":        r.state,
		"detailType":   r.detailType,
	}

	if r.favorite != 0 {
		params["favorite"] = strconv.FormatInt(int64(r.favorite-1), 10)
	}

	if r.search != "" {
		params["search"] = r.search
	}

	if r.domain != "" {
		params["domain"] = r.domain
	}

	if !r.since.IsZero() {
		params["since"] = strconv.FormatInt(r.since.Unix(), 10)
	}

	resp, err := r.api.client.sendRequest(ctx, r.api.client.post("/get").JSON(params))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	var buf1 bytes.Buffer
	io.Copy(&buffer, resp.Body)
	defer resp.Body.Close()

	//
	tee := io.TeeReader(&buffer, &buf1)

	// return empty list if there is no items searched
	var emptyResponse struct {
		List []string `json:"list"`
	}
	if err := json.NewDecoder(tee).Decode(&emptyResponse); err == nil {
		return nil, err
	}

	var response GetResponse
	if err := json.NewDecoder(&buf1).Decode(&response); err != nil {
		return nil, err
	}

	return response.List, nil
}

func (a *articlesAPIImpl) Add(url string) AddRequest {
	return &addRequestImpl{
		client: a.client,
		url:    url,
	}
}

type AddRequest interface {
	Do(ctx context.Context) (*AddResponse, error)
}

type addRequestImpl struct {
	client *clientImpl

	url string
}

var _ AddRequest = (*addRequestImpl)(nil)

type AddResponse struct {
	Item   Article `json:"item"`
	Status int     `json:"status"`
}

func (a *addRequestImpl) Do(ctx context.Context) (*AddResponse, error) {
	resp, err := a.client.sendRequest(ctx, a.client.post("/add").
		JSON(map[string]string{
			"url":          a.url,
			"consumer_key": a.client.consumerKey,
			"access_token": a.client.accessToken,
		}))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	response, err := goxp.ReadJSON[AddResponse](resp.Body)
	if err != nil {
		return nil, err
	}

	return response, nil
}

type ModifyAPI interface {
	Favorite(itemID ...string) ModifyRequest
	Delete(itemID ...string) ModifyRequest
}

var _ ModifyAPI = (*modifyAPIImpl)(nil)

type modifyAPIImpl struct {
	client *clientImpl
}

func (r *modifyAPIImpl) Delete(itemId ...string) ModifyRequest {
	req := &modifyRequestImpl{
		client: r.client,
	}
	return req.Delete(itemId...)
}

func (r *modifyAPIImpl) Favorite(itemId ...string) ModifyRequest {
	req := &modifyRequestImpl{
		client: r.client,
	}
	return req.Favorite(itemId...)
}

type ModifyRequest interface {
	ModifyAPI

	Do(context.Context) (*ModifyResponse, error)
}

type modifyRequestImpl struct {
	client  *clientImpl
	Actions []modifyAction `json:"actions,omitempty"`
}

var _ ModifyRequest = (*modifyRequestImpl)(nil)

type modifyAction struct {
	Action string `json:"action"`
	ItemID string `json:"item_id"`
	Time   string `json:"time"`
}

func (r *modifyRequestImpl) Delete(itemID ...string) ModifyRequest {
	for _, id := range itemID {
		r.Actions = append(r.Actions, modifyAction{
			Action: "delete",
			ItemID: id,
		})
	}

	return r
}

func (r *modifyRequestImpl) Favorite(itemID ...string) ModifyRequest {
	for _, id := range itemID {
		r.Actions = append(r.Actions, modifyAction{
			Action: "favorite",
			ItemID: id,
		})
	}

	return r
}

type ModifyResponse struct {
	ActionResults []bool `json:"action_results"`
	Status        int    `json:"status"`
}

func (r *modifyRequestImpl) Do(ctx context.Context) (*ModifyResponse, error) {
	params := &struct {
		ConsumerKey string         `json:"consumer_key"`
		AccessToken string         `json:"access_token"`
		Actions     []modifyAction `json:"actions,omitempty"`
	}{
		ConsumerKey: r.client.consumerKey,
		AccessToken: r.client.accessToken,
		Actions:     r.Actions,
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	for i := range params.Actions {
		params.Actions[i].Time = timestamp
	}

	resp, err := r.client.sendRequest(ctx, r.client.post("/send").JSON(params))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	response, err := goxp.ReadJSON[ModifyResponse](resp.Body)
	if err != nil {
		return nil, err
	}

	return response, nil
}

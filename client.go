package getpocket

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/whitekid/goxp/request"
)

//go:generate mockery --name Interface
type Interface interface {
	Get() *GetRequest
	Add(url string) AddRequester
	Modify() ModifyRequester
}

type clientImpl struct {
	client request.Interface

	consumerKey string
	accessToken string
}

var _ Interface = (*clientImpl)(nil)

func New(consumerKey string, accessToken string) Interface {
	return &clientImpl{
		consumerKey: consumerKey,
		accessToken: accessToken,
		client:      request.NewSession(nil),
	}
}

func (c *clientImpl) do(ctx context.Context, req *request.Request) (*request.Response, error) {
	resp, err := req.Do(ctx)
	if err != nil {
		return resp, errors.Wrap(err, "request failed")
	}

	if !resp.Success() {
		message := resp.Header.Get("x-error")
		code := resp.Header.Get("x-error-code")

		return nil, errors.Errorf("request failed with status %d: code=%s, message=%s",
			resp.StatusCode, code, message)
	}

	return resp, err
}

func (client *clientImpl) Get() *GetRequest {
	r := &GetRequest{
		client: client,
	}
	r.params.ConsumerKey = r.client.consumerKey
	r.params.AccessToken = r.client.accessToken

	return r
}

type GetRequest struct {
	client *clientImpl

	params struct {
		ConsumerKey string  `json:"consumer_key"`
		AccessToken string  `json:"access_token"`
		State       string  `json:"state"`
		DetailType  string  `json:"detailType"`
		Favorite    *string `json:"favorite,omitempty"`
		Sort        *string `json:"sort,omitempty"`
		Since       *string `json:"since,omitempty"`
		Count       *string `json:"count,omitempty"`
	}
}

type ArticleMap map[string]Article

func (r *GetRequest) Do(ctx context.Context) (*GetResponse, error) {
	req := r.client.client.Post("https://getpocket.com/v3/get").
		Header("X-Accept", "application/json").
		JSON(&r.params)

	resp, err := r.client.do(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get failed")
	}
	defer resp.Body.Close()

	var response GetResponse
	if err := resp.JSON(&response); err != nil {
		return nil, errors.Wrap(err, "fail to parse get response")
	}

	return &response, nil
}

func (r *GetRequest) Favorite(favorite bool) *GetRequest {
	var s string
	if favorite {
		s = "1"
	} else {
		s = "0"
	}
	r.params.Favorite = &s
	return r
}

type Order string

const (
	Newest Order = "newest"
)

func (r *GetRequest) Sort(sort Order) *GetRequest {
	s := string(sort)
	r.params.Sort = &s
	return r
}

func (r *GetRequest) Since(tm time.Time) *GetRequest {
	since := strconv.FormatInt(tm.Unix(), 10)
	r.params.Since = &since

	return r
}

func (r *GetRequest) Count(count int) *GetRequest {
	s := strconv.FormatInt(int64(count), 10)
	r.params.Count = &s
	return r
}

type GetResponse struct {
	Status   int                `json:"status"`
	Complete int                `json:"complete"`
	List     map[string]Article `json:"-"` // [] or {}
}

func (a *GetResponse) UnmarshalJSON(data []byte) error {
	type Response GetResponse
	var response Response

	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}
	*a = GetResponse(response)

	// extra
	var respExtra struct {
		List interface{} `json:"list"`
	}
	if err := json.Unmarshal(data, &respExtra); err != nil {
		return nil
	}

	if _, ok := respExtra.List.([]interface{}); ok {
		a.List = map[string]Article{}
	} else {
		if err := reEncode(respExtra.List, &a.List); err != nil {
			return err
		}
	}

	return nil
}

// func parseGetResponse(r io.Reader) (*GetResponse, error) {
// 	resp := &GetResponse{}

// 	if err := json.NewDecoder(r).Decode(resp); err != nil {
// 		return nil, errors.Wrap(err, "fail to decode response")
// 	}

// 	// article이 없으면 List가 []이고 있으면 map[string]Article이라서.. 꽁수
// 	if _, ok := resp.List.([]interface{}); ok {
// 		resp.List = ArticleMap{}
// 	} else {
// 		items := ArticleMap{}
// 		if err := reEncode(resp.List, &items); err != nil {
// 			return nil, err
// 		}

// 		resp.List = items
// 	}

// 	return resp, nil
// }

type Article struct {
	ItemID            string           `json:"item_id"`
	ResolvedID        string           `json:"resolved_id"`
	NormalURL         string           `json:"normal_url"`
	GivenURL          string           `json:"given_url"`
	GivelTitle        string           `json:"given_title"`
	Favorite          string           `json:"favorite"`
	Status            string           `json:"status"`
	ResolvedTitle     string           `json:"resolved_title"`
	ResolvedURL       string           `json:"resolved_url"`
	ResolvedNormalURL string           `json:"resolved_normal_url"`
	Excerpt           string           `json:"excerpt"`
	IsArticle         string           `json:"is_article"`
	HasVideo          string           `json:"has_video"`
	HasImage          string           `json:"has_image"`
	WordCount         string           `json:"word_count"`
	TopImageURL       string           `json:"top_image_url"`
	Images            map[string]Image `json:"-"`
	Authors           interface{}      `json:"authors"`
	Videos            map[string]Video `json:"-"`
}

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

	if _, ok := extras.Images.([]interface{}); ok {
		a.Images = make(map[string]Image)
	} else {
		if err := reEncode(extras.Images, &a.Images); err != nil {
			return err
		}
	}

	if _, ok := extras.Videos.([]interface{}); ok {
		a.Videos = make(map[string]Video)
	} else {
		if err := reEncode(extras.Videos, &a.Videos); err != nil {
			return err
		}
	}

	return nil
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

func (client *clientImpl) Add(url string) AddRequester {
	r := &AddRequest{client: client}

	r.params.ConsumerKey = r.client.consumerKey
	r.params.AccessToken = r.client.accessToken
	r.params.URL = url

	return r
}

//go:generate mockery --name AddRequester
type AddRequester interface {
	Do(ctx context.Context) (*AddResponse, error)
}

type AddRequest struct {
	client *clientImpl

	params struct {
		ConsumerKey string  `json:"consumer_key"`
		AccessToken string  `json:"access_token"`
		URL         string  `json:"url"`
		Title       *string `json:"title,omitempty"`
		Tags        *string `json:"tags,omitempty"`
	}
}

var _ AddRequester = (*AddRequest)(nil)

func (r *AddRequest) Do(ctx context.Context) (*AddResponse, error) {
	req := r.client.client.Post("https://getpocket.com/v3/add").
		Header("X-Accept", "application/json").
		JSON(&r.params)

	resp, err := r.client.do(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "add failed")
	}

	var response AddResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "fail to parse response")
	}

	return &response, nil
}

func (r *AddRequest) Title(title string) *AddRequest {
	r.params.Title = &title
	return r
}

func (r *AddRequest) Tags(tags string) *AddRequest {
	r.params.Tags = &tags
	return r
}

func reEncode(v interface{}, nv interface{}) error {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		return nil
	}

	if err := json.NewDecoder(&buf).Decode(nv); err != nil {
		return err
	}

	return nil
}

type AddResponse struct {
	Items  Article `json:"item"`
	Status int     `json:"status"`
}

func (client *clientImpl) Modify() ModifyRequester {
	r := &ModifyRequest{
		client: client,
	}
	r.params.ConsumerKey = r.client.consumerKey
	r.params.AccessToken = r.client.accessToken

	return r
}

//go:generate mockery --name ModifyRequester
type ModifyRequester interface {
	Favorite(itemID string) ModifyRequester
	Delete(itemID string) ModifyRequester

	Do(ctx context.Context) (*ModifyResponse, error)
}
type ModifyRequest struct {
	client *clientImpl

	params struct {
		ConsumerKey string         `json:"consumer_key"`
		AccessToken string         `json:"access_token"`
		Actions     []modifyAction `json:"actions,omitempty"`
	}
}

var _ ModifyRequester = (*ModifyRequest)(nil)

type modifyAction struct {
	Action string `json:"action"`
	ItemID string `json:"item_id"`
	Time   string `json:"time"`
}

func (r *ModifyRequest) Delete(itemID string) ModifyRequester {
	r.params.Actions = append(r.params.Actions, modifyAction{
		Action: "delete",
		ItemID: itemID,
	})
	return r
}

func (r *ModifyRequest) Favorite(itemID string) ModifyRequester {
	r.params.Actions = append(r.params.Actions, modifyAction{
		Action: "favorite",
		ItemID: itemID,
	})
	return r
}

func (r *ModifyRequest) Do(ctx context.Context) (*ModifyResponse, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	for i := range r.params.Actions {
		r.params.Actions[i].Time = timestamp
	}

	req := r.client.client.Post("https://getpocket.com/v3/send").
		Header("X-Accept", "application/json").
		JSON(&r.params)

	resp, err := r.client.do(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "modify request failed")
	}

	defer resp.Body.Close()
	response := &ModifyResponse{}
	if err := resp.JSON(response); err != nil {
		return nil, errors.Wrap(err, "fail to parser response")
	}

	return response, nil
}

type ModifyResponse struct {
	ActionResults []bool `json:"action_results"`
	Status        int    `json:"status"`
}

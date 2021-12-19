package lemmyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	HTTPClient *http.Client
	BaseUrl    string
}

func (c *Client) GetPosts(ctx context.Context, community string, page int) ([]PostView, error) {
	var query = fmt.Sprintf("%s/api/v3/post/list?page=%d", c.BaseUrl, page)
	if len(community) != 0 {
		query += "&community_name=" + community
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create http request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch posts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("could not fetch posts: status_code=%d", resp.StatusCode)
	}

	var out GetPostsResponse
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, fmt.Errorf("could not json decode posts response: %w", err)
	}

	return out.Posts, nil
}
func (c *Client) GetComments(ctx context.Context, id int) ([]CommentView, error) {
	var query = fmt.Sprintf("%s/api/v3/post?id=%d", c.BaseUrl, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create http request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch comments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("could not fetch comments: status_code=%d", resp.StatusCode)
	}

	var out GetPostResponse
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, fmt.Errorf("could not json decode comments response: %w", err)
	}

	return out.Comments, nil
}

type GetPostsResponse struct {
	Posts []PostView `json:"posts"`
}
type GetPostResponse struct {
	Comments []CommentView `json:"comments"`
}
type CommentView struct {
	Comment                    Comment           `json:"comment"`
	Creator                    Creator           `json:"creator"`
	Community                  Community         `json:"community"`
	CreatorBannedFromCommunity bool              `json:"creator_banned_from_community"`
	Counts                     CommentAggregates `json:"counts"`
	Subscribed                 bool              `json:"subscribed"`
	Saved                      bool              `json:"saved"`
	Read                       bool              `json:"read"`
	CreatorBlocked             bool              `json:"creator_blocked"`
	MyVote                     int               `json:"my_vote"`
}

type PostView struct {
	Post                       Post           `json:"post"`
	Creator                    Creator        `json:"creator"`
	Community                  Community      `json:"community"`
	CreatorBannedFromCommunity bool           `json:"creator_banned_from_community"`
	Counts                     PostAggregates `json:"counts"`
	Subscribed                 bool           `json:"subscribed"`
	Saved                      bool           `json:"saved"`
	Read                       bool           `json:"read"`
	CreatorBlocked             bool           `json:"creator_blocked"`
	MyVote                     int            `json:"my_vote"`
}

type Comment struct {
	ID        int    `json:"id"`
	CreatorID int    `json:"creator_id"`
	PostID    int    `json:"post_id"`
	ParentID  *int   `json:"parent_id,omitempty"`
	Content   string `json:"content"`
	Removed   bool   `json:"removed"`
	Read      bool   `json:"read"`
	Published string `json:"published"`
	Updated   string `json:"updated"`
	Deleted   bool   `json:"deleted"`
	ApID      string `json:"ap_id"`
	Local     bool   `json:"local"`
}

type Post struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	Body             string `json:"body"`
	CreatorID        int    `json:"creator_id"`
	CommunityID      int    `json:"community_id"`
	Removed          bool   `json:"removed"`
	Locked           bool   `json:"locked"`
	Published        string `json:"published"`
	Updated          string `json:"updated"`
	Deleted          bool   `json:"deleted"`
	Nsfw             bool   `json:"nsfw"`
	Stickied         bool   `json:"stickied"`
	EmbedTitle       string `json:"embed_title"`
	EmbedDescription string `json:"embed_description"`
	EmbedHTML        string `json:"embed_html"`
	ThumbnailURL     string `json:"thumbnail_url"`
	ApID             string `json:"ap_id"`
	Local            bool   `json:"local"`
}
type Creator struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Avatar         string `json:"avatar"`
	Banned         bool   `json:"banned"`
	Published      string `json:"published"`
	Updated        string `json:"updated"`
	ActorID        string `json:"actor_id"`
	Bio            string `json:"bio"`
	Local          bool   `json:"local"`
	Banner         string `json:"banner"`
	Deleted        bool   `json:"deleted"`
	InboxURL       string `json:"inbox_url"`
	SharedInboxURL string `json:"shared_inbox_url"`
	MatrixUserID   string `json:"matrix_user_id"`
	Admin          bool   `json:"admin"`
	BotAccount     bool   `json:"bot_account"`
}

type Community struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Removed     bool   `json:"removed"`
	Published   string `json:"published"`
	Updated     string `json:"updated"`
	Deleted     bool   `json:"deleted"`
	Nsfw        bool   `json:"nsfw"`
	ActorID     string `json:"actor_id"`
	Local       bool   `json:"local"`
	Icon        string `json:"icon"`
	Banner      string `json:"banner"`
}
type PostAggregates struct {
	ID                     int    `json:"id"`
	PostID                 int    `json:"post_id"`
	Comments               int    `json:"comments"`
	Score                  int    `json:"score"`
	Upvotes                int    `json:"upvotes"`
	Downvotes              int    `json:"downvotes"`
	Stickied               bool   `json:"stickied"`
	Published              string `json:"published"`
	NewestCommentTimeNecro string `json:"newest_comment_time_necro"`
	NewestCommentTime      string `json:"newest_comment_time"`
}
type CommentAggregates struct {
	ID        int    `json:"id"`
	CommentID int    `json:"comment_id"`
	Score     int    `json:"score"`
	Upvotes   int    `json:"upvotes"`
	Downvotes int    `json:"downvotes"`
	Published string `json:"published"`
}

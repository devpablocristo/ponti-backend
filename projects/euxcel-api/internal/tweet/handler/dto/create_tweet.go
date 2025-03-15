package dto

// Response
type CreateTweetResponse struct {
	Message string `json:"message"`
	TweetID string `json:"tweet_id"`
}

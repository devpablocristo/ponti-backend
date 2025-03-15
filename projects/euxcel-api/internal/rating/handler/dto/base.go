package dto

type Rating struct {
	ID       int    `json:"id"`
	Item     string `json:"item"`
	Score    int    `json:"score"`
	Comments string `json:"comments"`
}

// func ToRating(dto Rating) *domain.Rating {
// 	return &domain.Rating{
// 		Item:     dto.Item,
// 		Score:    dto.Score,
// 		Comments: dto.Comments,
// 	}
// }

// func ToRating(r *domain.Rating) Rating {
// 	return Rating{
// 		Item:     r.Item,
// 		Score:    r.Score,
// 		Comments: r.Comments,
// 	}
// }

package elasticsearch

type esAdDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Category    string `json:"category"`
	MainImage   string `json:"main_image"`
}

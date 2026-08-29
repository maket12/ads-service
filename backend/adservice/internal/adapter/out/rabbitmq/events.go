package rabbitmq

type AdPublishedEvent struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Category    string `json:"category"`
	MainImage   string `json:"main_image"`
}

type AdUpdatedEvent struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Category    string `json:"category"`
	MainImage   string `json:"main_image"`
}

type AdDeletedEvent struct {
	ID string `json:"id"`
}

type AdRejectedEvent struct {
	ID string `json:"id"`
}

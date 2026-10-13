package icon

const (
	IconSourceLucide      = "lucide"
	IconSourceSimpleIcons = "simpleicons"
)

type Icon struct {
	Id     string   `json:"id" bson:"_id" example:"lucide:a-arrow-down"`
	Name   string   `json:"name" bson:"name" example:"a-arrow-down"`
	Title  string   `json:"title" bson:"title" example:"A Arrow Down"`
	Source string   `json:"source" bson:"source" example:"lucide"`
	Tags   []string `json:"tags" bson:"tags"`
	Hex    string   `json:"hex,omitempty" bson:"hex,omitempty" example:"ECD53F"`
}

type IconFilter struct {
	Source string `form:"source,omitempty" binding:"omitempty,oneof=lucide simpleicons"`
}

type IconSearch struct {
	Query   string   `form:"q" binding:"max=100"`
	Sources []string `form:"source" binding:"dive,oneof=lucide simpleicons"`
	Tag     string   `form:"tag,omitempty" binding:"max=100"`
	Limit   int      `form:"limit,default=50" binding:"min=1,max=500"`
	Offset  int      `form:"offset,default=0" binding:"min=0"`
}

type IconSearchResult struct {
	Total int    `json:"total" example:"42"`
	Items []Icon `json:"items"`
}

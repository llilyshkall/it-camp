package model

import (
	"time"
)

type Error struct {
	Error interface{} `json:"error,omitempty"`
}

type Response struct {
	Body interface{} `json:"body,omitempty"`
}

// type Comment struct {
// 	ID        int    `json:"itemid,omitempty"`
// 	ProjectID int    `json:"userid"`
// 	Direction string `json:"direction"`
// 	Section   string `json:"section,omitempty"`
// 	Text      string `json:"text"`
// 	Urgency   string `json:"urgency"`
// }

type Remark struct {
	ID         int       `json:"id"`
	ProjectID  int       `json:"project_id"`
	Direction  string    `json:"direction"`
	Section    string    `json:"section"`
	Subsection string    `json:"subsection"`
	Content    string    `json:"content"`
	Urgency    string    `json:"urgency"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

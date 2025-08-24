package models

import "io"

type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

type Attach struct {
	Type    string
	File    File
	FileExt string
}

type UploadAttachResponse struct {
	File string `json:"file"`
}

type Error struct {
	Error interface{} `json:"error,omitempty"`
}

type Response struct {
	Body interface{} `json:"body,omitempty"`
}

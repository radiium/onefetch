package model

import "time"

type FSNode struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	ModTime    time.Time `json:"modTime"`
	IsDir      bool      `json:"isDir"`
	IsReadOnly bool      `json:"isReadOnly"`
	IsHidden   bool      `json:"isHidden"`
	IsTmp      bool      `json:"isTmp"`
	Children   []FSNode  `json:"children,omitempty"`
}

type CreateDirRequest struct {
	Path    string `json:"path"`
	DirName string `json:"dirname"`
}

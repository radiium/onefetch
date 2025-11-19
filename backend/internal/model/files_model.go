package model

type FSNode struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	IsDir      bool     `json:"isDir"`
	IsReadOnly bool     `json:"isReadOnly"`
	IsHidden   bool     `json:"isHidden"`
	IsTmp      bool     `json:"isTmp"`
	Children   []FSNode `json:"children,omitempty"`
}

type CreateDirRequest struct {
	Path    string `json:"path"`
	DirName string `json:"dirname"`
}

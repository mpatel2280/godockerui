package model

type Container struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	State   string `json:"state"`
	Status  string `json:"status"`
	Command string `json:"command"`
}

type Image struct {
	ID       string `json:"id"`
	RepoTags string `json:"repoTags"`
	SizeMB   int64  `json:"sizeMB"`
	Created  int64  `json:"created"`
}

type Dashboard struct {
	ContainersTotal int  `json:"containersTotal"`
	ContainersUp    int  `json:"containersUp"`
	ContainersDown  int  `json:"containersDown"`
	ImagesTotal     int  `json:"imagesTotal"`
	Simulated       bool `json:"simulated"`
}

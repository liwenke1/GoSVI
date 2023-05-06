package config

const (
	SemverFilePath   = "D:\\Code\\SemanticVersionStudy-data\\dataset\\semver\\semver_combine\\combine"
	SemverReportPath = "D:\\Code\\SemanticVersionStudy-data\\impact-go\\report\\impact"
)

type TaskList struct {
	Count int    `json:"total_count"`
	Items []Item `json:"items"`
}

type Item struct {
	Url    string        `json:"Url"`
	Detail []VersionInfo `json:"Detail"`
}

type VersionInfo struct {
	Version string   `json:"Version"`
	Path 	string 	 `json:"Path"`
	PkgInfo []string `json:"PkgInfo"`
}

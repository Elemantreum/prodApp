package section

type Section struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Logo  string `json:"logo"`
	Admin bool   `json:"admin"`
}

type SectionWithVenues struct {
	Section
	Venues []string `json:"venues"`
}

type Directory struct {
	Sections    []SectionWithVenues `json:"sections"`
	SystemAdmin bool                `json:"systemAdmin"`
}

type Member struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Admin     bool   `json:"admin"`
}

type MembershipChange struct {
	SectionID int  `json:"sectionId"`
	UserID    int  `json:"userId"`
	Admin     bool `json:"admin"`
	Remove    bool `json:"remove"`
}

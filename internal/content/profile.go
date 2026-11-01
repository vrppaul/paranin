package content

// Profile is the résumé: who, what they did and made, what they can do, and how to reach them.
//
// A text shown to people may carry two inline marks, which the faces draw: ==text== a result worth
// highlighting, and [text] a gap the author has still to fill.
type Profile struct {
	Site        Site            `yaml:"site"`
	Person      Person          `yaml:"person"`
	Summary     string          `yaml:"summary"`
	Facts       []Fact          `yaml:"facts"`
	Experience  []Role          `yaml:"experience"`
	Projects    []Project       `yaml:"projects"`
	SkillGroups []SkillGroup    `yaml:"skill_groups"`
	Skills      []Skill         `yaml:"skills"`
	Education   []School        `yaml:"education"`
	Languages   []Language      `yaml:"languages"`
	Health      []HealthSection `yaml:"health"`
	Contact     Contact         `yaml:"contact"`
}

// Site is where the résumé is found: its address, with no slash at the end, and the command that reaches it.
type Site struct {
	BaseURL string `yaml:"base_url"`
	SSH     string `yaml:"ssh"`
}

// Person is who the résumé is about.
type Person struct {
	Name           string       `yaml:"name"`
	GivenName      string       `yaml:"given_name"`
	AlternateNames []string     `yaml:"alternate_names"`
	Title          string       `yaml:"title"`
	Headline       string       `yaml:"headline"`
	Location       string       `yaml:"location"`
	Email          string       `yaml:"email"`
	Links          []SocialLink `yaml:"links"`
}

// SocialLink is a public profile: a short key for lists of keys and values, a label, its address, and how the
// address is written out.
type SocialLink struct {
	Key   string `yaml:"key"`
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
	Text  string `yaml:"text"`
}

// Fact is a label and a value.
type Fact struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

// Role is one place of work. A role still held ends "current".
type Role struct {
	Company  string   `yaml:"company"`
	Title    string   `yaml:"title"`
	Location string   `yaml:"location"`
	Start    Month    `yaml:"start"`
	End      Month    `yaml:"end"`
	Bullets  []string `yaml:"bullets"`
}

// Current tells a role still held.
func (role Role) Current() bool {
	return role.End.IsCurrent()
}

// Project is a piece of work that can be checked. Its slug names it in every address; a project with a page of
// its own gets one.
type Project struct {
	Slug        string    `yaml:"slug"`
	Name        string    `yaml:"name"`
	Tagline     string    `yaml:"tagline"`
	Description string    `yaml:"description"`
	Language    string    `yaml:"language"`
	Repo        string    `yaml:"repo"`
	Install     string    `yaml:"install"`
	Recording   string    `yaml:"recording"`
	Page        bool      `yaml:"page"`
	Facts       []Fact    `yaml:"facts"`
	Sections    []Section `yaml:"sections"`
}

// Section is a titled block of paragraphs on a project's page.
type Section struct {
	Title      string   `yaml:"title"`
	Paragraphs []string `yaml:"paragraphs"`
}

// SkillGroup is a group skills are shown in; the profile lists the groups in their order.
type SkillGroup struct {
	Key   string `yaml:"key"`
	Title string `yaml:"title"`
}

// Skill is something the author can do: its group, how long, the words a job description may use for it, and
// where it was proven.
type Skill struct {
	Name     string   `yaml:"name"`
	Group    string   `yaml:"group"`
	Duration string   `yaml:"duration"`
	Aliases  []string `yaml:"aliases"`
	Sources  []string `yaml:"sources"`
	Learning bool     `yaml:"learning"`
}

// School is a place of study; either month may be a gap.
type School struct {
	Name  string `yaml:"name"`
	Field string `yaml:"field"`
	Start Month  `yaml:"start"`
	End   Month  `yaml:"end"`
}

// Language is a spoken language and how well.
type Language struct {
	Name  string `yaml:"name"`
	Level string `yaml:"level"`
}

// HealthSection groups the checks of the hidden :checkhealth report.
type HealthSection struct {
	Name   string        `yaml:"name"`
	Checks []HealthCheck `yaml:"checks"`
}

// HealthCheck is one line of the report.
type HealthCheck struct {
	Level  Level  `yaml:"level"`
	Text   string `yaml:"text"`
	Advice string `yaml:"advice"`
}

// Level is how a health check reads.
type Level string

const (
	LevelOK      Level = "ok"
	LevelWarning Level = "warn"
	LevelError   Level = "error"
)

// IsKnown tells a level of one of the three kinds.
func (level Level) IsKnown() bool {
	switch level {
	case LevelOK, LevelWarning, LevelError:
		return true
	default:
		return false
	}
}

// Contact is what the contact form offers: what a message may be about, and the salary range it asks for.
type Contact struct {
	About  []string `yaml:"about"`
	Salary Salary   `yaml:"salary"`
}

// Salary is the range the form's slider offers, in whole units of a currency per period. A maximum above SoftMax
// means no upper limit; Step moves the slider, BigStep moves it far.
type Salary struct {
	Min        int      `yaml:"min"`
	SoftMax    int      `yaml:"soft_max"`
	Step       int      `yaml:"step"`
	BigStep    int      `yaml:"big_step"`
	Currencies []string `yaml:"currencies"`
	Period     string   `yaml:"period"`
}

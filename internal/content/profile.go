package content

// The marks a text of the profile may carry, which the faces draw: ==text== a result worth highlighting, and
// [text] a gap the author has still to fill. Prose and titles carry them: a name, a title, a line, a point, a
// value, a duration. A label, a place, a language's name, a stack's name, an address, a link's text and a command
// are shown as written. A gap is also how a month not known yet is written
const (
	HighlightMark = "=="
	GapOpening    = "["
	GapClosing    = "]"
)

// Profile is the résumé: who, what they did and made, what they can do, and how to reach them.
type Profile struct {
	Site        Site            `yaml:"site"`
	Person      Person          `yaml:"person"`
	Summary     string          `yaml:"summary"`
	Facts       []Fact          `yaml:"facts"`
	Experience  []Role          `yaml:"experience"`
	Projects    []Project       `yaml:"projects"`
	SkillGroups []SkillGroup    `yaml:"skill_groups"`
	Education   []School        `yaml:"education"`
	Languages   []Language      `yaml:"languages"`
	Health      []HealthSection `yaml:"health"`
	Contact     Contact         `yaml:"contact"`
}

// Site is where the résumé is found: its address, with no slash at the end, and the command that reaches it.
type Site struct {
	BaseURL WebAddress `yaml:"base_url"`
	SSH     string     `yaml:"ssh"`
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

// SocialLink is a public profile: its label, its address, and how the address is written out.
type SocialLink struct {
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

// Project is a piece of work that can be checked. Its slug names it in every address; a project with a page of
// its own gets one. Its stack is what it is built on, by name, after the language it is written in.
type Project struct {
	Slug        string    `yaml:"slug"`
	Name        string    `yaml:"name"`
	Tagline     string    `yaml:"tagline"`
	Description string    `yaml:"description"`
	Language    string    `yaml:"language"`
	Stack       []string  `yaml:"stack"`
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

// SkillGroup is a group of skills, under its title; the profile lists the groups, and the skills in each, in
// their order.
type SkillGroup struct {
	Title  string  `yaml:"title"`
	Skills []Skill `yaml:"skills"`
}

// Skill is something the author can do: how long, the words a job description may use for it, and where it was
// proven.
type Skill struct {
	Name     string   `yaml:"name"`
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

// levels are the levels a check may read as, each named once. exhaustive holds the literal to every Level, so a
// level added to the constants is added here too.
func levels() map[Level]bool {
	return map[Level]bool{LevelOK: true, LevelWarning: true, LevelError: true}
}

// IsKnown tells a level of one of the kinds the levels name.
func (level Level) IsKnown() bool {
	return levels()[level]
}

// Contact is what the contact form offers: what a message may be about, and the salary range it asks for.
type Contact struct {
	About  []string `yaml:"about"`
	Salary Salary   `yaml:"salary"`
}

// Salary is the range the form offers, in whole units of a currency per period. A maximum above SoftMax means no
// upper limit. How far a slider moves is the form's own.
type Salary struct {
	Min        int      `yaml:"min"`
	SoftMax    int      `yaml:"soft_max"`
	Currencies []string `yaml:"currencies"`
	Period     string   `yaml:"period"`
}

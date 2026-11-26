package content

// The marks a text of the profile may carry, which the faces draw: ==text== a result worth highlighting, and
// [text] a gap the author has still to fill. Prose and titles carry them: a name, a title, a line, a point, a
// description, a value. A label, a place, a language's name, a stack's or a technology's name, an address, a link's
// text and a command are shown as written. A gap is also how a duration not known yet is written, the gap its whole
// value
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
	Experience  Experience      `yaml:"experience"`
	Projects    []Project       `yaml:"projects"`
	SkillGroups []SkillGroup    `yaml:"skill_groups"`
	Languages   []Language      `yaml:"languages"`
	Health      []HealthSection `yaml:"health"`
	Contact     Contact         `yaml:"contact"`
}

// Site is where the résumé is found: its address, with no slash at the end, and the host a terminal reaches it at over
// ssh, which every ssh command the site writes names; and the tag authority the feeds and their posts are named by for
// good, wherever the site moves (RFC 4151).
type Site struct {
	BaseURL      WebAddress `yaml:"base_url"`
	SSHHost      string     `yaml:"ssh_host"`
	TagAuthority string     `yaml:"tag_authority"`
}

// Person is who the résumé is about.
type Person struct {
	Name           string       `yaml:"name"`
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

// Experience is the author's work: its line, the one thing a reader should keep of it; the skills it puts forward, by
// their names in the skill groups, which give how long each was used; and the roles, in their order.
type Experience struct {
	Line       string   `yaml:"line"`
	PutForward []string `yaml:"put_forward"`
	Roles      []Role   `yaml:"roles"`
}

// Role is one place of work: its points, and its description, paragraphs that tell it at length. A role still held
// ends "current".
type Role struct {
	Company     string   `yaml:"company"`
	Title       string   `yaml:"title"`
	Start       Month    `yaml:"start"`
	End         Month    `yaml:"end"`
	Bullets     []string `yaml:"bullets"`
	Description []string `yaml:"description"`
}

// Project is a piece of work that can be checked. Its slug names it in every address; a project with a page of
// its own gets one. Its stack is what it is built on, by name, after the language it is written in; its licence a
// word, as SPDX names one.
type Project struct {
	Slug        string    `yaml:"slug"`
	Name        string    `yaml:"name"`
	Tagline     string    `yaml:"tagline"`
	Description string    `yaml:"description"`
	Language    string    `yaml:"language"`
	Stack       []string  `yaml:"stack"`
	Repo        string    `yaml:"repo"`
	Install     *Install  `yaml:"install"`
	Release     *Release  `yaml:"release"`
	Licence     string    `yaml:"licence"`
	Page        bool      `yaml:"page"`
	Sections    []Section `yaml:"sections"`
}

// Install is how a project is installed: its command, and the other ways, each taken instead where its note says.
type Install struct {
	Command      string        `yaml:"command"`
	Alternatives []Alternative `yaml:"alternatives"`
}

// Alternative is another way to install a project: a note on when to take it instead, and its command.
type Alternative struct {
	Note    string `yaml:"note"`
	Command string `yaml:"command"`
}

// Release is where a project's releases are read: its newest, and the history of them all.
type Release struct {
	Latest    WebAddress `yaml:"latest"`
	Changelog WebAddress `yaml:"changelog"`
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

// Skill is something the author can do: how long, and where it was proven.
type Skill struct {
	Name     string   `yaml:"name"`
	Duration Duration `yaml:"duration"`
	Sources  []string `yaml:"sources"`
	Learning bool     `yaml:"learning"`
}

// Language is a language the author speaks.
type Language struct {
	Name string `yaml:"name"`
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

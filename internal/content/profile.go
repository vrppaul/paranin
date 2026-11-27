package content

import "time"

// The marks a text of the profile may carry, which the faces draw: ==text== a result worth highlighting, [text] a gap
// the author has still to fill, and `text` code, inside which no mark is read. Prose and titles carry them: a name, a
// title, a line, a point, a paragraph, a description, a value. A label, a place, a language's name, a stack's or a
// technology's name, an address, a link's text and a command are shown as written. A gap is also how a duration not
// known yet is written, the gap its whole value
const (
	HighlightMark = "=="
	GapOpening    = "["
	GapClosing    = "]"
	CodeMark      = "`"
)

// Profile is the résumé: who, what they did and made, what they can do, and how to reach them.
type Profile struct {
	Site        Site            `yaml:"site"`
	Person      Person          `yaml:"person"`
	Summary     string          `yaml:"summary"`
	Facts       []Fact          `yaml:"facts"`
	Experience  Experience      `yaml:"experience"`
	Projects    Projects        `yaml:"projects"`
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

// Person is who the résumé is about: among the rest, the city they are based in, as the page writes it, and where they
// work from.
type Person struct {
	Name           string        `yaml:"name"`
	AlternateNames []string      `yaml:"alternate_names"`
	Title          string        `yaml:"title"`
	Headline       string        `yaml:"headline"`
	HeadlineLinks  HeadlineLinks `yaml:"headline_links"`
	Location       string        `yaml:"location"`
	WorksFrom      string        `yaml:"works_from"`
	Email          string        `yaml:"email"`
	Links          []SocialLink  `yaml:"links"`
}

// HeadlineLinks are the words of the headline that name the author's two kinds of work, each once, each a link to the
// tab that holds it: the experience's and the projects'.
type HeadlineLinks struct {
	Experience string `yaml:"experience"`
	Projects   string `yaml:"projects"`
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

// Projects are the author's projects: their line, the one thing a reader should keep of them, as the experience has
// its own; and the projects, in their order.
type Projects struct {
	Line string    `yaml:"line"`
	List []Project `yaml:"list"`
}

// Project is a piece of work that can be checked. Its slug names it in every address; its level says how the
// projects' tab shows it. Its stack is what it is built on, by name, after the language it is written in; its licence
// a word, as SPDX names one. A featured project's page tells its story in its sections, then the details a user looks
// up.
type Project struct {
	Slug        string       `yaml:"slug"`
	Name        string       `yaml:"name"`
	Level       ProjectLevel `yaml:"level"`
	Tagline     string       `yaml:"tagline"`
	Description string       `yaml:"description"`
	Language    string       `yaml:"language"`
	Stack       []string     `yaml:"stack"`
	Repo        string       `yaml:"repo"`
	Install     *Install     `yaml:"install"`
	Release     *Release     `yaml:"release"`
	Licence     string       `yaml:"licence"`
	Sections    []Section    `yaml:"sections"`
	Details     []Section    `yaml:"details"`
}

// ProjectLevel is how far the projects' tab puts a project forward (DESIGN.md, ProjectList).
type ProjectLevel string

const (
	// ProjectFeatured is a project with a page of its own, a detailed tile on the tab leading there
	ProjectFeatured ProjectLevel = "featured"
	// ProjectAuxiliary is a small tile on the tab, leading to the project's repository
	ProjectAuxiliary ProjectLevel = "auxiliary"
	// ProjectTold is a project told in words alone, with no link
	ProjectTold ProjectLevel = "told"
)

// projectLevels are the levels a project may stand at, each named once. exhaustive holds the literal to every
// ProjectLevel, so a level added to the constants is added here too.
func projectLevels() map[ProjectLevel]bool {
	return map[ProjectLevel]bool{ProjectFeatured: true, ProjectAuxiliary: true, ProjectTold: true}
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

// Section is a titled part of a project's page: its paragraphs, then its points; and what adds to it, its clip and its
// links.
type Section struct {
	Title      string   `yaml:"title"`
	Paragraphs []string `yaml:"paragraphs"`
	Points     []string `yaml:"points"`
	Clip       *Clip    `yaml:"clip"`
	Links      []Link   `yaml:"links"`
}

// Clip is a recording of a project at work a section of its page plays (DESIGN.md, ProjectSection): its file, an MP4,
// and its still, a picture, both files of the project's own folder, and a line on what it shows. The load reads both
// files whole, and the recording's length from its file.
type Clip struct {
	File  string `yaml:"file"`
	Still string `yaml:"still"`
	Shows string `yaml:"shows"`
	// Recording and Picture are the clip's two files as the load reads them
	Recording Recording `yaml:"-"`
	Picture   Picture   `yaml:"-"`
}

// Recording is a recording a project's page plays, read whole as the content loads: the type it is sent as, how long
// it runs, read from the file, and what it holds.
type Recording struct {
	MediaType string
	Length    time.Duration
	Body      []byte
}

// Link is a place outside the site a section of a project's page leads to: its name, its address and a line on what
// a reader finds there.
type Link struct {
	Name string     `yaml:"name"`
	URL  WebAddress `yaml:"url"`
	Line string     `yaml:"line"`
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

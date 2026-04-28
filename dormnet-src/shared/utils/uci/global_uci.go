package uci

var rootUci UciContext

var GetString func(section string, option string, defVal string) string
var LookupString func(section string, option string) (string, bool)

var GetInt func(section string, option string, defVal int) int
var LookupInt func(section string, option string) (int, bool)

var GetBool func(section string, option string, defVal bool) bool
var LookupBool func(section string, option string) (bool, bool)

var GetList func(section string, option string) []string

var GetSections func(typ string) []string

var CreateSection func(typ string, section string) bool
var Set func(section string, option string, value ...string) bool
var Add func(section string, option string, value string) bool

func InitUci(config string) {
	rootUci = NewUciConfig(config)

	GetString = rootUci.GetString
	LookupString = rootUci.LookupString
	GetInt = rootUci.GetInt
	LookupInt = rootUci.LookupInt
	GetBool = rootUci.GetBool
	LookupBool = rootUci.LookupBool
	GetList = rootUci.GetList
	GetSections = rootUci.GetSections
	CreateSection = rootUci.CreateSection
	Set = rootUci.Set
	Add = rootUci.Add
}

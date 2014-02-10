package sftp

import (
	"path"
	"strings"
)

// Known file types
var KnownFileTypes = map[string]bool{
	".autotest":    true,
	".env":         true,
	".env-example": true,
	".gitignore":   true,
	".guardfile":   true,
	".irbrc":       true,
	".pryrc":       true,
	"builder":      true,
	"c":            true,
	"capfile":      true,
	"changelog":    true,
	"clj":          true,
	"coffee":       true,
	"configure":    true,
	"copying":      true,
	"cpp":          true,
	"credits":      true,
	"css":          true,
	"diff":         true,
	"ejs":          true,
	"erb":          true,
	"erl":          true,
	"gemfile":      true,
	"gemspec":      true,
	"go":           true,
	"groovy":       true,
	"guardfile":    true,
	"h":            true,
	"haml":         true,
	"hpp":          true,
	"html":         true,
	"install":      true,
	"irbrc":        true,
	"j2":           true,
	"java":         true,
	"jbuilder":     true,
	"js":           true,
	"json":         true,
	"jsp":          true,
	"less":         true,
	"license":      true,
	"lua":          true,
	"makefile":     true,
	"markdown":     true,
	"md":           true,
	"mysql":        true,
	"notice":       true,
	"php":          true,
	"pl":           true,
	"podfile":      true,
	"pryrc":        true,
	"py":           true,
	"rabl":         true,
	"rake":         true,
	"rakefile":     true,
	"rant":         true,
	"rantfile":     true,
	"rb":           true,
	"rbw":          true,
	"readme":       true,
	"rhtml":        true,
	"rjs":          true,
	"ru":           true,
	"rxml":         true,
	"sass":         true,
	"scala":        true,
	"scm":          true,
	"scss":         true,
	"sh":           true,
	"sql":          true,
	"thor":         true,
	"thorfile":     true,
	"txt":          true,
	"vagrantfile":  true,
	"xml":          true,
	"yaml":         true,
	"yml":          true,
}

func IsKnownTextFile(name string) (ok bool) {
	base := strings.ToLower(path.Base(name))
	if _, ok = KnownFileTypes[base]; ok {
		return
	}
	ext := strings.ToLower(path.Ext(name))
	if len(ext) > 0 {
		_, ok = KnownFileTypes[ext[1:len(ext)]]
		return
	}
	return
}

package vi

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode"

	. "github.com/diontr00/vi/internal/color"
)

// helper pattern map to support param matching
var helperPattern = map[string]string{
	"id":      `[\d]+`,
	"default": `[\w]+`,
}

var regexCache = map[string]*regexp.Regexp{}
var regexRW sync.RWMutex

// Use to register global helper pattern to be use in param matching
// example : ip ,`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`
// then you can use something like  v.Get("/location/:ip", ...)
func RegisterHelper(pattern, regex string) {
	helperPattern[pattern] = regex
}

type (
	matchKey string
	helper   string
	// Match param store the params matched.
	matchParams = map[matchKey]string
)

// Expect test result , for each param expect value in TestMatcher
type ExpectMatch map[string]string

// Represent the url to be test again pattern in TestMatcher
type TestUrl string

// Can be use to validate whether the regex pattern match correctly
func TestMatcher(expect bool, pattern string, tests ...map[TestUrl]ExpectMatch) []error {
	var errs []error
	for i := range tests {
		for url, result := range tests[i] {
			matched, params := match(string(url), pattern)
			var matchResult = false
			// loop match param key and value
			for k, v := range result {
				if pv, ok := params[matchKey(k)]; ok || v == pv {
					matchResult = true
					break
				}
			}

			if matched != expect || matchResult != expect {
				msg := Red("Got %t and %v when match %s and %s \n", matched, params, url, pattern)
				errs = append(errs, fmt.Errorf("%s", msg))
			}
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// Match route path again the url
// Should be use after perform static full path match
// Return whether match , and the map of param and its value.
func match(url, path string) (matched bool, results matchParams) {
	path = strings.Trim(path, " ")
	if path == "" {
		return false, nil
	}

	paths := strings.Split(path, "/")
	// empty path
	if len(paths) == 1 {
		return false, nil
	}
	var (
		// Names of matched param in pattern.
		matchName = []string{}
		// Regex pattern holder to match again url.
		tmp strings.Builder
	)

	for i, pth := range paths {
		if pth != "" {
			// Byte represent ation of path.
			pthB := []byte(pth)
			lastIdx := len(pthB) - 1

			firstCh := string(pthB[0])
			lastCh := string(pthB[lastIdx])

			// Check first path
			// 0 should be empty. "for /".
			if i == 1 {
				if firstCh == "*" {
					// Only wildcard is treated as regex when represent as standalone value in path.
					if len(pth) == 1 {
						matched = true
						// Return nil insteead of empty map , since map all.
						results = nil
						return
					}
				}
				// Other meta char should be treated as normal word by escape them.
				if isMeta(firstCh) {
					if len(pth) == 1 {
						tmp.WriteString("/")
						tmp.WriteString(escapeNonAlphaNum(pth))
						continue
					}
				}
			}

			if firstCh == "{" && lastCh == "}" {
				// named:regex
				ptrns := bytes.Split(pthB[1:lastIdx], []byte(":"))
				param := string(ptrns[0])

				matchName = append(matchName, param)
				var regex string
				l := len(ptrns)
				switch {
				case l == 2 || l > 2:
					regex = string(ptrns[1])
				default:
					if param == "" {
						return false, nil
					}
					regex = param
				}

				tmp.WriteString("/")
				tmp.WriteString("(")
				tmp.WriteString(string(regex))
				tmp.WriteString(")")
			} else if firstCh == ":" {
				pattern := string(pthB)
				// :named.
				patterns := strings.Split(pattern, ":")
				// named
				param := patterns[1]
				paramLastIdx := len(param) - 1
				paramLastChar := []byte(param)[paramLastIdx]
				// Whether the last char is alpha numeric , which signify regex modifier.
				if isMeta(string(paramLastChar)) {
					// [name , specialChar ].
					matchName = append(matchName, param[0:paramLastIdx])
					tmp.WriteString(patternGen(strings.TrimSuffix(param, string(paramLastChar)), true))
					// Add modifier to regex string.
					tmp.WriteString(string(paramLastChar))
					continue
				}

				matchName = append(matchName, patterns[1])
				tmp.WriteString(patternGen(patterns[1], false))
			} else {
				// Handle non define pattern case , escape all.
				tmp.WriteString("/" + escapeNonAlphaNum(pth))
			}
		}
	}

	return regexHelper(url, tmp.String(), matchName)
}

// Generate regrex pattern for s
// If group is true , the  pattern  is  prefix with / before capture , when we want to apply modifier to the pattern  like optional.
func patternGen(s string, group bool) string {
	var pattern strings.Builder
	pattern.WriteString("(")
	if group {
		pattern.WriteString("/")
	}

	if p, ok := helperPattern[s]; ok {
		pattern.WriteString(p)
	} else {
		pattern.WriteString(helperPattern["default"])
	}

	pattern.WriteString(")")
	if group {
		return pattern.String()
	}
	return "/" + pattern.String()
}

// Regex matching of url  again generated pattern.
func regexHelper(url, pattern string, matchedName []string) (matched bool, result map[matchKey]string) {
	result = make(map[matchKey]string, len(matchedName))

	regexRW.RLock()
	defer regexRW.RUnlock()
	var r *regexp.Regexp
	if rc, ok := regexCache[pattern]; !ok {
		r = regexp.MustCompile(pattern)
		go func(regex *regexp.Regexp) {
			defer regexRW.Unlock()
			regexRW.Lock()
			regexCache[pattern] = regex
		}(r)
	} else {
		r = rc
	}

	submatch := r.FindSubmatch([]byte(url))

	if submatch != nil {
		submatch = submatch[1:]
		for i, match := range submatch {
			if i > len(matchedName)-1 {
				break
			}
			if len(match) == 0 {
				continue
			}

			key := matchedName[i]
			if string(match[0]) == "/" {
				match = match[1:]
			}
			result[matchKey(key)] = string(match)
		}
		if len(result) == 0 {
			result = nil
		}
		matched = true
		return
	}
	result = nil
	matched = false
	return
}

// Check whether the string  contain meta character , that not curly bracket to signify regex pattern.
func isMeta(s string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9{}]+$`)
	match := re.FindSubmatch([]byte(s))

	return len(match) == 0
}

// Escape all non alpha numeric in the string.
func escapeNonAlphaNum(s string) string {
	var buffer bytes.Buffer
	for _, char := range s {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			buffer.WriteRune(char)
		} else {
			buffer.WriteString(fmt.Sprintf("\\%c", char))
		}
	}
	return buffer.String()
}

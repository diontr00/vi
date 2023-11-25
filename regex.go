package vi

import (
	"regexp"
	"strings"
)

// helper pattern map to support param matching
var helperPattern = map[string]string{
	"id":      `[\d]+`,
	"default": `[\w]+`,
}

// Use to register helper pattern to be use in param matching
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

// Match route path again the url
// Should be use after perform static full path match
// Return whether match , and the map of param and its value.
func match(url, path string) (matched bool, results matchParams) {
	path = strings.Trim(path, " ")
	paths := strings.Split(path, "/")
	// empty path
	if len(paths) == 1 {
		return false, nil
	}
	var (
		// Names of matched param in pattern.
		matchName = []string{}
		// Regex pattern holder to match again url.
		tmp = ""
	)

	for i, pth := range paths {
		if pth != "" {
			// Byte represent ation of path.
			pthB := []byte(pth)
			lastIdx := len(pthB) - 1

			// Check first path
			// 0 should be empty. "for /".
			if i == 1 {
				firstCh := string(pthB[0])
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
						tmp = tmp + "/" + escapeNonAlphaNum(pth)
						continue
					}
				}
			}

			if string(pthB[0]) == "{" && string(pthB[lastIdx]) == "}" {
				// {named:regex}
				t := string(pthB[1:lastIdx])
				// named:regex
				ptrns := strings.Split(t, ":")

				matchName = append(matchName, ptrns[0])
				regex := ptrns[1]

				tmp += "/" + "(" + string(regex) + ")"
			} else if string(pthB[0]) == ":" {
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
					tmp += patternGen(strings.TrimSuffix(param, string(paramLastChar)), true)
					// Add modifier to regex string.
					tmp += string(paramLastChar)
					continue
				}

				matchName = append(matchName, patterns[1])
				tmp += patternGen(patterns[1], false)
			} else {
				// Handle non define pattern case , escape all.
				tmp = tmp + "/" + escapeNonAlphaNum(pth)
			}
		}
	}

	return regexHelper(url, tmp, matchName)
}

// Generate regrex pattern for s
// If group is true , the  pattern  is  prefix with / before capture , when we want to apply modifier to the pattern  like optional.
func patternGen(s string, group bool) string {
	var pattern helper

	if p, ok := helperPattern[s]; ok {
		pattern = helper(p)
	} else {
		pattern = helper(helperPattern["default"])
	}

	if group {
		pattern = "/" + pattern
		return "(" + string(pattern) + ")"
	}
	return "/" + "(" + string(pattern) + ")"
}

// Regex matching of url  again generated pattern.
func regexHelper(url, pattern string, matchedName []string) (matched bool, result map[matchKey]string) {
	result = map[matchKey]string{}
	regex := regexp.MustCompile(pattern)
	submatch := regex.FindSubmatch([]byte(url))

	if submatch != nil {
		submatch = submatch[1:]
		for i, match := range submatch {
			if len(submatch[i]) == 0 {
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
	re := regexp.MustCompile("[^a-zA-Z0-9]+")
	escaped := re.ReplaceAllString(s, "\\$0")
	return escaped
}

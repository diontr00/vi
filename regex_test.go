package vi

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("Regex: Unit-test", func(url string, path string, result bool, vals map[matchKey]string) {
	matched, params := match(url, path)

	if result {
		Expect(matched).To(BeTrue(), "Expect %s will match %s", path, url)
	} else {
		Expect(matched).To(BeFalse(), "Expect %s will not match %s", path, url)
		return
	}
	if vals != nil {
		for k, v := range params {
			Expect(params[matchKey(k)]).To(Equal(v), "Expect param : %s to equal %s for path: %s and url: %s", path, url)
		}
		Expect(params).To(HaveLen(len(vals)), "Expect len of matched parms  to be %d , but get %d", len(vals), len(params))
	} else {
		Expect(params).To(BeNil())
	}

},
	Entry("Should match correctly simple named params", "/dion", "/:name", true, map[matchKey]string{"name": "dion"}),
	Entry("Should match optional modifiers with return values", "/user/dion/1234/vietnam", "/user/:name/1234/:nationality?", true, map[matchKey]string{"name": "dion", "nationality": "vietnam"}),
	Entry("Should match optional modifier with no return values", "/video/:title", "/video", true, nil),
	Entry("Should match star modifiers", "/user/dion/1234", "/user/:name/:id*", true, map[matchKey]string{"name": "dion", "id": "1234"}),
	Entry("Should match regex pattern", "/employee/153/accounting", "/employee/{uid:[0-9]+}/{department:[a-zA-Z]+}", true, map[matchKey]string{"uid": "153", "department": "accounting"}),
	Entry("Should match helper pattern", "/user/dion/1234", "/user/:name/:id", true, map[matchKey]string{"name": "dion", "id": "1234"}),
	Entry("Should match catchAll pattern", "/usr/dion/user/123/a", "/*", true, nil),
	Entry("Should escape meta when standalone suffix", "/something", "/?", false, nil),
	Entry("Should escape meta when standalone suffix true", "/?", "/?", true, nil),
	Entry("Should escape meta prefix", "/?something", "/something?", false, nil),
	Entry("Should escape meta prefix true", "/", "/:something?", true, nil),
	Entry("Should notMatch invalid path", "/anh", "/{name:{z%$!}}", false, nil),
	Entry("Should notMatch invalid path2", "/anh", "/:name!", false, nil),
	Entry("Should return not matched if receive invalid path", "/anh", "!", false, nil),
	Entry("Should not matched if empty", "/anh", "", false, nil),
	Entry("Should not panic when provide regex with no pattern", "/anh", "/{anh}", true, map[matchKey]string{"anh": "anh"}),
	Entry("Should not turn empty regex to empty string, since it can match something unexpected", "/anh", "/{}", false, nil),
)

var _ = DescribeTable("Check isMeta return true only if string contain meta character", func(pattern string, metaExist bool) {
	result := isMeta(pattern)
	Expect(metaExist).To(Equal(result), fmt.Sprintf("for pattern %s , expeted %v but got %v", pattern, metaExist, result))

},
	Entry("", "abc123", false),
	Entry("", "abc$123", true),
	Entry("", "+", true),
	Entry("", "*", true),
	Entry("", "^*?", true),
)

var _ = DescribeTable("Test escapeMeta to escape all the meta character in the string", func(input string, expected string) {
	result := escapeNonAlphaNum(input)
	Expect(result).To(Equal(expected), "for input %s, expected %s , but got %s")

},

	Entry("", "abc123", "abc123"),
	Entry("", "abc$123", "abc\\$123"),
	Entry("", "", ""),
	Entry("", "1@2#3$4", "1\\@2\\#3\\$4"),
)

var _ = Describe("Test pattern register", func() {
	It("Should matched correctly and return associate params", func() {
		RegisterHelper("ip", `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
		matched, param := match("/location/192.168.0.1", "/location/:ip")
		Expect(matched).To(BeTrue())
		Expect(param).To(HaveKeyWithValue(matchKey("ip"), "192.168.0.1"))
	})

})

var _ = Describe("TestMatcher", func() {
	It("Test expect match", func() {
		RegisterHelper("ip", `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
		RegisterHelper("phone", `([+]?[\s0-9]+)?(\d{3}|[(]?[0-9]+[)])?([-]?[\s]?[0-9])+`)
		errs := TestMatcher(true, "/location/:ip/:phone?", map[TestUrl]ExpectMatch{
			"/location/192.168.0.1/999-9999999":        {"ip": "192.168.0.1", "phone": "999-9999999"},
			"/location/192.168.0.2/+48(12)504-203-260": {"ip": "192.168.0.2", "phone": "+48(12)504-203-260"},
			"/location/192.168.0.3":                    {"ip": "192.168.0.3"},
		})

		Expect(errs).To(BeNil())
	})

	It("Test expect not match", func() {
		RegisterHelper("ip", `/caller/((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])`)
		errs := TestMatcher(false, "/location/:ip", map[TestUrl]ExpectMatch{
			"/location/192.168.0.256": {"ip": "192.168.0.1"},
		})

		for _, err := range errs {
			fmt.Println(err)
		}
		Expect(errs).To(BeNil())

		RegisterHelper("ip", `/caller/((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])`)
		errs = TestMatcher(true, "/location/:ip", map[TestUrl]ExpectMatch{
			"/location/192.168.0.256": {"ip": "192.168.0.1"},
		})

		Expect(errs).ToNot(BeNil())

	})

})

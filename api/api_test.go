package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

type configAPI struct {
	URL    string  `json:"url"`
	Secret string  `json:"secret"`
	SHA    SHA `json:"sha"`
}

type configClient struct {
	URL string `json:"url"`
	WS  string `json:"ws"`
}

type configBBB struct {
	API    configAPI	`json:"api"`
	Client configClient	`json:"client"`
}

type config struct {
	BBB configBBB `json:"bbb"`
}

// For reading config from a file or from environment variables
func readConfig(file string, t *testing.T) config {
	// Try to read from env
	conf := config {
		BBB: configBBB{
			API: configAPI{
				URL: os.Getenv("BBB_API_URL"),
				Secret: os.Getenv("BBB_API_SECRET"),
				SHA: SHA(os.Getenv("BBB_API_SECRET")),
			},
			Client: configClient{
				URL: os.Getenv("BBB_CLIENT_URL"),
				WS: os.Getenv("BBB_CLIENT_WS"),
			},
		},
	}

	if (conf.BBB.API.URL != "" && conf.BBB.API.Secret != "" && conf.BBB.API.SHA != "" && conf.BBB.Client.URL != "" && conf.BBB.Client.WS != ""){
		fmt.Println("Using env variables for config")
		return conf
	}

	// Open our jsonFile
	jsonFile, err := os.Open(file)
	// if we os.Open returns an error then handle it
	if (err != nil) {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	// read our opened jsonFile as a byte array.
	byteValue, err := io.ReadAll(jsonFile)
	if(err != nil) {
		panic(err)
	}
	// we unmarshal our byteArray which contains our jsonFile's content into conf
	json.Unmarshal([]byte(byteValue), &conf) 

	return conf
}

type testnewrequest struct {
	url        string
	secret     string
	shatype    SHA
	expected   ApiRequest
	shouldfail bool
}

func TestNewRequest(t *testing.T) {
	tests := []testnewrequest{
		{ //0
			url:     "https://example.com",
			secret:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			shatype: SHA256,
			expected: ApiRequest{
				url:     "https://example.com/api/",
				secret:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
				shatype: SHA256,
			},
			shouldfail: false,
		},
		{ //1
			url:     "https://example.com/",
			secret:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			shatype: SHA1,
			expected: ApiRequest{
				url:     "https://example.com/api/",
				secret:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
				shatype: SHA1,
			},
			shouldfail: false,
		},
		{ //2
			url:     "http://example.com/",
			secret:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			shatype: SHA256,
			expected: ApiRequest{
				url:     "http://example.com/api/",
				secret:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
				shatype: SHA256,
			},
			shouldfail: false,
		},
		{ //3
			url:        "example.com",
			secret:     "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			expected:   ApiRequest{},
			shouldfail: true,
		},
		{ //4
			url:     "https://example.com",
			secret:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			shatype: "bla",
			expected: ApiRequest{
				url:     "https://example.com/api/",
				secret:  "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
				shatype: SHA256,
			},
			shouldfail: false,
		},
	}

	for num, test := range tests {
		result, err := NewRequest(test.url, test.secret, test.shatype)
		fmt.Println(test.shatype)

		if err != nil && !test.shouldfail {
			t.Errorf("NewRequest(%s,%s) %d FAILED: Error %s", test.url, test.secret, num, err)
		}

		if !reflect.DeepEqual(result, &test.expected) {
			t.Errorf("NewRequest(%s,%s) %d FAILED: Object is not correct", test.url, test.secret, num)
		} else {
			t.Logf("NewRequest(%s,%s) %d PASSED", test.url, test.secret, num)
		}
	}
}

type testgeneratechecksum struct {
	action          action
	params          []params
	expected_sha1   string
	expected_sha256 string
	shouldfail      bool
}

// Test for generateChecksum
// https://mconf.github.io/api-mate/
func TestGenerateChecksum(t *testing.T) {
	tests := []testgeneratechecksum{
		{ //0
			action: CREATE,
			params: []params{
				{
					name:  ALLOW_START_STOP_RECORDING,
					value: "true",
				},
				{
					name:  ATTENDEE_PW,
					value: "ap",
				},
				{
					name:  AUTO_START_RECORDING,
					value: "false",
				},
				{
					name:  MEETING_ID,
					value: "random-4026116",
				},
				{
					name:  MODERATOR_PW,
					value: "mp",
				},
				{
					name:  NAME,
					value: "random-4026116",
				},
				{
					name:  RECORD,
					value: "false",
				},
				{
					name:  VOICE_BRIDGE,
					value: "70848",
				},
				{
					name:  WELCOME,
					value: "Hello you there",
				},
			},
			expected_sha1:   "2c2f2b2f6050bda0ff2c6dacd9d51e09951810ae",
			expected_sha256: "ae982d76751077e4e1eae8a667d5f74fe4f9c9a9df7d30ff2e56b3a025f1828d",
			shouldfail:      false,
		},
		{ //1
			action: CREATE,
			params: []params{
				{
					name:  ALLOW_START_STOP_RECORDING,
					value: "true",
				},
				{
					name:  ATTENDEE_PW,
					value: "ap",
				},
				{
					name:  AUTO_START_RECORDING,
					value: "false",
				},
				{
					name:  MEETING_ID,
					value: "random-3098916",
				},
				{
					name:  MODERATOR_PW,
					value: "mp",
				},
				{
					name:  NAME,
					value: "Hallöchen",
				},
				{
					name:  RECORD,
					value: "false",
				},
				{
					name:  VOICE_BRIDGE,
					value: "75469",
				},
				{
					name: WELCOME,
					value: `<br>Welcome to <b>%%CONFNAME%%</b>!
					This is a test & it shouldn't "work" ;-<)
					!"§$%&/()=?*'_:;>{[]}\/()=+#-.,<|`,
				},
			},
			expected_sha1:   "58ac486010b9e5b90ef43900a479a3acffeda337",
			expected_sha256: "fb1ae91324df1d4a523174d752e37a4880d06541cbab93345df6ff2056a4d377",
			shouldfail:      false,
		},
	}

	bbbapi_sha1, _ := NewRequest("https://example.com/bigbluebutton/api/", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", SHA1)
	bbbapi_sha256, _ := NewRequest("https://example.com/bigbluebutton/api/", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", SHA256)

	for num, test := range tests {
		params := bbbapi_sha1.buildParams(test.params...)

		//Sha1
		resultsha1 := bbbapi_sha1.generateChecksum(test.action, params)
		if resultsha1 != test.expected_sha1 {
			if test.shouldfail {
				t.Logf("generateChecksumSHA1(%s,...) %d PASSED", test.action, num)
			} else {
				t.Errorf("generateChecksumSHA1(%s,...) %d FAILED: Cheksum is wrong: %s", test.action, num, bbbapi_sha1.url+string(test.action)+"?"+params+"&checksum="+resultsha1)
			}
		} else {
			t.Logf("generateChecksumSHA1(%s,...) %d PASSED", test.action, num)
		}

		//Sha256
		resultsha256 := bbbapi_sha256.generateChecksum(test.action, params)
		if resultsha256 != test.expected_sha256 {
			if test.shouldfail {
				t.Logf("generateChecksumSHA256(%s,...) %d PASSED", test.action, num)
			} else {
				t.Errorf("generateChecksumSHA256(%s,...) %d FAILED: Cheksum is wrong: %s", test.action, num, bbbapi_sha256.url+string(test.action)+"?"+params+"&checksum="+resultsha256)
			}
		} else {
			t.Logf("generateChecksumSHA256(%s,...) %d PASSED", test.action, num)
		}
	}
}

type testmakeRequest struct {
	url        string
	secret     string
	action     action
	params     []params
	expected   any
	shouldfail bool
}

// Test for makeRequest
func TestMakeRequest(t *testing.T) {
	conf := readConfig("../_example/config.json", t)

	tests := []testmakeRequest{
		{ //0
			url:        "https://examfgfgfgfffple.com/bigbluebutton/api/",
			secret:     "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			action:     GET_MEETINGS,
			params:     []params{},
			expected:   "",
			shouldfail: true,
		},
		{ //1
			url:        conf.BBB.API.URL,
			secret:     "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			action:     GET_MEETINGS,
			params:     []params{},
			expected:   "",
			shouldfail: true,
		},
		{ //2
			url:        conf.BBB.API.URL,
			secret:     conf.BBB.API.Secret,
			action:     GET_MEETINGS,
			params:     []params{},
			expected:   "",
			shouldfail: false,
		},
		{ //3
			url:        strings.Replace(conf.BBB.API.URL, "bigbluebutton", "wrong", -1),
			secret:     conf.BBB.API.Secret,
			action:     GET_MEETINGS,
			params:     []params{},
			expected:   "",
			shouldfail: true,
		},
	}

	for num, test := range tests {
		bbbapi, err := NewRequest(test.url, test.secret, SHA1)
		if err != nil {
			t.Errorf("makeRequest(...,%s,...) %d FAILED: NewRequest: %s", test.action, num, err)
			continue
		}

		var response responsegetmeetings
		err = bbbapi.makeRequest(&response, test.action, test.params...)
		if err != nil {
			if !test.shouldfail {
				t.Errorf("makeRequest(...,%s,...) %d FAILED: err: %s", test.action, num, err)
				continue
			}
		}
		t.Logf("makeRequest(...,%s,...) %d PASSED", test.action, num)
	}
}

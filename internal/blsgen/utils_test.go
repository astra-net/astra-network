package blsgen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testPrompt = yesNoPrompt

func init() {
	// Move the test data to temp directory
	os.RemoveAll(baseTestDir)
	os.MkdirAll(baseTestDir, 0777)
}

var baseTestDir = filepath.Join(".testdata")

type testKey struct {
	publicKey   string
	privateKey  string
	passphrase  string
	keyFileData string
}

// testKeys are keys with valid passphrase and valid .pass file
var testKeys = []testKey{
	{
		// key with empty passphrase
		publicKey:   "0e969f8b302cf7648bc39652ca7a279a8562b72933a3f7cddac2252583280c7c3495c9ae854f00f6dd19c32fc5a17500",
		privateKey:  "78c88c331195591b396e3205830071901a7a79e14fd0ede7f06bfb4c5e9f3473",
		passphrase:  "",
		keyFileData: "1d97f32175d8875f251e15805fd08f0cda794d827cb02d2de7b10d10f36f951d68347bef1e7a3018bd865c6966219cd9c4d20b055c50f8e09a6a3a1666b7c112450f643cc3c175f541fae75da8a843d47993fe89ec85788fd6ea2e98",
	},
	{
		// key with non empty passphrase
		publicKey:   "c4e9adcd322fbdfa69575b1edae2b428425ab9c4096d0113eb5502e73df5bf737cfd0786db2fa4c5a0ff6eac59873190",
		privateKey:  "88741632f82bceb162d2dd92a6812de3c8bbf0e6dd31c2a449623ecd73d41c42",
		passphrase:  "astra",
		keyFileData: "d510deb36c6d563d320c6b7154340112c89a64c8cf37bd0f86cbe0821e5828287d144dae16126510e2928a830005edb8754b8ea86e0cdc59a64124f85eb09c908d4572dfbcb10ec7c9814d8257d1d6f325701d4120b95efdbecd7008",
	},
}

func writeFile(file string, data string) error {
	dir := filepath.Dir(file)
	os.MkdirAll(dir, 0700)
	return ioutil.WriteFile(file, []byte(data), 0600)
}

func TestPromptYesNo(t *testing.T) {
	tests := []struct {
		inputs     []string
		lenOutputs int
		expRes     bool
		expErr     error
	}{
		{
			inputs:     []string{"yes"},
			lenOutputs: 1,
			expRes:     true,
		},
		{
			inputs:     []string{"YES\n"},
			lenOutputs: 1,
			expRes:     true,
		},
		{
			inputs:     []string{"y"},
			lenOutputs: 1,
			expRes:     true,
		},
		{
			inputs:     []string{"Y"},
			lenOutputs: 1,
			expRes:     true,
		},
		{
			inputs:     []string{"\tY"},
			lenOutputs: 1,
			expRes:     true,
		},
		{
			inputs:     []string{"No"},
			lenOutputs: 1,
			expRes:     false,
		},
		{
			inputs:     []string{"\tn"},
			lenOutputs: 1,
			expRes:     false,
		},
		{
			inputs:     []string{"invalid input", "y"},
			lenOutputs: 2,
			expRes:     true,
		},
	}
	for i, test := range tests {
		tc := newTestConsole()
		setTestConsole(tc)
		for _, input := range test.inputs {
			tc.In <- input
		}

		got, err := promptYesNo(testPrompt)
		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Errorf("Test %v: %v", i, assErr)
		} else if assErr != nil {
			continue
		}

		// check results
		if got != test.expRes {
			t.Errorf("Test %v: result unexpected %v / %v", i, got, test.expRes)
		}
		gotOutputs := drainCh(tc.Out)
		if len(gotOutputs) != test.lenOutputs {
			t.Errorf("unexpected output size: %v / %v", len(gotOutputs), test.lenOutputs)
		}
		if clean, msg := tc.checkClean(); !clean {
			t.Errorf("Test %v: console unclean with message [%v]", i, msg)
		}
	}
}

func drainCh(c chan string) []string {
	var res []string
	for {
		select {
		case gotOut := <-c:
			res = append(res, gotOut)
		default:
			return res
		}
	}
}

func assertError(got, expect error) error {
	if (got == nil) != (expect == nil) {
		return fmt.Errorf("unexpected error [%v] / [%v]", got, expect)
	}
	if (got == nil) || (expect == nil) {
		return nil
	}
	if !strings.Contains(got.Error(), expect.Error()) {
		return fmt.Errorf("unexpected error [%v] / [%v]", got, expect)
	}
	return nil
}

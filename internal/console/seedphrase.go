package console

import (
	"fmt"
	"strings"

	"github.com/tyler-smith/go-bip39"
	"github.com/c-bata/go-prompt"
)

// AutoComplete Listener
type acl struct{}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}
	for _, w := range bip39.GetWordList() {
		t := prompt.Suggest{Text: w}
		s = append(s, t)
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

// AskMnemonicPhrase asks for the mnemonic phrase, which is autocompleted.
func AskMnemonicPhrase() string {
	fmt.Println("Please enter your mnemonic words and press enter after each word. End with an empty line.")

	return prompt.Input("> ", completer)
}

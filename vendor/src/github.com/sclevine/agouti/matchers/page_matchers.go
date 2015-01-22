package matchers

import (
	"github.com/onsi/gomega/types"
	"github.com/sclevine/agouti/matchers/internal/page"
)

// HaveTitle passes when the expected title is equivalent to the
// title of the provided page.
func HaveTitle(title string) types.GomegaMatcher {
	return &page.HaveTitleMatcher{ExpectedTitle: title}
}

// HavePopupText passes when the expected text is equivalent to the
// text contents of an open alert, confirm, or prompt popup.
func HavePopupText(text string) types.GomegaMatcher {
	return &page.HavePopupTextMatcher{ExpectedText: text}
}

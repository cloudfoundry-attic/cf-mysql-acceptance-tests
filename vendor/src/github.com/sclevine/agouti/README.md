Agouti
======

[![Build Status](https://api.travis-ci.org/sclevine/agouti.png?branch=master)](http://travis-ci.org/sclevine/agouti)

Integration testing for Go using Ginkgo and Gomega!

Install:
```bash
$ go get github.com/sclevine/agouti
```
To use with PhantomJS (OS X):
```bash
$ brew install phantomjs
```
To use with ChromeDriver (OS X):
```bash
$ brew install chromedriver
```
To use with Selenium Webdriver (OS X):
```bash
$ brew install selenium-server-standalone
```
If you encounter issues with Safari, [see this thread](https://code.google.com/p/selenium/issues/detail?can=2&q=7933&colspec=ID%20Stars%20Type%20Status%20Priority%20Milestone%20Owner%20Summary&id=7933).

To use the `matcher` package, which provides Gomega matchers:
```bash
$ go get github.com/onsi/gomega
```
To use the `dsl` package, which defines tests that can be run with Ginkgo:
```bash
$ go get github.com/onsi/ginkgo/ginkgo
```

If you use the `dsl` package, note that:
 * `Feature` is a Ginkgo `Describe`
 * `Scenario` is a Ginkgo `It`
 * `Background` is a Ginkgo `BeforeEach`
 * `Step` is a Ginkgo `By`

Feel free to import Ginkgo and use any of its container blocks instead! Agouti is 100% compatible with Ginkgo and Gomega.

The `core` package is a flexible, general-purpose WebDriver API for Go. Unlike the `dsl` package, `core` allows unlimited and simultaneous usage of PhantomJS, ChromeDriver, and Selenium. Using `core` and `matchers` with Ginkgo and Gomega (and without the `dsl` package) is the recommended way to use Agouti. The `dsl` package exists primarily to provide a familiar environment for Capybara users.

Godoc is available for
[`core`](https://godoc.org/github.com/sclevine/agouti/core),
[`dsl`](https://godoc.org/github.com/sclevine/agouti/dsl), and [`matchers`](https://godoc.org/github.com/sclevine/agouti/matchers).

If you plan to use Agouti `dsl` to write Ginkgo tests, add the start and stop commands for your choice of WebDriver in Ginkgo `BeforeSuite` and `AfterSuite` blocks.

See this example `project_suite_test.go` file:
```Go
package project_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/dsl"

	"testing"
)

func TestProject(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Project Suite")
}

var _ = BeforeSuite(func() {
	StartPhantomJS()
	// OR
	StartChrome()
	// OR
	StartSelenium()
});

var _ = AfterSuite(func() {
	StopWebdriver()
});
```

Example:

```Go
import (
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/core"
	. "github.com/sclevine/agouti/dsl"
	. "github.com/sclevine/agouti/matchers"
)

...

var _ = Feature("Agouti", func() {
	var page Page

	Background(func() {
		page = CreatePage()
		page.Size(640, 480)
		page.Navigate("http://example.com/")
	})

	AfterEach(func() {
		page.Destroy()
	})

	Scenario("finding the page title", func() {
		Expect(page).To(HaveTitle("Page Title"))
	})

	Scenario("finding page elements", func() {
		Step("finding a header in the page", func() {
			Expect(page.Find("header")).To(BeFound())
		})

		Step("finding text in the header", func() {
			Expect(page.Find("header")).To(HaveText("Title"))
		})

		Step("asserting that text is not in the header", func() {
			Expect(page.Find("header")).NotTo(HaveText("Not-Title"))
		})

		Step("referring to an element by selection index", func() {
			Expect(page.All("option").At(0)).To(HaveText("first option"))
			Expect(page.All("select").At(1).All("option").At(0)).To(HaveText("third option"))
		})

		Step("matching text in the header", func() {
			Expect(page.Find("header")).To(MatchText("T.+e"))
		})

		Step("scoping selections by chaining", func() {
			Expect(page.Find("header").Find("h1")).To(HaveText("Title"))
		})

		Step("locating elements by XPath", func() {
			Expect(page.Find("header").FindByXPath("//h1")).To(HaveText("Title"))
		})

		Step("comparing two selections for equality", func() {
			Expect(page.Find("#some_element")).To(EqualElement(page.FindByXPath("//div[@class='some-element']")))
		})
	})

	Scenario("selecting multiple elements", func() {
		Step("asserting on their state", func() {
			Expect(page.All("select").All("option")).To(BeVisible())
			Expect(page.All("h1,h2")).NotTo(BeVisible())
		})
	})

	Scenario("finding form elements by label", func() {
		Step("finding an element by label text", func() {
			Expect(page.FindByLabel("Some Label")).To(HaveAttribute("value", "some labeled value"))
		})

		Step("finding an element embedded in a label", func() {
			Expect(page.FindByLabel("Some Container Label")).To(HaveAttribute("value", "some embedded value"))
		})
	})

	Scenario("element visibility", func() {
		Expect(page.Find("header h1")).To(BeVisible())
		Expect(page.Find("header h2")).NotTo(BeVisible())
	})

	Scenario("asynchronous javascript and DOM assertions", func() {
		Step("waiting for matchers to be true", func() {
			Expect(page.Find("#some_element")).NotTo(HaveText("some text"))
			Eventually(page.Find("#some_element"), 4*time.Second).Should(HaveText("some text"))
			Consistently(page.Find("#some_element")).Should(HaveText("some text"))
		})

		Step("serializing the current page HTML", func() {
			Expect(page.HTML()).To(ContainSubstring(`<div id="some_element" class="some-element" style="color: blue;">some text</div>`))
		})

		Step("executing arbitrary javascript", func() {
			arguments := map[string]interface{}{"elementID": "some_element"}
			var result string
			Expect(page.RunScript("return document.getElementById(elementID).innerHTML;", arguments, &result)).To(Succeed())
			Expect(result).To(Equal("some text"))
		})
	})

	Scenario("filling fields and asserting on their values", func() {
		Step("entering values into fields", func() {
			Fill(page.Find("#some_input"), "some other value")
		})

		Step("retrieving attributes by name", func() {
			Expect(page.Find("#some_input")).To(HaveAttribute("value", "some other value"))
		})
	})

	Scenario("CSS styles", func() {
		Expect(page.Find("#some_element")).To(HaveCSS("color", "rgba(0, 0, 255, 1)"))
		Expect(page.Find("#some_element")).To(HaveCSS("color", "rgb(0, 0, 255)"))
		Expect(page.Find("#some_element")).To(HaveCSS("color", "blue"))
	})

	Scenario("form actions", func() {
		Step("double-clicking on an element", func() {
			selection := page.Find("#double_click")
			DoubleClick(selection)
			Expect(selection).To(HaveText("double-click success"))
		})

		Step("checking a checkbox", func() {
			checkbox := page.Find("#some_checkbox")
			Check(checkbox)
			Expect(checkbox).To(BeSelected())
		})

		Step("selecting an option by text", func() {
			selection := page.Find("#some_select")
			Select(selection, "second option")
			Expect(selection.Find("option:last-child")).To(BeSelected())
		})

		Step("submitting a form", func() {
			Submit(page.Find("#some_form"))
			Eventually(Submitted).Should(BeTrue())
		})
	})

	Scenario("links and navigation", func() {
		Step("allows clicking on a link", func() {
			Click(page.FindByLink("Click Me"))
			Expect(page.URL()).To(ContainSubstring("#new_page"))
		})

		Step("allows navigating through browser history", func() {
			Expect(page.Back()).To(Succeed())
			Expect(page.URL()).NotTo(ContainSubstring("#new_page"))
			Expect(page.Forward()).To(Succeed())
			Expect(page.URL()).To(ContainSubstring("#new_page"))
		})

		Step("allows refreshing the page", func() {
			checkbox := page.Find("#some_checkbox")
			Check(checkbox)
			Expect(page.Refresh()).To(Succeed())
			Expect(checkbox).NotTo(BeSelected())
		})
	})
})
```

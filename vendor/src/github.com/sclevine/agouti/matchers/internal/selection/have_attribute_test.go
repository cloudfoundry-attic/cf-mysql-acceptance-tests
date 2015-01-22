package selection_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers/internal/mocks"
	. "github.com/sclevine/agouti/matchers/internal/selection"
)

var _ = Describe("HaveAttributeMatcher", func() {
	var (
		matcher   *HaveAttributeMatcher
		selection *mocks.Selection
	)

	BeforeEach(func() {
		selection = &mocks.Selection{}
		selection.StringCall.ReturnString = "CSS: #selector"
		matcher = &HaveAttributeMatcher{ExpectedAttribute: "some-attribute", ExpectedValue: "some value"}
	})

	Describe("#Match", func() {
		Context("when the actual object is a selection", func() {
			It("should request the provided page attribute", func() {
				matcher.Match(selection)
				Expect(selection.AttributeCall.Attribute).To(Equal("some-attribute"))
			})

			Context("when the expected attribute value matches the actual attribute value", func() {
				It("should successfully return true", func() {
					selection.AttributeCall.ReturnValue = "some value"
					success, err := matcher.Match(selection)
					Expect(success).To(BeTrue())
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the expected attribute value does not match the actual attribute value", func() {
				It("should successfully return false", func() {
					selection.AttributeCall.ReturnValue = "some other value"
					success, err := matcher.Match(selection)
					Expect(success).To(BeFalse())
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when retrieving the attribute value fails", func() {
				It("should return an error", func() {
					selection.AttributeCall.Err = errors.New("some error")
					_, err := matcher.Match(selection)
					Expect(err).To(MatchError("some error"))
				})
			})
		})

		Context("when the actual object is not a selection", func() {
			It("should return an error", func() {
				_, err := matcher.Match("not a selection")
				Expect(err).To(MatchError("HaveAttribute matcher requires a Selection.  Got:\n    <string>: not a selection"))
			})
		})
	})

	Describe("#FailureMessage", func() {
		It("should return a failure message", func() {
			selection.AttributeCall.ReturnValue = "some other value"
			matcher.Match(selection)
			message := matcher.FailureMessage(selection)
			Expect(message).To(ContainSubstring("Expected selection 'CSS: #selector' to have attribute matching\n    [some-attribute=\"some value\"]"))
			Expect(message).To(ContainSubstring("but found\n    [some-attribute=\"some other value\"]"))
		})
	})

	Describe("#NegatedFailureMessage", func() {
		It("should return a negated failure message", func() {
			selection.AttributeCall.ReturnValue = "some value"
			matcher.Match(selection)
			message := matcher.NegatedFailureMessage(selection)
			Expect(message).To(ContainSubstring("Expected selection 'CSS: #selector' not to have attribute matching\n    [some-attribute=\"some value\"]"))
			Expect(message).To(ContainSubstring("but found\n    [some-attribute=\"some value\"]"))
		})
	})
})

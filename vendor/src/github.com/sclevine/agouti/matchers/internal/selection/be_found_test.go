package selection_test

import (
	"github.com/sclevine/agouti/matchers/internal/mocks"
	. "github.com/sclevine/agouti/matchers/internal/selection"

	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BeFoundMatcher", func() {
	var (
		matcher   *BeFoundMatcher
		selection *mocks.Selection
	)

	BeforeEach(func() {
		selection = &mocks.Selection{}
		selection.StringCall.ReturnString = "CSS: #selector"
		matcher = &BeFoundMatcher{}
	})

	Describe("#Match", func() {
		Context("when the actual object is a selection", func() {
			Context("when the element is found", func() {
				It("should successfully return true", func() {
					selection.CountCall.ReturnCount = 1
					success, err := matcher.Match(selection)
					Expect(success).To(BeTrue())
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the element is not found", func() {
				It("should successfully return false", func() {
					selection.CountCall.ReturnCount = 0
					success, err := matcher.Match(selection)
					Expect(success).To(BeFalse())
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when the actual object is not a selection", func() {
			It("should return an error", func() {
				_, err := matcher.Match("not a selection")
				Expect(err).To(MatchError("BeFound matcher requires a Selection.  Got:\n    <string>: not a selection"))
			})
		})

		Context("when there is an error retrieving the count", func() {
			Context("when the error is an 'element not found' error", func() {
				It("should successfully return false", func() {
					selection.CountCall.Err = errors.New("element not found")
					success, err := matcher.Match(selection)
					Expect(success).To(BeFalse())
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the error is an 'element index out of range' error", func() {
				It("should successfully return false", func() {
					selection.CountCall.Err = errors.New("element index out of range")
					success, err := matcher.Match(selection)
					Expect(success).To(BeFalse())
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the error is any other error", func() {
				It("should return an error", func() {
					selection.CountCall.Err = errors.New("some error")
					_, err := matcher.Match(selection)
					Expect(err).To(MatchError("some error"))
				})
			})
		})
	})

	Describe("#FailureMessage", func() {
		It("should return a failure message", func() {
			selection.CountCall.ReturnCount = 0
			matcher.Match(selection)
			message := matcher.FailureMessage(selection)
			Expect(message).To(Equal("Expected selection 'CSS: #selector' to be found"))
		})
	})

	Describe("#NegatedFailureMessage", func() {
		It("should return a negated failure message", func() {
			selection.CountCall.ReturnCount = 1
			matcher.Match(selection)
			message := matcher.NegatedFailureMessage(selection)
			Expect(message).To(Equal("Expected selection 'CSS: #selector' not to be found"))
		})
	})
})

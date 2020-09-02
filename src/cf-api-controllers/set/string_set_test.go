package set_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/set"
)

var _ = Describe("StringSet", func() {
	var (
		set     StringSet
		strings []string
	)

	Context("NewStringSet", func() {
		JustBeforeEach(func() {
			set = NewStringSet(strings)
		})

		Context("given a non-empty slice of strings as input", func() {
			BeforeEach(func() {
				strings = []string{"foo", "bar"}
			})

			It("contains the input strings", func() {
				Expect(set).To(HaveLen(2))
				Expect(set).To(HaveKey("foo"))
				Expect(set).To(HaveKey("bar"))
			})
		})
	})

	Context("Difference", func() {
		var (
			left  StringSet
			right StringSet
		)

		Context("given an input which is a superset", func() {
			BeforeEach(func() {
				left = NewStringSet([]string{"one", "two"})
				right = NewStringSet([]string{"one", "two", "three"})
			})

			It("returns a set containing the difference of the two sets", func() {
				diff := left.Difference(right)
				Expect(diff).To(BeEmpty())
			})
		})

		Context("given an input which is the subset", func() {
			BeforeEach(func() {
				left = NewStringSet([]string{"one", "two", "three"})
				right = NewStringSet([]string{"one", "two"})
			})

			It("returns a set containing the difference of the two sets", func() {
				diff := left.Difference(right)
				Expect(diff).To(Equal(NewStringSet([]string{"three"})))
			})
		})

		Context("given an input which is the same set", func() {
			BeforeEach(func() {
				left = NewStringSet([]string{"one", "two"})
				right = NewStringSet([]string{"one", "two"})
			})

			It("returns a set containing the difference of the two sets", func() {
				diff := left.Difference(right)
				Expect(diff).To(BeEmpty())
			})
		})
	})

	Context("ToSlice", func() {
		var set StringSet

		Context("given a non-empty StringSet", func() {
			BeforeEach(func() {
				set = NewStringSet([]string{"one", "two", "three"})
			})

			It("returns a slice of strings containing all the elements of the set", func() {
				Expect(set.ToSlice()).To(Equal([]string{"one", "two", "three"}))
			})
		})
	})
})

package indexmanager

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("utils", func() {
	Context("parseIndexName", func() {

		When("the index name includes an inner name", func() {
			It("should return the document kind, version, and inner name", func() {
				parts := parseIndexName("grafeas-v1beta1-rode-notes")

				Expect(parts.DocumentKind).To(Equal("notes"))
				Expect(parts.Version).To(Equal("v1beta1"))
				Expect(parts.Inner).To(Equal("rode"))
			})
		})

		When("the index name includes an inner name with hyphens", func() {
			It("should return the name correctly", func() {
				parts := parseIndexName("grafeas-v1beta1-longer-inner-name-notes")

				Expect(parts.DocumentKind).To(Equal("notes"))
				Expect(parts.Version).To(Equal("v1beta1"))
				Expect(parts.Inner).To(Equal("longer-inner-name"))
			})
		})

		When("the index name does not include an inner name", func() {
			It("should return the document kind, and version", func() {
				parts := parseIndexName("grafeas-v1beta1-projects")

				Expect(parts.DocumentKind).To(Equal("projects"))
				Expect(parts.Version).To(Equal("v1beta1"))
				Expect(parts.Inner).To(BeEmpty())
			})
		})
	})
})

package store_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/georgisomnoev/feature-flag-api/internal/auth/model"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/store"
)

var _ = Describe("Store", func() {
	When("created", func() {
		It("exists", func() {
			Expect(store.NewStore(nil)).NotTo(BeNil())
		})
	})
	var (
		s         *store.Store
		userID    uuid.UUID
		testUser  model.User
		errAction error
	)

	ItSucceeds := func() {
		It("succeeds", func() {
			Expect(errAction).ToNot(HaveOccurred())
		})
	}

	BeforeEach(func() {
		s = store.NewStore(pool)
		userID = uuid.New()
		testUser = model.User{
			ID:       userID,
			Username: "testuser",
			Password: "testpassword",
			Role:     model.RoleEditor,
		}
	})

	Describe("GetByUsername", func() {
		var (
			result *model.User
		)

		BeforeEach(func() {
			err := s.AddUser(ctx, testUser)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := s.DeleteUserByID(ctx, testUser.ID)
			Expect(err).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			result, errAction = s.GetByUsername(ctx, testUser.Username)
		})

		ItSucceeds()
		It("retrieves the correct user", func() {
			Expect(*result).To(MatchFields(IgnoreExtras, Fields{
				"ID":       Equal(testUser.ID),
				"Username": Equal(testUser.Username),
				"Password": Equal(testUser.Password),
				"Role":     Equal(testUser.Role),
			}))
		})

		Context("when the user does not exist", func() {
			BeforeEach(func() {
				testUser.Username = "nonexistentuser"
			})

			ItSucceeds()
			It("returns nil", func() {
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("UserExists", func() {
		var (
			result bool
		)

		BeforeEach(func() {
			err := s.AddUser(ctx, testUser)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := s.DeleteUserByID(ctx, testUser.ID)
			Expect(err).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			result, errAction = s.UserExists(ctx, userID)
		})

		ItSucceeds()
		It("returns true for existing users", func() {
			Expect(result).To(BeTrue())
		})

		Context("when the user does not exist", func() {
			BeforeEach(func() {
				userID = uuid.New()
			})

			ItSucceeds()
			It("returns false without errors", func() {
				Expect(result).To(BeFalse())
			})
		})
	})
})

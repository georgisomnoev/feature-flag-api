package store_test

import (
	"fmt"
	"time"

	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/model"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/store"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Feature Flags Store", func() {
	When("created", func() {
		It("exists", func() {
			Expect(store.NewStore(nil)).NotTo(BeNil())
		})
	})
	var (
		s         *store.Store
		flagID    uuid.UUID
		flag      model.FeatureFlag
		errAction error
	)

	BeforeEach(func() {
		s = store.NewStore(pool)

		flagID = uuid.New()
		flag = model.FeatureFlag{
			ID:          flagID,
			Key:         "test-flag",
			Description: "test-description",
			Enabled:     true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
	})

	ItSucceeds := func() {
		It("succeeds", func() {
			Expect(errAction).NotTo(HaveOccurred())
		})
	}

	Describe("ListFlags", func() {
		var flags []model.FeatureFlag

		BeforeEach(func() {
			err := s.AddTestFlag(ctx, flag)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := s.RemoveTestFlag(ctx, flag.ID)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			flags, errAction = s.ListFlags(ctx)
		})

		ItSucceeds()
		It("returns all the feature flags", func() {
			Expect(flags).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"ID":          Equal(flag.ID),
				"Key":         Equal(flag.Key),
				"Description": Equal(flag.Description),
				"Enabled":     Equal(flag.Enabled),
				"CreatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
				"UpdatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
			})))
		})
	})

	Describe("GetFlagByID", func() {
		var fetchedFlag model.FeatureFlag

		BeforeEach(func() {
			err := s.AddTestFlag(ctx, flag)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := s.RemoveTestFlag(ctx, flag.ID)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			fetchedFlag, errAction = s.GetFlagByID(ctx, flagID)
		})

		ItSucceeds()
		It("returns the matching feature flag", func() {
			Expect(fetchedFlag.ID).To(Equal(flag.ID))
			Expect(fetchedFlag.Key).To(Equal(flag.Key))
			Expect(fetchedFlag.Description).To(Equal(flag.Description))
			Expect(fetchedFlag.Enabled).To(Equal(flag.Enabled))
		})

		Context("when the feature flag does not exist", func() {
			BeforeEach(func() {
				flagID = uuid.New()
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(model.ErrNotFound))
			})
		})
	})

	Describe("CreateFlag", func() {
		JustBeforeEach(func() {
			errAction = s.CreateFlag(ctx, flag)
		})

		JustAfterEach(func() {
			err := s.RemoveTestFlag(ctx, flag.ID)
			Expect(err).NotTo(HaveOccurred())
		})

		ItSucceeds()
		It("inserts the feature flag into the database", func() {
			insertedFlag, err := s.FetchTestFlagByID(ctx, flag.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(insertedFlag).To((MatchFields(IgnoreExtras, Fields{
				"ID":          Equal(flag.ID),
				"Key":         Equal(flag.Key),
				"Description": Equal(flag.Description),
				"Enabled":     Equal(flag.Enabled),
				"CreatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
				"UpdatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
			})))
		})
	})

	Describe("UpdateFlag", func() {
		BeforeEach(func() {
			err := s.AddTestFlag(ctx, flag)
			Expect(err).NotTo(HaveOccurred())

			flag.Key = "updated-flag"
			flag.Description = "updated-description"
			flag.UpdatedAt = time.Now().UTC()
		})

		AfterEach(func() {
			err := s.RemoveTestFlag(ctx, flag.ID)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			errAction = s.UpdateFlag(ctx, flag)
		})

		ItSucceeds()
		It("updates the feature flag in the database", func() {
			updatedFlag, err := s.FetchTestFlagByID(ctx, flag.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedFlag).To((MatchFields(IgnoreExtras, Fields{
				"ID":          Equal(flag.ID),
				"Key":         Equal(flag.Key),
				"Description": Equal(flag.Description),
				"Enabled":     Equal(flag.Enabled),
				"CreatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
				"UpdatedAt":   BeTemporally("~", time.Now().UTC(), time.Second),
			})))
		})
	})

	Describe("DeleteFlag", func() {
		JustBeforeEach(func() {
			errAction = s.DeleteFlag(ctx, flag.ID)
		})

		Context("when the feature flag exist", func() {
			BeforeEach(func() {
				err := s.AddTestFlag(ctx, flag)
				Expect(err).NotTo(HaveOccurred())
			})

			ItSucceeds()
			It("deletes the feature flag from the database", func() {
				var exists bool
				query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE id = $1)", store.FeatureFlagsTable)
				row := pool.QueryRow(ctx, query, flagID)
				Expect(row.Scan(&exists)).To(BeNil())
				Expect(exists).To(BeFalse())
			})
		})

		Context("when the feature flag does not exist", func() {
			BeforeEach(func() {
				flag.ID = uuid.New()
			})

			It("returns not found error", func() {
				Expect(errAction).To(MatchError(model.ErrNotFound))
			})
		})
	})
})

package vault

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFiles_Merge(t *testing.T) {
	testCases := []struct {
		fs1, fs2, fs3 Files
	}{
		{
			fs1: Files{},
			fs2: Files{{}},
			fs3: Files{{}},
		},
		{
			fs1: Files{{}},
			fs2: Files{},
			fs3: Files{{}},
		},
		{
			fs1: Files{
				{
					ID:         "1",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 41, 0, time.UTC),
					IsDeleted:  false,
				},
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
			fs2: Files{
				{
					ID:         "1",
					LastUpdate: time.Date(2024, 3, 8, 15, 45, 41, 0, time.UTC),
					IsDeleted:  true,
				},
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
			fs3: Files{
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
		},
		{
			fs1: Files{
				{
					ID:         "1",
					LastUpdate: time.Date(2024, 3, 8, 15, 45, 41, 0, time.UTC),
					IsDeleted:  true,
				},
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
			fs2: Files{
				{
					ID:         "1",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 41, 0, time.UTC),
					IsDeleted:  false,
				},
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
			fs3: Files{
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
		},
		{
			fs1: Files{
				{
					ID:         "1",
					LastUpdate: time.Date(2024, 3, 8, 15, 45, 41, 0, time.UTC),
					IsDeleted:  true,
				},
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
			fs2: Files{
				{
					ID:         "1",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 41, 0, time.UTC),
					IsDeleted:  false,
				},
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
				{
					ID:         "3",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
			fs3: Files{
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
				{
					ID:         "3",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
		},
		{
			fs1: Files{
				{
					ID:         "1",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 41, 0, time.UTC),
					IsDeleted:  false,
				},
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
				{
					ID:         "3",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
			fs2: Files{
				{
					ID:         "1",
					LastUpdate: time.Date(2024, 3, 8, 15, 45, 41, 0, time.UTC),
					IsDeleted:  true,
				},
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
			fs3: Files{
				{
					ID:         "2",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
				{
					ID:         "3",
					LastUpdate: time.Date(2024, 3, 8, 15, 30, 44, 0, time.UTC),
					IsDeleted:  false,
				},
			},
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.fs3, tc.fs1.Merge(tc.fs2))
	}
}

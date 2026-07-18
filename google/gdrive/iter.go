package gdrive

import (
	"context"
	"fmt"
	"iter"

	"google.golang.org/api/drive/v3"
)

func SharedDrives(ctx context.Context, ds *drive.Service) iter.Seq2[*drive.Drive, error] {
	return func(yield func(*drive.Drive, error) bool) {
		var pageToken string
		for {
			call := ds.Drives.List().PageSize(100).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			result, err := call.Do()
			if err != nil && !yield(nil, fmt.Errorf("drives.list: %w", err)) {
				return
			}
			for _, f := range result.Drives {
				if !yield(f, nil) {
					return
				}
			}
			if pageToken = result.NextPageToken; pageToken == "" {
				return
			}
		}
	}
}

// func FetchAll[T, U any](call Doer[T, U], yield Yielder[T]) {
// 	var pageToken string
// 	for {
// 		if pageToken != "" {
// 			call = call.PageToken(pageToken)
// 		}
// 		result, err := call.Do()
// 		if err != nil && !yield(nil, fmt.Errorf("drives.list: %w", err)) {
// 			return
// 		}
// 		for _, f := range result.Drives {
// 			if !yield(f, nil) {
// 				return
// 			}
// 		}
// 		if pageToken = result.NextPageToken; pageToken == "" {
// 			return
// 		}
// 	}
// }
//
// type Yielder[T any] func(*T, error) bool
// type Doer[T, U any] interface {
// 	PageToken(pageToken string) Doer[T, U]
// 	Do(opts ...googleapi.CallOption) (*T, error)
// }
//
// func Do[T any](opts ...googleapi.CallOption) (*T, error) {
// 	return nil, nil
// }

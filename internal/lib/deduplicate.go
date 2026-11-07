package lib

import (
	"context"
	"fmt"
	db "kimishinu/internal/db/sqlc"
)

func Deduplicate(ctx context.Context, keywords []string, body any, queries *db.Queries) ([]string, error) {
	switch item := body.(type) {
	case VideoResp:
		ids := make([]string, 0, len(item.Items))
		seen := make(map[string]struct{}, len(item.Items))
		for _, it := range item.Items {
			if _, s := seen[it.ContentDetails.VideoId]; s {
				continue
			}
			if Match(it.Snippet.Title, keywords) || Match(it.Snippet.Description, keywords) {
				// We choose empty strcuts over boolean because there is 0 memory allocation
				// I mean, it works even with boolean though
				seen[it.ContentDetails.VideoId] = struct{}{}
				ids = append(ids, it.ContentDetails.VideoId)
			}
		}
		if len(ids) == 0 {
			return nil, nil
		}
		return queries.InsertYoutubeIds(ctx, ids)
	case []Tweet:
		ids := make([]string, 0, len(item))
		seen := make(map[string]struct{}, len(item))
		bypass := map[string]struct{}{
			"aooont":          {},
			"kimishinu":       {},
			"anime_kimishinu": {},
			"kimishinu_event": {},
		}
		for _, tw := range item {
			if _, ok := bypass[tw.Author.UserName]; ok {
				seen[tw.ID] = struct{}{}
				ids = append(ids, tw.ID)
				continue
			}
			if _, s := seen[tw.ID]; s {
				continue
			}
			if Match(tw.Text, keywords) {
				seen[tw.ID] = struct{}{}
				ids = append(ids, tw.ID)
			}
		}
		if len(ids) == 0 {
			return nil, nil
		}
		return queries.InsertTwitterIds(ctx, ids)
	case TwitterResp:
		return Deduplicate(ctx, keywords, item.Data.Tweets, queries)
	case []MDData:
		ids := make([]string, 0, len(item))
		seen := make(map[string]struct{}, len(item))
		for _, md := range item {
			if _, s := seen[md.ID]; s {
				continue
			}
			seen[md.ID] = struct{}{}
			ids = append(ids, md.ID)
		}
		if len(ids) == 0 {
			return nil, nil
		}
		return queries.InsertMangaDexIds(ctx, ids)
	case MangaDexResp:
		return Deduplicate(ctx, keywords, item.Data, queries)
	default:
		return nil, fmt.Errorf("unsupported deduplication payload type %T", body)
	}
}

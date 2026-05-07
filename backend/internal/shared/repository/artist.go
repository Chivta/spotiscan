package repository

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/domain"
)

const (
	ruArtistsRedisKey = "ru_artists"
)

func NewArtistRepo(db *sql.DB, redisClient *redis.Client) *ArtistRepo {
	return &ArtistRepo{
		db:    db,
		redis: redisClient,
	}
}

type ArtistRepo struct {
	db     *sql.DB
	redis  *redis.Client
	loadMu sync.Mutex
}

func (r *ArtistRepo) LoadRussianArtistsToRedis(ctx context.Context) error {
	log.Info().Msg("loading ru_artists set from DB")
	allNames, err := r.GetAllRussianArtistNames(ctx)
	if err != nil {
		return err
	}
	err = r.SetRussianArtistNames(ctx, allNames)
	if err != nil {
		return err
	}
	log.Info().Msgf("successfully loaded %d ru artists into redis", len(allNames))
	return nil
}

func (r *ArtistRepo) GetAllRussianArtistNames(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT name FROM ru_artists`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ruNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		ruNames = append(ruNames, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ruNames, nil
}

func (r *ArtistRepo) SetRussianArtistNames(ctx context.Context, names []string) error {
	// Clear existing set and add new names
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, ruArtistsRedisKey)
	if len(names) > 0 {
		pipe.SAdd(ctx, ruArtistsRedisKey, names)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Fetches info for artists whose names are in DB, so implicitly filters out non-Russian
// names must be lowercase
func (r *ArtistRepo) GetRussianWithInfo(ctx context.Context, names []string) ([]domain.Artist, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id,name,description_ua,description_en,source,source_url,country,confirmed FROM ru_artists WHERE name = ANY($1)
    `, pq.Array(names))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ruArtists []domain.Artist
	for rows.Next() {
		var artist domain.Artist
		var descUA, descEN, source, sourceURL, country sql.NullString
		if err := rows.Scan(&artist.ID, &artist.Name, &descUA, &descEN, &source, &sourceURL, &country, &artist.Confirmed); err != nil {
			return nil, err
		}
		artist.DescriptionUA = descUA.String
		artist.DescriptionEN = descEN.String
		artist.Source = source.String
		artist.SourceURL = sourceURL.String
		artist.Country = country.String
		ruArtists = append(ruArtists, artist)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ruArtists, nil
}

func (r *ArtistRepo) InsertArtists(ctx context.Context, artists []domain.Artist) error {
	if len(artists) == 0 {
		return nil
	}
	var (
		names          = make([]string, 0, len(artists))
		descriptionsUA = make([]string, 0, len(artists))
		descriptionsEN = make([]string, 0, len(artists))
		sources        = make([]string, 0, len(artists))
		sourceURLs     = make([]string, 0, len(artists))
		countries      = make([]string, 0, len(artists))
		confirmed      = make([]bool, 0, len(artists))
	)

	seen := make(map[string]struct{}, len(artists))
	for _, artist := range artists {
		name := strings.ToLower(artist.Name)
		if name == "" {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}

		source := artist.Source
		if source == "" {
			source = "manual"
		}

		country := artist.Country
		if country == "" {
			country = "RU"
		}

		names = append(names, name)
		descriptionsUA = append(descriptionsUA, artist.DescriptionUA)
		descriptionsEN = append(descriptionsEN, artist.DescriptionEN)
		sources = append(sources, source)
		sourceURLs = append(sourceURLs, artist.SourceURL)
		countries = append(countries, country)
		confirmed = append(confirmed, artist.Confirmed)
	}

	if len(names) == 0 {
		return nil
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO ru_artists (name, description_ua, description_en, source, source_url, country, confirmed)
		 SELECT *
		 FROM unnest(
			$1::text[],
			$2::text[],
			$3::text[],
			$4::artist_source[],
			$5::text[],
			$6::text[],
			$7::boolean[]
		 )
		 ON CONFLICT (name) DO UPDATE SET
			description_ua = EXCLUDED.description_ua,
			description_en = EXCLUDED.description_en,
			source = EXCLUDED.source,
			source_url = EXCLUDED.source_url,
			country = EXCLUDED.country,
			confirmed = EXCLUDED.confirmed`,
		pq.Array(names),
		pq.Array(descriptionsUA),
		pq.Array(descriptionsEN),
		pq.Array(sources),
		pq.Array(sourceURLs),
		pq.Array(countries),
		pq.Array(confirmed),
	)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	return nil
}

// GetRuTags and GetRuRegionIds is only needed for scraper
// they are added to artist repo because its repo scraper already uses
func (r *ArtistRepo) GetRuTags(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT name FROM lastfm_ru_tags`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *ArtistRepo) GetRuRegionIds(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT mbid FROM musicbrainz_ru_regions`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *ArtistRepo) ArtistExists(ctx context.Context, name string) (bool, error) {
	name = strings.ToLower(name)
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM ru_artists WHERE name = $1)`, name).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

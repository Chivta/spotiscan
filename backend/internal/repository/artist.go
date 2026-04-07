package repository

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiscan/internal/appErrors"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/lib/pq"

	"github.com/redis/go-redis/v9"
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

func (r *ArtistRepo) FilterRussian(ctx context.Context, names []string) ([]string, error) {
	if r.redis != nil {
		ruNames, err := r.FilterRussianArtistNames(ctx, names)
		if err != nil {
			log.Warn().Msgf("redis error: %T: %v", err, err)
		} else {
			return ruNames, nil
		}
	}
	// fallback to db
	ruNames, err := r.filterRussianWithDB(ctx, names)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, appErrors.ErrDatabaseFailure
	}
	return ruNames, nil
}

func (r *ArtistRepo) FilterRussianArtistNames(ctx context.Context, names []string) ([]string, error) {
	if len(names) == 0 {
		return []string{}, nil
	}

	exists, err := r.redis.Exists(ctx, ruArtistsRedisKey).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		r.loadMu.Lock()
		defer r.loadMu.Unlock()
		// Re-check after acquiring lock in case another goroutine already loaded
		exists, err = r.redis.Exists(ctx, ruArtistsRedisKey).Result()
		if err != nil {
			return nil, err
		}

		if exists == 0 {
			log.Warn().Msg("ru_artists set not found in redis, loading from DB")
			err = r.LoadRussianArtistsToRedis(ctx)
			if err != nil {
				log.Error().Msgf("Failed to load ru_artists to redis: %v", err)
				return nil, err
			}
		}
	}

	// Use a pipeline to batch SISMEMBER commands
	pipe := r.redis.Pipeline()
	cmds := make([]*redis.BoolCmd, len(names))
	for i, name := range names {
		cmds[i] = pipe.SIsMember(ctx, ruArtistsRedisKey, name)
	}

	// Execute the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	// Collect results
	var ruNames []string
	for i, cmd := range cmds {
		isMember, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		if isMember {
			ruNames = append(ruNames, names[i])
		}
	}
	return ruNames, nil
}

func (r *ArtistRepo) filterRussianWithDB(ctx context.Context, names []string) ([]string, error) {
	ruNames, err := r.GetRussianArtistNames(ctx, names)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, appErrors.ErrDatabaseFailure
	}
	return ruNames, nil
}

func (r *ArtistRepo) GetRussianArtistNames(ctx context.Context, names []string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT name FROM ru_artists WHERE name = ANY($1)
    `, pq.Array(names))
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

func (r *ArtistRepo) GetArtistsInfo(ctx context.Context, names []string) ([]models.Artist, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT name,description_ua,description_en,source,source_url,country,confirmed FROM ru_artists WHERE name = ANY($1)
    `, pq.Array(names))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ruNames []models.Artist
	for rows.Next() {
		var artist models.Artist
		var descUA, descEN, source, sourceURL, country sql.NullString
		if err := rows.Scan(&artist.Name, &descUA, &descEN, &source, &sourceURL, &country, &artist.Confirmed); err != nil {
			return nil, err
		}
		artist.DescriptionUA = descUA.String
		artist.DescriptionEN = descEN.String
		artist.Source = source.String
		artist.SourceURL = sourceURL.String
		artist.Country = country.String
		ruNames = append(ruNames, artist)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ruNames, nil
}

func (r *ArtistRepo) InsertArtists(ctx context.Context, artists []models.Artist) error {
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
		return appErrors.ErrDatabaseFailure
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

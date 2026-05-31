package repository

import (
	"context"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/domain"
)

func NewArtistRepo(db *pgxpool.Pool) *ArtistRepo {
	return &ArtistRepo{
		db: db,
	}
}

type ArtistRepo struct {
	db     *pgxpool.Pool
	loadMu sync.Mutex
}

func (r *ArtistRepo) GetAllRussianArtistNames(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT name FROM ru_artists`)
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

	return ruNames, rows.Err()
}

// names must be lowercase
func (r *ArtistRepo) GetRussianWithInfo(ctx context.Context, names []string) ([]domain.Artist, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id,name,description_ua,description_en,source,source_url,country,confirmed FROM ru_artists WHERE name = ANY($1)
    `, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ruArtists []domain.Artist
	for rows.Next() {
		var artist domain.Artist
		var descUA, descEN, source, sourceURL, country *string
		if err := rows.Scan(&artist.ID, &artist.Name, &descUA, &descEN, &source, &sourceURL, &country, &artist.Confirmed); err != nil {
			return nil, err
		}
		if descUA != nil {
			artist.DescriptionUA = *descUA
		}
		if descEN != nil {
			artist.DescriptionEN = *descEN
		}
		if source != nil {
			artist.Source = *source
		}
		if sourceURL != nil {
			artist.SourceURL = *sourceURL
		}
		if country != nil {
			artist.Country = *country
		}
		ruArtists = append(ruArtists, artist)
	}

	return ruArtists, rows.Err()
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

	_, err := r.db.Exec(
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
		names,
		descriptionsUA,
		descriptionsEN,
		sources,
		sourceURLs,
		countries,
		confirmed,
	)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	return nil
}

func (r *ArtistRepo) GetRuTags(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT name FROM lastfm_ru_tags`)
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

	return tags, rows.Err()
}

func (r *ArtistRepo) GetRuRegionIds(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT mbid FROM musicbrainz_ru_regions`)
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

	return ids, rows.Err()
}

func (r *ArtistRepo) GetArtistCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ru_artists`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ArtistRepo) ArtistExists(ctx context.Context, name string) (bool, error) {
	name = strings.ToLower(name)
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ru_artists WHERE name = $1)`, name).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS lastfm_ru_tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL UNIQUE
);

INSERT INTO lastfm_ru_tags (name) VALUES 
('russian');

CREATE TABLE IF NOT EXISTS musicbrainz_ru_regions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL UNIQUE,
    mbid VARCHAR(36) NOT NULL UNIQUE
);

INSERT INTO musicbrainz_ru_regions (name, mbid) VALUES 
    ('Altayskiy kray', '7228e432-b2eb-42c3-8b12-063c63038c54'),
    ('Amurskaya oblast', 'd4568475-5a6a-43e9-ab49-4bdbe2b746ce'),
    ('Arkhangel''skaya oblast', '42600c00-4b16-41b0-9c6e-4eda7bf00680'),
    ('Chechenskaya Respublika', 'b6514d25-505b-413a-a2d0-62b9464496f4'),
    ('Chelyabinskaya oblast', '26b7ffd5-5df6-4c21-be93-7143df1b45ea'),
    ('Chukotskiy avtonomnyy okrug', '978b67a8-46e3-4718-984f-5d997eed9913'),
    ('Chuvashskaya Respublika', '1a64314a-ea0d-488d-9be2-ee619ad12a36'),
    ('Irkutskaya oblast', '405a06f8-4745-4507-9efb-762dfa3f2676'),
    ('Kabardino-Balkarskaya Respublika', '7d47d86a-6f5f-4c21-a1da-3140c8958715'),
    ('Kaliningradskaya oblast', 'e030d5d7-2d03-456f-91b8-a3fadfd6050d'),
    ('Kamchatskiy kray', '189f41da-4ec2-4b0d-a870-97d1582b880c'),
    ('Karachayevo-Cherkesskaya Respublika', '13be50b7-7b24-4ea3-b5aa-93a301c600fb'),
    ('Khabarovskiy kray', 'fc3cff5b-c47f-4ea2-aa93-2edcb8f95ef1'),
    ('Khanty-Mansiyskiy avtonomnyy okrug-Yugra', '485bcc38-f11a-4758-99c3-245eac07dd64'),
    ('Kirovskaya oblast', '4ee64aff-5e83-4747-92e8-7d0ab2a1310a'),
    ('Krasnodarskiy kray', 'b1cfa078-6827-4af2-b197-5d3522367a3d'),
    ('Krasnoyarskiy kray', '28e0abb1-6e1b-4291-ba60-bce4633f3f36'),
    ('Kurganskaya oblast', 'c770c24b-a616-44ec-a0a6-037cf768f37f'),
    ('Kurskaya oblast', 'f2725db1-295d-4d20-8f35-406798dbbaed'),
    ('Leningradskaya oblast', 'b64ada09-41aa-4b45-8d53-07c3dc47e6f1'),
    ('Moscow', 'f310740c-ad62-48c0-839b-e86581b9f464'),
    ('Moskovskaya oblast', 'd59ab45e-edc4-4ddf-a3df-5733db641f3e'),
    ('Murmanskaya oblast', '8607538f-3cc9-4680-8c88-e9ea8e7fae83'),
    ('Novosibirskaya oblast', 'fa06c373-a91e-427a-9718-e6b45503ff24'),
    ('Orenburgskaya oblast', '82eab6be-2d28-471a-a906-9c3c1a9f85f6'),
    ('Penzenskaya oblast', '88f4107c-905d-4887-be20-44eaf7cfe058'),
    ('Permskiy kray', '9d563979-3915-4716-9027-0d8e12983039'),
    ('Primorskiy kray', 'd73a1012-5ca0-443b-bc77-9fac9010b4eb'),
    ('Respublika Adygeya', 'a39536e1-8b06-4d20-a3f0-802c0be3a921'),
    ('Rostovskaya oblast', '4f9df1a3-8b21-40cd-bf2c-7dd230938286'),
    ('Sakhalinskaya oblast', '85b1246b-8936-4f93-b91e-8f17ba9905e5'),
    ('Sankt-Peterburg', '808e1ef8-5390-4300-a615-c4df977cc349'),
    ('Saratovskaya oblast', 'a72b86b4-732b-411e-be92-11de8821e70f'),
    ('Smolenskaya oblast', '391c3811-f741-4926-b1b7-eb09c46cf961'),
    ('Stavropol''skiy kray', 'eae1f983-f8dd-4e81-9b1f-9af8b48fc51c'),
    ('Sverdlovskaya oblast', '7d3118da-038c-46dd-bae9-16689b926862'),
    ('Tomskaya oblast', 'd87db114-8c07-41cf-af7c-40d4e6fa13ce'),
    ('Tverskaya oblast', '93b43808-412b-463c-bddd-8ee227766197'),
    ('Udmurtskaya Respublika', '85cfb970-3954-432e-a9ba-658e097c9772'),
    ('Vladimirskaya oblast', 'd9850091-5957-41bd-ab1f-e04a84f0c5fb'),
    ('Volgogradskaya oblast', '123ec471-830e-4dee-bcc6-dae79e5acf7d'),
    ('Yamalo-Nenetskiy avtonomnyy okrug', '4e2042d5-1c82-4015-aa7d-30cc2f9e482a'),
    ('Yaroslavskaya oblast', '472bcb98-ee64-4381-aa23-6bd8e173c752'),
    ('Yevreyskaya avtonomnaya oblast', '622d7402-109c-426a-9628-91f7565a0b88'),
    ('Zabaykal''skiy kray', 'c4b1c8e8-1876-47ce-a2a0-5c34db33568f'),
    ('Respublika Altay', 'ce3b5bfd-8507-4553-9f20-22ca625462e5'),
    ('Respublika Bashkortostan', '8570458e-202d-44b9-8d9b-f1d3f4926c6b'),
    ('Respublika Buryatiya', '51fdd4e4-6c11-46d4-8bec-0c7385d0f83d'),
    ('Respublika Dagestan', '37572420-4b2c-47e5-bf2b-536c9a50a362'),
    ('Respublika Ingushetiya', 'dbd26a8a-03de-4a71-a81e-11d700e08d4f'),
    ('Respublika Kalmykiya', '7fa4dd42-21f3-4c3d-9862-6845e9153e06'),
    ('Respublika Kareliya', 'a8c4b5ca-8e7a-41f3-893f-df91c5d51732'),
    ('Respublika Khakasiya', 'fc8d4e40-57bb-42fe-a46a-92c935520298'),
    ('Respublika Komi', '7fc7ef7e-ee42-40d3-bdd7-e993171cba13'),
    ('Respublika Mariy El', '75db1d6c-b090-4114-895e-338b6fb1228c'),
    ('Respublika Mordoviya', 'fbb13feb-0fcf-4f99-83fa-7a5239e68b82'),
    ('Respublika Sakha (Yakutiya)', '9795004e-e957-4fc2-8938-aef1b573b297'),
    ('Respublika Severnaya Osetiya-Alaniya', 'e4269131-c8f7-41ff-8600-1f49fe3b319c'),
    ('Respublika Tatarstan', '9593b81f-1249-48a9-9155-6456488e3cde');


CREATE TYPE artist_source AS ENUM ('lastfm', 'musicbrainz', 'phonkers_db', 'manual');

ALTER TABLE ru_artists 
    ADD COLUMN description_ua TEXT,
    ADD COLUMN description_en TEXT,
    ADD COLUMN source artist_source DEFAULT 'manual',
    ADD COLUMN source_url TEXT,
    ADD COLUMN country CHAR(2) DEFAULT 'RU',
    ADD COLUMN confirmed BOOLEAN DEFAULT FALSE,
    ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ru_artists_updated_at
    BEFORE UPDATE ON ru_artists
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS ru_artists_updated_at ON ru_artists;
DROP FUNCTION IF EXISTS set_updated_at();
ALTER TABLE ru_artists
    DROP COLUMN IF EXISTS description_ua,
    DROP COLUMN IF EXISTS description_en,
    DROP COLUMN IF EXISTS source,
    DROP COLUMN IF EXISTS source_url,
    DROP COLUMN IF EXISTS country,
    DROP COLUMN IF EXISTS confirmed,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;
DROP TYPE IF EXISTS artist_source;

-- +goose StatementEnd
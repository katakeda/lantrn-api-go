-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE IF NOT EXISTS facility_media_id_seq;
CREATE TABLE "facility_media" (
    "id" int4 NOT NULL DEFAULT nextval('facility_media_id_seq'::regclass),
    "title" varchar,
    "url" varchar,
    "is_primary" bool NOT NULL,
    "facility_id" varchar NOT NULL,
    PRIMARY KEY ("id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "facility_media";
-- +goose StatementEnd

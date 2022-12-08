-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE IF NOT EXISTS facility_id_seq;
CREATE TABLE "facility" (
    "id" int4 NOT NULL DEFAULT nextval('facility_id_seq'::regclass),
    "name" varchar NOT NULL DEFAULT ''::character varying,
    "description" text,
    "latitude" float4,
    "longitude" float4,
    "facility_id" varchar NOT NULL,
    "geom" geometry,
    PRIMARY KEY ("id")
);
CREATE INDEX "facility_geom_idx" ON "facility" USING BTREE ("geom");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX "facility_geom_idx";
DROP TABLE "facility";
-- +goose StatementEnd

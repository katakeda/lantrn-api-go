-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE IF NOT EXISTS subscription_id_seq;
CREATE TABLE "subscription" (
    "id" int4 NOT NULL DEFAULT nextval('subscription_id_seq'::regclass),
    "email" varchar NOT NULL,
    "target_date" varchar NOT NULL,
    "facility_id" int4 NOT NULL,
    CONSTRAINT "subscription_facility_id_fkey" FOREIGN KEY ("facility_id") REFERENCES "public"."facility"("id"),
    PRIMARY KEY ("id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "subscription";
-- +goose StatementEnd

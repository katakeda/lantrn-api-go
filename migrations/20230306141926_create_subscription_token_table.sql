-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE IF NOT EXISTS subscription_token_id_seq;
CREATE TABLE "subscription_token" (
    "id" int4 NOT NULL DEFAULT nextval('subscription_token_id_seq'::regclass),
    "subscription_id" int4 NOT NULL,
    "token" varchar NOT NULL,
    CONSTRAINT "subscription_token_subscription_id_fkey" FOREIGN KEY ("subscription_id") REFERENCES "public"."subscription"("id"),
    PRIMARY KEY ("id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "subscription_token";
-- +goose StatementEnd

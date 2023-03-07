-- +goose Up
-- +goose StatementBegin
ALTER TABLE "subscription" ADD COLUMN status VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "subscription" DROP COLUMN status;
-- +goose StatementEnd

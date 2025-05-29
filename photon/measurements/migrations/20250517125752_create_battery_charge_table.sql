-- +goose Up
CREATE TABLE measurements (
  timestamp TEXT NOT NULL,
  microvolts INTEGER NOT NULL
);

-- +goose Down
DROP TABLE measurements;

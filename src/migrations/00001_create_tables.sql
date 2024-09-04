-- +goose Up

CREATE TABLE "whitelist" (
  "network" varchar UNIQUE NOT NULL,
  "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "blacklist" (
  "network" varchar UNIQUE NOT NULL,
  "created_at" timestamptz DEFAULT (now())
);


-- +goose Down

DROP TABLE "whitelist";
DROP TABLE "blacklist";
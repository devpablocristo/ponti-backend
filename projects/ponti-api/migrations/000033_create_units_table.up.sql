CREATE TABLE IF NOT EXISTS units (
  id          BIGSERIAL PRIMARY KEY,
  name        VARCHAR(50) UNIQUE NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at  TIMESTAMPTZ,
  created_by  BIGINT,
  updated_by  BIGINT,
  deleted_by  BIGINT
);
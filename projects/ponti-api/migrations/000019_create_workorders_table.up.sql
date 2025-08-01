CREATE SEQUENCE IF NOT EXISTS workorder_number_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

CREATE TABLE workorders (
  number VARCHAR(20) PRIMARY KEY DEFAULT nextval('workorder_number_seq'),
  project_id    BIGINT NOT NULL REFERENCES projects(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  field_id      BIGINT NOT NULL REFERENCES fields(id)   ON UPDATE CASCADE ON DELETE RESTRICT,
  lot_id        BIGINT NOT NULL REFERENCES lots(id)     ON UPDATE CASCADE ON DELETE RESTRICT,
  crop_id       BIGINT NOT NULL REFERENCES crops(id)    ON UPDATE CASCADE ON DELETE RESTRICT,
  labor_id      BIGINT NOT NULL REFERENCES labors(id)   ON UPDATE CASCADE ON DELETE RESTRICT,
  contractor    VARCHAR(100),
  observations  TEXT,
  date          DATE    NOT NULL,
  investor_id   BIGINT  NOT NULL,
  effective_area NUMERIC(18,6) NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at    TIMESTAMPTZ,
  created_by    BIGINT,
  updated_by    BIGINT,
  deleted_by    BIGINT
);

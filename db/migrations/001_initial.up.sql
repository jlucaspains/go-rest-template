CREATE TABLE person (
  id                SERIAL PRIMARY KEY,
  name              varchar(255)    NOT NULL,
  email             varchar(255)    NOT NULL,
  created_at        timestamp       NOT NULL,
  updated_at        timestamp       NOT NULL,
  update_user       varchar(255)    NOT NULL,
  UNIQUE(email)
);
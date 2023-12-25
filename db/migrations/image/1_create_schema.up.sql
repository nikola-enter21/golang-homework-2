CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS images
(
    name       UUID         NOT NULL DEFAULT UUID_GENERATE_V4(),
    filename   VARCHAR(255) NOT NULL,
    alt_text   VARCHAR(255) NOT NULL,
    title      VARCHAR(255) NOT NULL,
    width      VARCHAR(255) NOT NULL,
    height     VARCHAR(255) NOT NULL,
    format     VARCHAR(50)  NOT NULL,
    source_url VARCHAR(255) NOT NULL
);

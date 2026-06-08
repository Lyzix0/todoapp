create schema todoapp;

create table todoapp.users (
    id           serial                primary key,
    version      bigint       not null default 1,
    full_name    varchar(100) not null check(char_length(full_name) between 3 and 100),
    phone_number varchar(15)           check(
        phone_number ~ '^\+[0-9]+$'
        and 
        char_length(phone_number) between 10 and 15
    )
);

CREATE TABLE todoapp.tasks (
    id           SERIAL                PRIMARY KEY,
    version      BIGINT       NOT NULL DEFAULT 1,
    title        VARCHAR(100) NOT NULL CHECK(char_length(title) BETWEEN 1 AND 1000),
    description  VARCHAR(1000)         CHECK(char_length(description) BETWEEN 1 AND 1000),
    completed    BOOLEAN      NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL,
    completed_at TIMESTAMPTZ,

    CHECK(
        (completed=FALSE AND completed_at IS NULL)
        OR 
        (completed=TRUE AND completed_at IS NOT NULL AND completed_at >= created_at)
    ),

    author_user_id INTEGER    NOT NULL REFERENCES todoapp.users(id)
);
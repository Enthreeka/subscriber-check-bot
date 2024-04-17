set timezone = 'Europe/Moscow';

DO $$
BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status') THEN
CREATE TYPE role AS ENUM ('main', 'secondary');
END IF;
END $$;

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role_user') THEN
            CREATE TYPE role_user AS ENUM ('user', 'admin','superAdmin');
        END IF;
    END $$;


create table if not exists "user"(
    id           bigint unique,
    tg_username  text not null ,
    created_at   timestamp  default now(),
    user_role         role_user default 'user' not null,
    primary key (id)
);

create table if not exists channel(
    id int generated always as identity,
    channel_telegram_id bigint not null,
    name varchar(200) not null,
    url varchar(250) not null,
    channel_status        role default 'secondary' not null,
    primary key(id)
);



create table if not exists message(
    id int generated always as identity,
    message text null,
    file_id varchar(150) null,
    button_url varchar(150) null,
    button_text varchar(150) null,
    primary key (id)
);




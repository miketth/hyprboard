CREATE TABLE last_layouts (
    app text not null,
    device text not null,
    code text not null,
    variant text not null,
    primary key (app, device)
);

CREATE TABLE schema_migrations (version uint64,dirty bool);

CREATE UNIQUE INDEX version_unique ON schema_migrations (version);


create table sqlite_master (
    type     text,
    name     text,
    tbl_name text,
    rootpage int,
    sql      text
);

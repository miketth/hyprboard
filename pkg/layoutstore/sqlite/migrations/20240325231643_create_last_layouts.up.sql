create table if not exists last_layouts (
    app text not null,
    device text not null,
    code text not null,
    variant text not null,
    primary key (app, device)
);


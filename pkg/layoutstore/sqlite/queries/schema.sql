-- name: DumpTables :many
select sql
from sqlite_master
where type = 'table'
and sql not null
order by name;

-- name: DumpRest :many
select sql
from sqlite_master
where type is not 'table'
and sql not null
order by name;

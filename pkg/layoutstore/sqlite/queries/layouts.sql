-- name: GetLayoutsForApp :many
select *
from last_layouts
where app = ?;

-- name: SetLayout :exec
insert into last_layouts (app, device, code, variant)
values (?1, ?2, ?3, ?4)
on conflict do update
set code = ?3, variant = ?4;

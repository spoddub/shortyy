-- name: CreateLinkVisit :execrows
INSERT INTO link_visits (link_id, ip, user_agent, referer, status)
VALUES ($1, $2, $3, $4, $5);

-- name: CountLinkVisits :one
SELECT count(*)::bigint AS total
FROM link_visits;

-- name: ListLinkVisitsRange :many
SELECT id, link_id, created_at, ip, user_agent, status
FROM link_visits
ORDER BY id
    LIMIT $1 OFFSET $2;

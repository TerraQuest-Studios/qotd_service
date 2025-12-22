-- name: GetQuote :one
SELECT q.id, q.quote, t.name AS type, q.inserter, q.inserted
FROM quotes q
JOIN type t ON q.type = t.id
WHERE q.state = 1
ORDER BY RAND()
LIMIT 1;

-- name: TypeExists :one
SELECT COUNT(*) > 0 AS `exists`
FROM type
WHERE name = ?;

-- name: GetRandomQuoteByType :one
SELECT q.id, q.quote, t.name AS type, q.inserter, q.inserted
FROM quotes q
JOIN type t ON q.type = t.id
WHERE q.state = 1 AND t.name = ?
ORDER BY RAND()
LIMIT 1;

-- name: GetLatestQuoteByType :one
SELECT q.id, q.quote, t.name AS type, q.inserter, q.inserted
FROM quotes q
JOIN type t ON q.type = t.id
WHERE q.state = 1 AND t.name = ?
ORDER BY q.inserted DESC
LIMIT 1;

-- name: ActivateOldestQuote :exec
UPDATE quotes
SET state = 1
WHERE id = (
    SELECT id
    FROM quotes
    WHERE state = 0
    ORDER BY inserted ASC
    LIMIT 1
);
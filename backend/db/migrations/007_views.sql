/* VIEW: rating summary per manga */

CREATE VIEW vw_manga_rating AS
SELECT
    m.id            AS manga_id,
    m.title         AS title,
    AVG(r.score)    AS average_rating,
    COUNT(r.id)     AS rating_count
FROM mangas m
LEFT JOIN ratings r ON m.id = r.manga_id
GROUP BY m.id, m.title;

/* VIEW: manga with tags */

CREATE VIEW vw_manga_with_tags AS
SELECT
    m.id                        AS manga_id,
    m.title                     AS title,
    GROUP_CONCAT(t.name, ', ')  AS tags
FROM mangas m
LEFT JOIN manga_tags mt ON m.id = mt.manga_id
LEFT JOIN tags t       ON mt.tag_id = t.id
GROUP BY m.id, m.title;
